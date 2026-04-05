package tools

import (
	"context"
	"fmt"
	"sync"
	"time"

	"jane/pkg/providers"
	"jane/pkg/utils"
)

// AgentDispatcher executes a task on a specific agent instance
type AgentDispatcher interface {
	DispatchSubagent(
		ctx context.Context,
		agentID, task, originChannel, originChatID, sessionKey string,
	) (*ToolResult, error)
}

type SubagentTask struct {
	ID                string
	Task              string
	Label             string
	Codename          string
	AgentID           string
	ParentToolCallID  string
	BatchID           string
	OriginChannel     string
	OriginChatID      string
	OriginSession     string
	ParentSessionID   string
	Status            string
	Result            string
	Summary           string
	LatestEvent       string
	LastOutputExcerpt string
	Error             string
	ToolName          string
	ToolStatus        string
	ProgressPercent   int
	Created           int64
	Started           int64
	Updated           int64
}

type SubagentManager struct {
	tasks             map[string]*SubagentTask
	events            map[string][]SubagentProgressEvent
	progressCallbacks map[string]SubagentProgressCallback
	mu                sync.RWMutex
	provider          providers.LLMProvider
	defaultModel      string
	workspace         string
	tools             *ToolRegistry
	maxIterations     int
	maxTokens         int
	temperature       float64
	hasMaxTokens      bool
	hasTemperature    bool
	nextID            int
	dispatcher        AgentDispatcher
}

func NewSubagentManager(
	provider providers.LLMProvider,
	defaultModel, workspace string,
) *SubagentManager {
	return &SubagentManager{
		tasks:             make(map[string]*SubagentTask),
		events:            make(map[string][]SubagentProgressEvent),
		progressCallbacks: make(map[string]SubagentProgressCallback),
		provider:          provider,
		defaultModel:      defaultModel,
		workspace:         workspace,
		tools:             NewToolRegistry(),
		maxIterations:     10,
		nextID:            1,
	}
}

// SetLLMOptions sets max tokens and temperature for subagent LLM calls.
func (sm *SubagentManager) SetLLMOptions(maxTokens int, temperature float64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.maxTokens = maxTokens
	sm.hasMaxTokens = true
	sm.temperature = temperature
	sm.hasTemperature = true
}

// SetDispatcher sets the agent dispatcher for multi-agent delegation.
func (sm *SubagentManager) SetDispatcher(dispatcher AgentDispatcher) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.dispatcher = dispatcher
}

// SetTools sets the tool registry for subagent execution.
func (sm *SubagentManager) SetTools(tools *ToolRegistry) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.tools = tools
}

// RegisterTool registers a tool for subagent execution.
func (sm *SubagentManager) RegisterTool(tool Tool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.tools.Register(tool)
}

func (sm *SubagentManager) Spawn(
	ctx context.Context,
	task, label, agentID, originChannel, originChatID, originSession string,
	callback AsyncCallback,
) (string, error) {
	sm.mu.Lock()
	index := sm.nextID
	taskID := fmt.Sprintf("subagent-%d", index)
	sm.nextID++
	now := time.Now().UnixMilli()
	batchID := ToolAsyncBatchID(ctx)
	if batchID == "" {
		batchID = originSession
	}
	subagentTask := &SubagentTask{
		ID:               taskID,
		Task:             task,
		Label:            label,
		Codename:         nextCodename(index),
		AgentID:          agentID,
		ParentToolCallID: ToolCallID(ctx),
		BatchID:          batchID,
		OriginChannel:    originChannel,
		OriginChatID:     originChatID,
		OriginSession:    originSession,
		ParentSessionID:  originSession,
		Status:           SubagentQueued,
		Summary:          "Queued for execution",
		LatestEvent:      "task.queued",
		Created:          now,
		Updated:          now,
	}
	sm.tasks[taskID] = subagentTask
	if cb := ToolSubagentProgress(ctx); cb != nil {
		sm.progressCallbacks[taskID] = cb
	}
	sm.mu.Unlock()

	go sm.runTask(ctx, taskID, callback)

	if label != "" {
		return fmt.Sprintf("Spawned subagent %s (%s) for task: %s",
			subagentTask.Codename, label, task), nil
	}
	return fmt.Sprintf("Spawned subagent %s for task: %s", subagentTask.Codename, task), nil
}

