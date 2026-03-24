package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// EnvBindings represents the environment bindings injected into the Yaegi interpreter.
// It exposes standard operations like HTTP client and File System access, sandboxed
// to the agent's workspace.
type EnvBindings struct {
	Workspace string
}

// HTTPClient returns a default HTTP client.
func (e *EnvBindings) HTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

// ReadFile reads a file from the workspace. It prevents directory traversal.
func (e *EnvBindings) ReadFile(path string) ([]byte, error) {
	cleanPath := filepath.Clean(path)
	if filepath.IsAbs(cleanPath) || cleanPath == ".." || strings.HasPrefix(cleanPath, "../") || strings.HasPrefix(cleanPath, "..\\") {
		return nil, fmt.Errorf("access denied: cannot read files outside workspace")
	}
	return os.ReadFile(filepath.Join(e.Workspace, cleanPath))
}

// WriteFile writes a file to the workspace. It prevents directory traversal.
func (e *EnvBindings) WriteFile(path string, data []byte) error {
	cleanPath := filepath.Clean(path)
	if filepath.IsAbs(cleanPath) || cleanPath == ".." || strings.HasPrefix(cleanPath, "../") || strings.HasPrefix(cleanPath, "..\\") {
		return fmt.Errorf("access denied: cannot write files outside workspace")
	}
	return os.WriteFile(filepath.Join(e.Workspace, cleanPath), data, 0644)
}

// HTTPGet is a helper to perform an HTTP GET request and return the response body as string.
func (e *EnvBindings) HTTPGet(url string) (string, error) {
	resp, err := e.HTTPClient().Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

type threadSafeBuffer struct {
	b  []byte
	mu sync.Mutex
}

func (b *threadSafeBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.b = append(b.b, p...)
	return len(p), nil
}

func (b *threadSafeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.b)
}

func (b *threadSafeBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.b)
}

type GoEvalTool struct {
	workspace string
	timeout   time.Duration
}

func NewGoEvalTool(workspace string) *GoEvalTool {
	return &GoEvalTool{
		workspace: workspace,
		timeout:   60 * time.Second, // Default timeout
	}
}

func (t *GoEvalTool) Name() string {
	return "go_eval"
}

func (t *GoEvalTool) Description() string {
	return "Executes Go code dynamically using Yaegi interpreter. Provide valid Go source code. The code will be interpreted and executed safely without requiring the Go toolchain. Useful for complex logic or tasks that require writing a Go script. Internal bindings are exposed via the 'jane/env' synthetic package. You can import 'jane/env' and access the injected *EnvBindings object via 'env.Env'. Available methods: Env.HTTPClient() *http.Client, Env.HTTPGet(url string) (string, error), Env.ReadFile(path string) ([]byte, error), Env.WriteFile(path string, data []byte) error. These bindings allow safe, workspace-sandboxed file access and HTTP execution directly inside the script."
}

func (t *GoEvalTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"code": map[string]any{
				"type":        "string",
				"description": "Valid Go source code to execute. It does not need to be a complete package with main func, scripts can just be valid Go statements or functions.",
			},
		},
		"required": []string{"code"},
	}
}

func (t *GoEvalTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	code, ok := args["code"].(string)
	if !ok || code == "" {
		return ErrorResult("code is required")
	}

	cmdCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	var stdout, stderr threadSafeBuffer

	i := interp.New(interp.Options{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{},
	})

	if err := i.Use(stdlib.Symbols); err != nil {
		return &ToolResult{
			ForLLM:  fmt.Sprintf("Failed to initialize standard library symbols: %v", err),
			ForUser: "Execution environment setup failed.",
			IsError: true,
		}
	}

	// Inject internal JANE bindings into the `jane/env` synthetic package.
	// As noted in memory, we provide a fallback path "jane/env/env" to prevent
	// "unable to find source related to" import errors.
	envBindings := &EnvBindings{Workspace: t.workspace}
	bindings := map[string]reflect.Value{
		"Env": reflect.ValueOf(envBindings),
	}
	exports := interp.Exports{
		"jane/env/env": bindings,
		"jane/env":     bindings,
	}
	if err := i.Use(exports); err != nil {
		return &ToolResult{
			ForLLM:  fmt.Sprintf("Failed to initialize jane/env bindings: %v", err),
			ForUser: "Execution environment setup failed.",
			IsError: true,
		}
	}

	// Channel to capture execution result
	done := make(chan error, 1)
	go func() {
		_, err := i.EvalWithContext(cmdCtx, code)
		done <- err
	}()

	var err error
	select {
	case <-cmdCtx.Done():
		err = cmdCtx.Err()
	case err = <-done:
	}

	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\nSTDERR:\n"
		}
		output += stderr.String()
	}

	if err != nil {
		if err == context.DeadlineExceeded {
			return &ToolResult{
				ForLLM:  fmt.Sprintf("Execution timed out after %v.\nOutput so far:\n%s", t.timeout, output),
				ForUser: "Execution timed out.",
				IsError: true,
			}
		}

		return &ToolResult{
			ForLLM:  fmt.Sprintf("Execution failed: %v\nOutput:\n%s", err, output),
			ForUser: "Execution failed.",
			IsError: true,
		}
	}

	if output == "" {
		output = "(no output)"
	}

	// Truncate output if necessary
	maxLen := 10000
	if len(output) > maxLen {
		output = output[:maxLen] + fmt.Sprintf("\n... (truncated, %d more chars)", len(output)-maxLen)
	}

	return &ToolResult{
		ForLLM:  output,
		ForUser: output,
	}
}

func (t *GoEvalTool) RequiresApproval() bool {
	return false
}
