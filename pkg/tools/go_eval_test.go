package tools

import (
	"context"
	"strings"
	"testing"
)

func TestGoEvalTool_Bindings(t *testing.T) {
	tool := NewGoEvalTool(".")
	tool.SetBindings(map[string]any{
		"Workspace": "/tmp",
	})

	args := map[string]any{
		"code": `
package main
import "jane/env"
func main() {
    print(env.Workspace)
}
`,
	}
	res := tool.Execute(context.Background(), args)
	if res.IsError {
		t.Fatalf("Expected no error, got: %s", res.ForLLM)
	}

	if !strings.Contains(res.ForLLM, "/tmp") {
		t.Fatalf("Expected output to contain '/tmp', got: %s", res.ForLLM)
	}
}