func (sm *SubagentManager) runTask(ctx context.Context, taskID string, callback AsyncCallback) {
	sm.emit(taskID, "task.started", SubagentRunning, "Started delegated task", nil)
	sm.emit(taskID, "task.progress", SubagentRunning, "Building delegated prompt", nil)

	task, ok := sm.GetTask(taskID)
	if !ok {
		return
	}
	systemPrompt := `You are a subagent. Complete the given task independently and report the result.
You have access to tools - use them as needed to complete your task.
After completing the task, provide a clear summary of what was done.`
	messages := []providers.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: task.Task},
	}

	select {
	case <-ctx.Done():
		sm.emit(taskID, "task.failed", SubagentCanceled, "Task canceled before execution",
			map[string]any{"error": "context canceled"})
		return
	default:
	}

	var result *ToolResult

	sm.mu.RLock()
	dispatcher := sm.dispatcher
	sm.mu.RUnlock()

	if dispatcher != nil && task.AgentID != "" {
		var err error
		result, err = dispatcher.DispatchSubagent(
			ctx,
			task.AgentID,
			task.Task,
			task.OriginChannel,
			task.OriginChatID,
			task.ParentSessionID,
		)
		if err != nil {
			status := SubagentFailed
			message := fmt.Sprintf("Delegated task failed: %v", err)
			if ctx.Err() != nil {
				status = SubagentCanceled
				message = "Task canceled during execution"
			}
			sm.emit(taskID, "task.failed", status, message, map[string]any{"error": err.Error()})
			result = ErrorResult(message).WithError(err)
		} else {
			sm.emit(taskID, "task.completed", SubagentCompleted, "Delegated task completed",
				map[string]any{
					"result":           utils.Truncate(result.ForLLM, 240),
					"progress_percent": 100,
				})
			result.ForLLM = fmt.Sprintf("Subagent '%s' (Agent %s) completed:\n%s",
				task.Codename, task.AgentID, result.ForLLM)
		}
	} else {
		sm.mu.RLock()
		tools := sm.tools
		maxIter := sm.maxIterations
		maxTokens := sm.maxTokens
		temperature := sm.temperature
		hasMaxTokens := sm.hasMaxTokens
		hasTemperature := sm.hasTemperature
		sm.mu.RUnlock()

		var llmOptions map[string]any
		if hasMaxTokens || hasTemperature {
			llmOptions = map[string]any{}
			if hasMaxTokens {
				llmOptions["max_tokens"] = maxTokens
			}
			if hasTemperature {
				llmOptions["temperature"] = temperature
			}
		}

		loopResult, err := RunToolLoop(ctx, ToolLoopConfig{
			Provider:      sm.provider,
			Model:         sm.defaultModel,
			Tools:         tools,
			MaxIterations: maxIter,
			LLMOptions:    llmOptions,
			OnToolCall: func(tc providers.ToolCall, status string, toolResult *ToolResult) {
				sm.recordToolEvent(taskID, tc, status, toolResult)
			},
		}, messages, task.OriginChannel, task.OriginChatID)

		if err != nil {
			status := SubagentFailed
			message := fmt.Sprintf("Delegated task failed: %v", err)
			if ctx.Err() != nil {
				status = SubagentCanceled
				message = "Task canceled during execution"
			}
			sm.emit(taskID, "task.failed", status, message, map[string]any{"error": err.Error()})
			result = ErrorResult(message).WithError(err)
		} else {
			sm.emit(taskID, "task.completed", SubagentCompleted, "Delegated task completed",
				map[string]any{
					"result":           utils.Truncate(loopResult.Content, 240),
					"progress_percent": 100,
				})
			result = &ToolResult{
				ForLLM: fmt.Sprintf("Subagent '%s' completed (iterations: %d): %s",
					task.Codename, loopResult.Iterations, loopResult.Content),
				ForUser: loopResult.Content,
				Silent:  false,
				IsError: false,
				Async:   false,
			}
		}
	}

	if callback != nil && result != nil {
		callback(ctx, result)
	}
}

func (sm *SubagentManager) GetTask(taskID string) (*SubagentTask, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	task, ok := sm.tasks[taskID]
	if !ok {
		return nil, false
	}
	clone := *task
	return &clone, true
}

func (sm *SubagentManager) ListTasks() []*SubagentTask {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	tasks := make([]*SubagentTask, 0, len(sm.tasks))
	for _, task := range sm.tasks {
		clone := *task
		tasks = append(tasks, &clone)
	}
	return tasks
}

