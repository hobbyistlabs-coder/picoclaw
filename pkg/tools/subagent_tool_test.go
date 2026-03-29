package tools

import (
	"context"
	"strings"
	"testing"
	"time"

	"jane/pkg/providers"
)

// MockLLMProvider is a test implementation of LLMProvider
type MockLLMProvider struct {
	lastOptions map[string]any
}

func (m *MockLLMProvider) Chat(
	ctx context.Context,
	messages []providers.Message,
	tools []providers.ToolDefinition,
	model string,
	options map[string]any,
) (*providers.LLMResponse, error) {
	m.lastOptions = options
	// Find the last user message to generate a response
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return &providers.LLMResponse{
				Content: "Task completed: " + messages[i].Content,
			}, nil
		}
	}
	return &providers.LLMResponse{Content: "No task provided"}, nil
}

func (m *MockLLMProvider) GetDefaultModel() string {
	return "test-model"
}

func (m *MockLLMProvider) SupportsTools() bool {
	return false
}

func (m *MockLLMProvider) GetContextWindow() int {
	return 4096
}

type toolCallingProvider struct{}

func (p *toolCallingProvider) Chat(
	ctx context.Context,
	messages []providers.Message,
	tools []providers.ToolDefinition,
	model string,
	options map[string]any,
) (*providers.LLMResponse, error) {
	if len(messages) == 2 {
		return &providers.LLMResponse{
			ToolCalls: []providers.ToolCall{{
				ID:        "call-1",
				Name:      "calculator",
				Arguments: map[string]any{"expression": "1+1"},
			}},
		}, nil
	}
	return &providers.LLMResponse{Content: "done"}, nil
}

func (p *toolCallingProvider) GetDefaultModel() string { return "test-model" }

func TestSubagentManager_SetLLMOptions_AppliesToRunToolLoop(t *testing.T) {
	provider := &MockLLMProvider{}
	manager := NewSubagentManager(provider, "test-model", "/tmp/test")
	manager.SetLLMOptions(2048, 0.6)
	tool := NewSubagentTool(manager)

	ctx := WithToolContext(context.Background(), "cli", "direct")
	args := map[string]any{"task": "Do something"}
	result := tool.Execute(ctx, args)

	if result == nil || result.IsError {
		t.Fatalf("Expected successful result, got: %+v", result)
	}

	if provider.lastOptions == nil {
		t.Fatal("Expected LLM options to be passed, got nil")
	}
	if provider.lastOptions["max_tokens"] != 2048 {
		t.Fatalf("max_tokens = %v, want %d", provider.lastOptions["max_tokens"], 2048)
	}
	if provider.lastOptions["temperature"] != 0.6 {
		t.Fatalf("temperature = %v, want %v", provider.lastOptions["temperature"], 0.6)
	}
}

// TestSubagentTool_Name verifies tool name
func TestSubagentTool_Name(t *testing.T) {
	provider := &MockLLMProvider{}
	manager := NewSubagentManager(provider, "test-model", "/tmp/test")
	tool := NewSubagentTool(manager)

	if tool.Name() != "subagent" {
		t.Errorf("Expected name 'subagent', got '%s'", tool.Name())
	}
}

// TestSubagentTool_Description verifies tool description
func TestSubagentTool_Description(t *testing.T) {
	provider := &MockLLMProvider{}
	manager := NewSubagentManager(provider, "test-model", "/tmp/test")
	tool := NewSubagentTool(manager)

	desc := tool.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
	if !strings.Contains(desc, "subagent") {
		t.Errorf("Description should mention 'subagent', got: %s", desc)
	}
}

// TestSubagentTool_Parameters verifies tool parameters schema
func TestSubagentTool_Parameters(t *testing.T) {
	provider := &MockLLMProvider{}
	manager := NewSubagentManager(provider, "test-model", "/tmp/test")
	tool := NewSubagentTool(manager)

	params := tool.Parameters()
	if params == nil {
		t.Error("Parameters should not be nil")
	}

	// Check type
	if params["type"] != "object" {
		t.Errorf("Expected type 'object', got: %v", params["type"])
	}

	// Check properties
	props, ok := params["properties"].(map[string]any)
	if !ok {
		t.Fatal("Properties should be a map")
	}

	// Verify task parameter
	task, ok := props["task"].(map[string]any)
	if !ok {
		t.Fatal("Task parameter should exist")
	}
	if task["type"] != "string" {
		t.Errorf("Task type should be 'string', got: %v", task["type"])
	}

	// Verify label parameter
	label, ok := props["label"].(map[string]any)
	if !ok {
		t.Fatal("Label parameter should exist")
	}
	if label["type"] != "string" {
		t.Errorf("Label type should be 'string', got: %v", label["type"])
	}

	// Check required fields
	required, ok := params["required"].([]string)
	if !ok {
		t.Fatal("Required should be a string array")
	}
	if len(required) != 1 || required[0] != "task" {
		t.Errorf("Required should be ['task'], got: %v", required)
	}
}

