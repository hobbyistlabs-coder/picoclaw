1. **Implement `LogSessionEvent` in `pkg/logger/replay.go`**
   - Create a new file `pkg/logger/replay.go`.
   - Implement the observability logging structured around the JSON schema in `AGENT_LOOP_IMPROVEMENTS.md` and memory constraints.
   - Specifically, implement thread-safe append-only file writes using a `sync.Map` of per-session `sync.Mutex` locks to `{workspacePath}/logs/{session_id}/events/events.jsonl`.
   - Do not return errors from `LogSessionEvent` to avoid interrupting the core loop. Handle internal errors silently on best-effort basis.
2. **Integrate `LogSessionEvent` into `pkg/agent/loop_llm.go`**
   - Add calls to `logger.LogSessionEvent` for:
     - `cot` (reasoning text)
     - `tool_call` (tool invocation)
     - `tool_result` (tool response)
     - `error` (LLM/execution failures mapped to `ReplayErrorCategory`)
     - `state_transition` (e.g., pausing for approval)
3. **Integrate `LogSessionEvent` into `pkg/agent/loop_process.go`**
   - Add calls to `logger.LogSessionEvent` for:
     - `state_transition` (e.g., resuming from approval)
4. **Update `AGENT_LOOP_IMPROVEMENTS.md`**
   - Mark "Session Replay" as complete `[x]`.
5. **Complete pre-commit steps to ensure proper testing, verification, review, and reflection are done.**
6. **Submit PR**