func (sm *SubagentManager) ListActiveSubagents(sessionID string) []*SubagentTask {
	tasks := sm.ListTasks()
	out := make([]*SubagentTask, 0, len(tasks))
	for _, task := range tasks {
		if sessionID != "" && task.ParentSessionID != sessionID {
			continue
		}
		switch task.Status {
		case SubagentQueued, SubagentRunning, SubagentWaitingForTool, SubagentBlocked:
			out = append(out, task)
		}
	}
	return out
}

func (sm *SubagentManager) GetSubagentStatus(taskID string) (*SubagentTask, bool) {
	return sm.GetTask(taskID)
}

func (sm *SubagentManager) GetSubagentEvents(taskID string, limit int) []SubagentProgressEvent {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	events := sm.events[taskID]
	if limit <= 0 || limit >= len(events) {
		out := make([]SubagentProgressEvent, len(events))
		copy(out, events)
		return out
	}
	out := make([]SubagentProgressEvent, limit)
	copy(out, events[len(events)-limit:])
	return out
}

func (sm *SubagentManager) GetBatchStatus(batchID string) *SubagentBatchStatus {
	if batchID == "" {
		return nil
	}
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	status := &SubagentBatchStatus{BatchID: batchID}
	for _, task := range sm.tasks {
		if task.BatchID != batchID {
			continue
		}
		status.Total++
		if task.Updated > status.LatestUpdate {
			status.LatestUpdate = task.Updated
			status.Summary = task.Summary
		}
		switch task.Status {
		case SubagentQueued, SubagentRunning, SubagentWaitingForTool:
			status.Running++
		case SubagentBlocked:
			status.Blocked++
		case SubagentFailed:
			status.Failed++
		case SubagentCompleted:
			status.Completed++
		case SubagentCanceled:
			status.Canceled++
		}
	}
	if status.Total == 0 {
		return nil
	}
	return status
}

func (sm *SubagentManager) GetStalledSubagents(threshold time.Duration) []*SubagentTask {
	if threshold <= 0 {
		return nil
	}
	now := time.Now().UnixMilli()
	tasks := sm.ListTasks()
	out := make([]*SubagentTask, 0, len(tasks))
	for _, task := range tasks {
		if task.Status != SubagentQueued &&
			task.Status != SubagentRunning &&
			task.Status != SubagentWaitingForTool {
			continue
		}
		if now-task.Updated >= threshold.Milliseconds() {
			out = append(out, task)
		}
	}
	return out
}

func (sm *SubagentManager) recordToolEvent(
	taskID string,
	tc providers.ToolCall,
	status string,
	result *ToolResult,
) {
	message := fmt.Sprintf("Using tool %s", tc.Name)
	taskStatus := SubagentWaitingForTool
	meta := map[string]any{"tool_name": tc.Name, "tool_status": status}
	if status == "completed" {
		taskStatus = SubagentRunning
		message = fmt.Sprintf("Completed tool %s", tc.Name)
		if result != nil {
			meta["result"] = utils.Truncate(toolResultSnippet(result), 180)
		}
	}
	sm.emit(taskID, "tool."+status, taskStatus, message, meta)
}

func (sm *SubagentManager) emit(
	taskID string,
	eventType string,
	status string,
	message string,
	meta map[string]any,
) {
	sm.mu.Lock()
	task, ok := sm.tasks[taskID]
	if !ok {
		sm.mu.Unlock()
		return
	}
	now := time.Now().UnixMilli()
	if task.Started == 0 && status != SubagentQueued {
		task.Started = now
	}
	task.Updated = now
	task.Status = status
	task.LatestEvent = eventType
	if message != "" {
		task.Summary = message
	}
	if meta != nil {
		if toolName, _ := meta["tool_name"].(string); toolName != "" {
			task.ToolName = toolName
		}
		if toolStatus, _ := meta["tool_status"].(string); toolStatus != "" {
			task.ToolStatus = toolStatus
		}
		if result, _ := meta["result"].(string); result != "" {
			task.LastOutputExcerpt = result
			task.Result = result
		}
		if errMsg, _ := meta["error"].(string); errMsg != "" {
			task.Error = errMsg
		}
		if progress, ok := meta["progress_percent"].(int); ok {
			task.ProgressPercent = progress
		}
	}
	event := SubagentProgressEvent{
		TaskID:     task.ID,
		Codename:   task.Codename,
		Timestamp:  now,
		EventType:  eventType,
		Status:     task.Status,
		Message:    message,
		ToolName:   task.ToolName,
		ToolStatus: task.ToolStatus,
		Error:      task.Error,
		Metadata:   meta,
	}
	sm.events[taskID] = append(sm.events[taskID], event)
	snapshot := *task
	cb := sm.progressCallbacks[taskID]
	if isTerminalStatus(task.Status) {
		delete(sm.progressCallbacks, taskID)
	}
	sm.mu.Unlock()
	if cb != nil {
		cb(&snapshot, &event)
	}
}

