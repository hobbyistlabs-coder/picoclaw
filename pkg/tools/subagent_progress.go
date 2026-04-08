package tools

type SubagentProgressCallback func(task *SubagentTask, event *SubagentProgressEvent)

const (
	SubagentQueued         = "queued"
	SubagentRunning        = "running"
	SubagentWaitingForTool = "waiting_for_tool"
	SubagentBlocked        = "blocked"
	SubagentCompleted      = "completed"
	SubagentFailed         = "failed"
	SubagentCanceled      = "canceled"
)

type SubagentProgressEvent struct {
	TaskID     string
	Codename   string
	Timestamp  int64
	EventType  string
	Status     string
	Message    string
	ToolName   string
	ToolStatus string
	Error      string
	Metadata   map[string]any
}

type SubagentBatchStatus struct {
	BatchID      string
	Total        int
	Running      int
	Blocked      int
	Failed       int
	Completed    int
	Canceled    int
	LatestUpdate int64
	Summary      string
}
