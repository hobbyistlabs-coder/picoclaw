1.  **Implement Session Replay Logging**:
    *   Create `pkg/logger/replay.go` to implement structured JSON logging for session replays.
    *   Define the JSON schema structures matching the requirements in `AGENT_LOOP_IMPROVEMENTS.md` (e.g., `SessionEvent`, `ReplayErrorCategory`, etc.).
    *   Implement `LogSessionEvent` utilizing a `sync.Map` to manage file handles (`sessionFileState` with `sync.Mutex` and `*os.File`) to safely append JSONL events to `{workspacePath}/logs/{session_id}/events/events.jsonl`.
    *   Implement `CleanupSessionLocks` to clean up and close file handles when a session is complete.

2.  **Integrate Session Replay Logging into the Agent Loop**:
    *   Modify `pkg/agent/loop_llm.go` (and other related files if needed) to call `LogSessionEvent` at key points:
        *   When CoT (Reasoning Content) is generated (`event_type: "cot"`).
        *   When a tool call is requested (`event_type: "tool_call"`).
        *   When a tool execution completes (`event_type: "tool_result"`).
        *   When an error occurs (`event_type: "error"`).
    *   Ensure that `logger.CleanupSessionLocks(sessionID)` is called appropriately (e.g., via `defer` in `runAgentLoop`) to prevent memory/descriptor leaks.
    *   Ensure any file path manipulation uses safe practices as mandated by the memory (e.g., `filepath.Clean()`, flattening `.` and `/` using `strings.ReplaceAll` for composite keys).

3.  **Update AGENT_LOOP_IMPROVEMENTS.md**:
    *   Mark `**Session Replay:**` as complete under Phase 5.

4.  **Complete pre commit steps**:
    *   Complete pre-commit steps to ensure proper testing, verification, review, and reflection are done.

5.  **Submit the change**:
    *   Use the `submit` tool to push the branch.
