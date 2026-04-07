# CI Failure Analysis

1.  **Linter (`golangci-lint`)**: The linter fails with exactly 291 issues across various unmodified parts of the codebase. A significant portion of these are style checks (`gofumpt`, `golines`, `gci`, `goimports`) on files I have not touched in this PR (e.g., `pkg/channels/...`, `web/backend/api/...`). I also see warnings about misspelling (`canceled` vs `cancelled`), `nakedret`, `govet` shadow declarations, etc., again in untouched areas.
2.  **Tests**: The test check run `FAIL	jane/web/backend/api	0.220s` reports exactly the 2 pre-existing failures (`TestFindPicoclawBinary_EnvOverride_InvalidPath` and `TestHandleListModels_ConfiguredStatusForOAuthModelWithCredential`) which were documented in memory as known, safe to ignore test failures since they are unrelated to my modifications.
3.  **Code Review Rule context**: Per project instructions, "When addressing CI failures, generally do not fix preexisting issues (e.g., failing tests, linter warnings, or security vulnerabilities) in files unrelated to your current task. Over-reaching to fix unrelated codebase issues will pollute the pull request. Only fix CI errors introduced by your specific changes, unless the automated CI check system explicitly demands a fix for annotated issues after you have attempted to document and ignore them."

However, I notice there *are* 10 specific annotations at the top of the linter job output:
  [FAILURE] File: pkg/etl/pipeline.go, Line: 25
    Message: File is not properly formatted (gci)
  [FAILURE] File: pkg/channels/gmessages/messages.go, Line: 10
    Message: File is not properly formatted (gci)
  [FAILURE] File: pkg/channels/gmessages/events.go, Line: 10
    Message: File is not properly formatted (gci)
  [FAILURE] File: pkg/channels/gmessages/client.go, Line: 13
    Message: File is not properly formatted (gci)
  [FAILURE] File: cmd/picoclaw/internal/jules/jules_test.go, Line: 9
    Message: File is not properly formatted (gci)
  [FAILURE] File: cmd/picoclaw-launcher-tui/internal/ui/style.go, Line: 9
    Message: File is not properly formatted (gci)
  [FAILURE] File: cmd/picoclaw/internal/jules/jules_test.go, Line: 80
    Message: non-canonical header "x-goog-api-key", instead use: "X-Goog-Api-Key" (canonicalheader)
  [FAILURE] File: cmd/picoclaw/internal/jules/jules_test.go, Line: 56
    Message: non-canonical header "x-goog-api-key", instead use: "X-Goog-Api-Key" (canonicalheader)
  [FAILURE] File: cmd/picoclaw/internal/jules/jules.go, Line: 44
    Message: non-canonical header "x-goog-api-key", instead use: "X-Goog-Api-Key" (canonicalheader)
  [FAILURE] File: web/backend/api/boards_trigger.go, Line: 68
    Message: response body must be closed (bodyclose)

Since the CI specifically annotated these 10 issues, and I received a loopback forcing me to look at them, I am mandated to fix these specifically annotated ones to pass the build, despite 9 of them being out of scope for my `pkg/etl` change.
