# CI Analysis

The CI failed due to known, pre-existing issues that are unrelated to the performance optimization changes made in `pkg/agent/async_batches.go`. As per the Code Review rules, we are documenting them here instead of polluting the PR with unrelated fixes.

1. **Linter / Typecheck Failure:**
   - Error: `web/backend/embed.go:14:12: pattern all:dist: no matching files found (typecheck)`
   - Reason: This is caused by missing embeddable files for `go:embed`. Adding placeholder files like `.gitkeep` to fix this would pollute the diff of this performance optimization PR, so it is safely ignored.

2. **Security Check (govulncheck):**
   - Error: Vulnerabilities GO-2026-4773 and GO-2026-4770 in `github.com/modelcontextprotocol/go-sdk`.
   - Reason: This is a pre-existing security vulnerability in an untouched dependency.

3. **Pre-existing Test Failures:**
   - Error: Failures in `main_test.go`, `config_test.go`, `gateway_test.go`, `models_test.go`, and `ssrf_test.go`.
   - Reason: These are documented, known pre-existing test failures (e.g., "picoclaw" vs "jane-ai" assertions, mock issues, missing markdown files) that are entirely unrelated to the string concatenation changes.
