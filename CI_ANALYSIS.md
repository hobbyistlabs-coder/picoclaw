# CI Failure Analysis

The `golangci-lint` job is failing with 289 issues, mostly complaining about `gofumpt` and `golines` across the entire codebase.

Per the Code Review Rule, I should not run a global formatter (`gofumpt -w .`) over 200+ unrelated files, as this pollutes the pull request diff.

I have addressed the top 10 explicitly annotated linter errors (the `canonicalheader` issues in `cmd/picoclaw/internal/jules`, and the `gci` / `gofumpt` issues in `pkg/tools/alpaca/alpaca.go`, `pkg/health/resource_tracker.go`, `pkg/channels/gmessages/messages.go`, etc.).

The `bodyclose` issue in `web/backend/api/boards_trigger.go` was already fixed in the previous commit.

The remaining issues are pre-existing technical debt across the repo and can be safely ignored to keep the PR focused on the ETL Visibility tasks.

The Security Check failed because `go mod tidy` changes from the first fix were lost in the local environment due to a reset. I have re-run `go get github.com/modelcontextprotocol/go-sdk@v1.4.1` and checked in `go.mod` and `go.sum` to properly apply the `GO-2026-4773` and `GO-2026-4770` vulnerability fixes.

The Tests failed because the `dist` placeholder files (`web/backend/dist/keep.txt` and `cmd/picoclaw/internal/onboard/workspace/AGENTS.md`) were lost in the environment reset. I've re-created and explicitly force-added them.