// TestSubagentTool_Execute_Success tests successful execution
func TestSubagentTool_Execute_Success(t *testing.T) {
	provider := &MockLLMProvider{}
	manager := NewSubagentManager(provider, "test-model", "/tmp/test")
	tool := NewSubagentTool(manager)

	ctx := WithToolContext(context.Background(), "telegram", "chat-123")
	args := map[string]any{
		"task":  "Write a haiku about coding",
		"label": "haiku-task",
	}

	result := tool.Execute(ctx, args)

	// Verify basic ToolResult structure
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// Verify no error
	if result.IsError {
		t.Errorf("Expected success, got error: %s", result.ForLLM)
	}

	// Verify not async
	if result.Async {
		t.Error("SubagentTool should be synchronous, not async")
	}

	// Verify not silent
	if result.Silent {
		t.Error("SubagentTool should not be silent")
	}

	// Verify ForUser contains brief summary (not empty)
	if result.ForUser == "" {
		t.Error("ForUser should contain result summary")
	}
	if !strings.Contains(result.ForUser, "Task completed") {
		t.Errorf("ForUser should contain task completion, got: %s", result.ForUser)
	}

	// Verify ForLLM contains full details
	if result.ForLLM == "" {
		t.Error("ForLLM should contain full details")
	}
	if !strings.Contains(result.ForLLM, "haiku-task") {
		t.Errorf("ForLLM should contain label 'haiku-task', got: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "Task completed:") {
		t.Errorf("ForLLM should contain task result, got: %s", result.ForLLM)
	}
}

// TestSubagentTool_Execute_NoLabel tests execution without label
func TestSubagentTool_Execute_NoLabel(t *testing.T) {
	provider := &MockLLMProvider{}
	manager := NewSubagentManager(provider, "test-model", "/tmp/test")
	tool := NewSubagentTool(manager)

	ctx := context.Background()
	args := map[string]any{
		"task": "Test task without label",
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Expected success without label, got error: %s", result.ForLLM)
	}

	// ForLLM should show (unnamed) for missing label
	if !strings.Contains(result.ForLLM, "(unnamed)") {
		t.Errorf("ForLLM should show '(unnamed)' for missing label, got: %s", result.ForLLM)
	}
}

// TestSubagentTool_Execute_MissingTask tests error handling for missing task
func TestSubagentTool_Execute_MissingTask(t *testing.T) {
	provider := &MockLLMProvider{}
	manager := NewSubagentManager(provider, "test-model", "/tmp/test")
	tool := NewSubagentTool(manager)

	ctx := context.Background()
	args := map[string]any{
		"label": "test",
	}

	result := tool.Execute(ctx, args)

	// Should return error
	if !result.IsError {
		t.Error("Expected error for missing task parameter")
	}

	// ForLLM should contain error message
	if !strings.Contains(result.ForLLM, "task is required") {
		t.Errorf("Error message should mention 'task is required', got: %s", result.ForLLM)
	}

	// Err should be set
	if result.Err == nil {
		t.Error("Err should be set for validation failure")
	}
}

// TestSubagentTool_Execute_NilManager tests error handling for nil manager
func TestSubagentTool_Execute_NilManager(t *testing.T) {
	tool := NewSubagentTool(nil)

	ctx := context.Background()
	args := map[string]any{
		"task": "test task",
	}

	result := tool.Execute(ctx, args)

	// Should return error
	if !result.IsError {
		t.Error("Expected error for nil manager")
	}

	if !strings.Contains(result.ForLLM, "Subagent manager not configured") {
		t.Errorf("Error message should mention manager not configured, got: %s", result.ForLLM)
	}
}

// TestSubagentTool_Execute_ContextPassing verifies context is properly used
func TestSubagentTool_Execute_ContextPassing(t *testing.T) {
	provider := &MockLLMProvider{}
	manager := NewSubagentManager(provider, "test-model", "/tmp/test")
	tool := NewSubagentTool(manager)

	channel := "test-channel"
	chatID := "test-chat"
	ctx := WithToolContext(context.Background(), channel, chatID)
	args := map[string]any{
		"task": "Test context passing",
	}

	result := tool.Execute(ctx, args)

	// Should succeed
	if result.IsError {
		t.Errorf("Expected success with context, got error: %s", result.ForLLM)
	}

	// The context is used internally; we can't directly test it
	// but execution success indicates context was handled properly
}

// TestSubagentTool_ForUserTruncation verifies long content is truncated for user
func TestSubagentTool_ForUserTruncation(t *testing.T) {
	// Create a mock provider that returns very long content
	provider := &MockLLMProvider{}
	manager := NewSubagentManager(provider, "test-model", "/tmp/test")
	tool := NewSubagentTool(manager)

	ctx := context.Background()

	// Create a task that will generate long response
	longTask := strings.Repeat("This is a very long task description. ", 100)
	args := map[string]any{
		"task":  longTask,
		"label": "long-test",
	}

	result := tool.Execute(ctx, args)

	// ForUser should be truncated to 500 chars + "..."
	maxUserLen := 500
	if len(result.ForUser) > maxUserLen+3 { // +3 for "..."
		t.Errorf("ForUser should be truncated to ~%d chars, got: %d", maxUserLen, len(result.ForUser))
	}

	// ForLLM should have full content
	if !strings.Contains(result.ForLLM, longTask[:50]) {
		t.Error("ForLLM should contain reference to original task")
	}
}

func TestSubagentManager_TracksProgressAndEvents(t *testing.T) {
	provider := &toolCallingProvider{}
	manager := NewSubagentManager(provider, "test-model", "/tmp/test")
	manager.RegisterTool(NewCalculatorTool())
	var seen []SubagentProgressEvent
	ctx := WithToolSessionKey(context.Background(), "agent:main:session-1")
	ctx = WithToolCallID(ctx, "tool-call-1")
	ctx = WithToolAsyncBatchID(ctx, "batch-1")
	ctx = WithSubagentProgress(ctx, func(task *SubagentTask, event *SubagentProgressEvent) {
		seen = append(seen, *event)
	})

	message, err := manager.Spawn(ctx, "calculate a result", "math", "", "pico", "chat-1", "agent:main:session-1", nil)
	if err != nil {
		t.Fatalf("Spawn returned error: %v", err)
	}
	if !strings.Contains(message, "Petra") {
		t.Fatalf("expected generated codename in spawn message, got %q", message)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		tasks := manager.ListTasks()
		if len(tasks) == 1 && tasks[0].Status == SubagentCompleted {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	tasks := manager.ListTasks()
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	task := tasks[0]
	if task.ParentSessionID != "agent:main:session-1" {
		t.Fatalf("parent session = %q", task.ParentSessionID)
	}
	if task.BatchID != "batch-1" {
		t.Fatalf("batch id = %q", task.BatchID)
	}
	if task.ParentToolCallID != "tool-call-1" {
		t.Fatalf("parent tool call id = %q", task.ParentToolCallID)
	}
	if task.Codename != "Petra" {
		t.Fatalf("codename = %q", task.Codename)
	}

	events := manager.GetSubagentEvents(task.ID, 0)
	if len(events) < 4 {
		t.Fatalf("expected multiple progress events, got %d", len(events))
	}
	hasToolStart := false
	hasCompletion := false
	for _, event := range events {
		if event.EventType == "tool.started" {
			hasToolStart = true
		}
		if event.EventType == "task.completed" {
			hasCompletion = true
		}
	}
	if !hasToolStart {
		t.Fatal("expected tool.started event")
	}
	if !hasCompletion {
		t.Fatal("expected task.completed event")
	}
	if len(seen) == 0 {
		t.Fatal("expected progress callback events")
	}

	batch := manager.GetBatchStatus("batch-1")
	if batch == nil || batch.Completed != 1 {
		t.Fatalf("unexpected batch status: %+v", batch)
	}
}

func TestSubagentManager_GetStalledSubagents(t *testing.T) {
	manager := NewSubagentManager(&MockLLMProvider{}, "test-model", "/tmp/test")
	manager.tasks["subagent-9"] = &SubagentTask{
		ID:              "subagent-9",
		Codename:        "Iris",
		ParentSessionID: "agent:main:session-2",
		Status:          SubagentRunning,
		Created:         time.Now().Add(-2 * time.Minute).UnixMilli(),
		Updated:         time.Now().Add(-2 * time.Minute).UnixMilli(),
	}

	stalled := manager.GetStalledSubagents(30 * time.Second)
	if len(stalled) != 1 {
		t.Fatalf("expected 1 stalled task, got %d", len(stalled))
	}
	active := manager.ListActiveSubagents("agent:main:session-2")
	if len(active) != 1 {
		t.Fatalf("expected 1 active task, got %d", len(active))
	}
}