func toolResultSnippet(result *ToolResult) string {
	if result == nil {
		return ""
	}
	if result.ForUser != "" {
		return result.ForUser
	}
	if result.ForLLM != "" {
		return result.ForLLM
	}
	if result.Err != nil {
		return result.Err.Error()
	}
	return ""
}

func isTerminalStatus(status string) bool {
	return status == SubagentCompleted ||
		status == SubagentFailed ||
		status == SubagentCanceled
}

func nextCodename(index int) string {
	names := []string{"Petra", "Bean", "Shen", "Dink", "Mica", "Rune", "Vale", "Iris"}
	base := names[(index-1)%len(names)]
	round := (index-1)/len(names) + 1
	if round == 1 {
		return base
	}
	return fmt.Sprintf("%s-%d", base, round)
}

// SubagentTool executes a subagent task synchronously and returns the result.
type SubagentTool struct {
	manager *SubagentManager
}

func NewSubagentTool(manager *SubagentManager) *SubagentTool {
	return &SubagentTool{manager: manager}
}

func (t *SubagentTool) Name() string {
	return "subagent"
}

func (t *SubagentTool) Description() string {
	return "Execute a subagent task synchronously and return the result. Use this for delegating specific tasks to an independent agent instance. Returns execution summary to user and full details to LLM."
}

func (t *SubagentTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"task": map[string]any{
				"type":        "string",
				"description": "The task for subagent to complete",
			},
			"label": map[string]any{
				"type":        "string",
				"description": "Optional short label for the task (for display)",
			},
		},
		"required": []string{"task"},
	}
}

func (t *SubagentTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	task, ok := args["task"].(string)
	if !ok {
		return ErrorResult("task is required").WithError(fmt.Errorf("task parameter is required"))
	}
	label, _ := args["label"].(string)
	if t.manager == nil {
		return ErrorResult(
			"Subagent manager not configured",
		).WithError(fmt.Errorf("manager is nil"))
	}
	messages := []providers.Message{
		{
			Role:    "system",
			Content: "You are a subagent. Complete the given task independently and provide a clear, concise result.",
		},
		{Role: "user", Content: task},
	}

	sm := t.manager
	sm.mu.RLock()
	tools := sm.tools
	maxIter := sm.maxIterations
	maxTokens := sm.maxTokens
	temperature := sm.temperature
	hasMaxTokens := sm.hasMaxTokens
	hasTemperature := sm.hasTemperature
	sm.mu.RUnlock()

	var llmOptions map[string]any
	if hasMaxTokens || hasTemperature {
		llmOptions = map[string]any{}
		if hasMaxTokens {
			llmOptions["max_tokens"] = maxTokens
		}
		if hasTemperature {
			llmOptions["temperature"] = temperature
		}
	}

	channel := ToolChannel(ctx)
	if channel == "" {
		channel = "cli"
	}
	chatID := ToolChatID(ctx)
	if chatID == "" {
		chatID = "direct"
	}

	loopResult, err := RunToolLoop(ctx, ToolLoopConfig{
		Provider:      sm.provider,
		Model:         sm.defaultModel,
		Tools:         tools,
		MaxIterations: maxIter,
		LLMOptions:    llmOptions,
	}, messages, channel, chatID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Subagent execution failed: %v", err)).WithError(err)
	}

	userContent := loopResult.Content
	if len(userContent) > 500 {
		userContent = userContent[:500] + "..."
	}
	labelStr := label
	if labelStr == "" {
		labelStr = "(unnamed)"
	}
	llmContent := fmt.Sprintf("Subagent task completed:\nLabel: %s\nIterations: %d\nResult: %s",
		labelStr, loopResult.Iterations, loopResult.Content)
	return &ToolResult{
		ForLLM:  llmContent,
		ForUser: userContent,
		Silent:  false,
		IsError: false,
		Async:   false,
	}
}

func (t *SubagentTool) RequiresApproval() bool {
	return false
}
