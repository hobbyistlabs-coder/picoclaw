package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ReplayErrorCategory string

const (
	ModelFailure          ReplayErrorCategory = "model_failure"
	InfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	LogicFailure          ReplayErrorCategory = "logic_failure"
	NoneFailure           ReplayErrorCategory = "none"
)

type sessionEventDetails struct {
	CotText   string         `json:"cot_text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Inputs    map[string]any `json:"inputs,omitempty"`
	Outputs   map[string]any `json:"outputs,omitempty"`
	FromState string         `json:"from_state,omitempty"`
	ToState   string         `json:"to_state,omitempty"`
}

type sessionEvent struct {
	Timestamp     time.Time            `json:"timestamp"`
	SessionID     string               `json:"session_id"`
	EventType     string               `json:"event_type"` // cot, tool_call, tool_result, state_transition, error
	Details       *sessionEventDetails `json:"details,omitempty"`
	ErrorCategory string               `json:"error_category,omitempty"`
	ErrorMessage  string               `json:"error_message,omitempty"`
}

var (
	sessionMutexes sync.Map
)

func getSessionMutex(sessionID string) *sync.Mutex {
	mtx, _ := sessionMutexes.LoadOrStore(sessionID, &sync.Mutex{})
	return mtx.(*sync.Mutex)
}

func LogSessionEvent(workspacePath, sessionID, eventType string, details map[string]any, errCat ReplayErrorCategory, errMsg string) {
	// Best-effort execution
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Suppress panics
			}
		}()

		mtx := getSessionMutex(sessionID)
		mtx.Lock()
		defer mtx.Unlock()

		logDir := filepath.Join(workspacePath, "logs", sessionID, "events")
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			return
		}

		logFile := filepath.Join(logDir, "events.jsonl")
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return
		}
		defer f.Close()

		var det *sessionEventDetails
		if details != nil {
			det = &sessionEventDetails{}
			if val, ok := details["cot_text"].(string); ok {
				det.CotText = val
			}
			if val, ok := details["tool_name"].(string); ok {
				det.ToolName = val
			}
			if val, ok := details["inputs"].(map[string]any); ok {
				det.Inputs = val
			}
			if val, ok := details["outputs"].(map[string]any); ok {
				det.Outputs = val
			}
			if val, ok := details["from_state"].(string); ok {
				det.FromState = val
			}
			if val, ok := details["to_state"].(string); ok {
				det.ToState = val
			}
		}

		event := sessionEvent{
			Timestamp:     time.Now().UTC(),
			SessionID:     sessionID,
			EventType:     eventType,
			Details:       det,
			ErrorCategory: string(errCat),
			ErrorMessage:  errMsg,
		}

		if string(errCat) == "" {
			event.ErrorCategory = string(NoneFailure)
		}

		data, err := json.Marshal(event)
		if err != nil {
			return
		}

		_, _ = f.Write(append(data, '\n'))
	}()
}
