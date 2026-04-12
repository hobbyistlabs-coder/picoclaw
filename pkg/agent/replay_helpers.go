package agent

import (
	"jane/pkg/logger"
	"jane/pkg/tools"
)

func logSessionStateTransition(workspace, sessionID, fromState, toState string) {
	logger.LogSessionEvent(workspace, logger.SessionEvent{
		SessionID: sessionID,
		EventType: logger.EventTypeStateTransition,
		Details: logger.SessionEventDetails{
			FromState: fromState,
			ToState:   toState,
		},
	})
}

func logSessionToolResult(workspace, sessionID, toolName string, result *tools.ToolResult) {
	errCat := logger.ReplayErrorCategoryNone
	var errMsg string
	if result.IsError {
		errCat = logger.ReplayErrorCategoryLogicFailure
		if result.Err != nil {
			errMsg = result.Err.Error()
		} else {
			errMsg = result.ForLLM
		}
	}
	logger.LogSessionEvent(workspace, logger.SessionEvent{
		SessionID:     sessionID,
		EventType:     logger.EventTypeToolResult,
		ErrorCategory: errCat,
		ErrorMessage:  errMsg,
		Details: logger.SessionEventDetails{
			ToolName: toolName,
			Outputs: map[string]any{
				"for_llm":  result.ForLLM,
				"for_user": result.ForUser,
			},
		},
	})
}
