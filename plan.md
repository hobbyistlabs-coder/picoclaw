1. **Implement `pkg/logger/replay.go`**
   - Provide a `SessionEvent` structure representing a structured JSON log.
   - Provide `LogSessionEvent` to append events to `{workspace}/logs/{session_id}/events/events.jsonl`.
   - Add required constants for error categories.

2. **Update `pkg/agent/loop_llm.go` to capture Chain of Thought (CoT), tool calls, and transitions**
   - After the LLM response is received, log the reasoning/CoT text.
   - When tools are called, log a tool_call event.
   - After tools are executed, log a tool_result event.

3. **Update `pkg/agent/loop_process.go` to capture state transitions**
   - Log transitions when moving into or out of the human-in-the-loop approval state.

4. **Complete pre-commit steps to ensure proper testing, verification, review, and reflection are done.**
