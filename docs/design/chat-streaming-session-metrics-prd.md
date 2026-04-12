# PRD: Live Chat Streaming And Session Cost Metrics

## Summary
Fix two chat UX failures in the web UI:
1. Assistant text only appears as a final message instead of streaming live.
2. Session metrics, especially `Est. cost`, stay `n/a` during live chats and only appear after reopening via history.

## Problem
- Users cannot see incremental generation progress in chat.
- Live sessions show stale or missing metrics, which breaks trust in model pricing.
- Current behavior makes model-card pricing look disconnected from runtime behavior.

## User Impact
- Users think the app is frozen while the model is generating.
- Cost visibility is unreliable even when the model card has valid pricing.
- Session history behaves differently from the active session, which feels like a bug.

## Goals
- Stream assistant text and reasoning into the active chat as chunks arrive.
- Update session metrics on every live turn without requiring a history reload.
- Use model pricing from the configured model card as the source of truth.
- Support separate input and output token pricing when available.

## Non-Goals
- Do not fix browser extension errors like `bootstrap-autofill-overlay.js`.
- Do not redesign the chat UI beyond what is needed for live streaming and metrics.
- Do not backfill old sessions unless explicitly requested later.

## Current Findings
- The agent loop already enables streaming and publishes `OutboundStreamMessage`.
- The Pico UI path now has a `message.stream` event, but live behavior still needs end-to-end verification against provider output.
- Session cost aggregation was improved, but active-session metrics still differ from history-loaded metrics.
- Screenshot evidence shows the session metrics bar rendering while the gateway is offline, so state transitions and cached metrics need review.

## Hypotheses
- Providers may not emit stream callbacks for the active model/provider pair.
- The Pico transport may emit chunks, but the frontend may replace the pending message with the final payload before chunks become visible.
- Live metrics may only be attached to final `message.create` payloads, while the active-session rollup ignores pending/live state transitions.
- The active session may not recompute metrics when the selected model or gateway state changes.

## Requirements
- Add traceable logs for stream callback emission, Pico stream delivery, and frontend chunk receipt.
- Verify streaming on the active `miniMax` OpenRouter path and at least one known-streaming provider.
- Make active-session metrics derive from the same source of truth as history-loaded sessions.
- Show `Est. cost` as soon as a live turn has token usage and model pricing.
- If only final usage is available, update cost immediately on final turn completion without requiring session reload.

## Proposed Plan
- Phase 1: Instrument provider, bus, Pico channel, and frontend chat hook to confirm where live chunks stop.
- Phase 2: Normalize live message lifecycle so pending, streamed, and final assistant states merge into one message.
- Phase 3: Refactor active-session metric calculation to recompute from current live messages plus history baseline on every relevant event.
- Phase 4: Add regression tests for streaming visibility and live metric updates.

## Acceptance Criteria
- In a new chat, assistant text visibly grows before the final response completes.
- In a new chat, token and cost metrics update in the session metrics bar on turn completion without reopening history.
- With configured split pricing, cost reflects prompt and completion token rates separately.
- With only flat `price_per_m_token`, cost still appears for new sessions.
- Reopening the same session from history shows the same totals as the live session view.

## Test Plan
- Provider-level test: stream callback fires and publishes multiple chunks.
- Channel-level test: Pico emits ordered `message.stream` then final `message.create`.
- Frontend test: pending assistant message appends chunks and preserves final metrics.
- Integration test: new session shows non-`n/a` cost after a completed turn with priced model config.
- Manual Docker test: verify in `docker/docker-compose.full.yml` with `miniMax`.

## Risks
- Some providers may not support token streaming even if they support final responses.
- Final usage may arrive only once, requiring careful merge logic to avoid double counting.
- Gateway disconnects or offline state can mask streaming failures during manual verification.
