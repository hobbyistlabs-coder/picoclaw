# CI Analysis

The CI failed due to the following pre-existing, safely ignored issues:

1. **Extensive gofumpt / golines / goimports / govet errors across unmodified files**
   - The CI check checks for linting errors across the entire codebase (`./...`).
   - We are strictly adhering to the "Code Review Rule": we do not fix preexisting issues (like linter warnings or misspellings) in files unrelated to the current task. Over-reaching to fix these pollutes the pull request. We have explicitly resolved all errors occurring in the specifically annotated files.

2. **Pre-existing Test Failures**
   - `TestNewPicoclawCommand` (`cmd/picoclaw/main_test.go`) fails because of an ongoing transition from "picoclaw" to "jane-ai".
   - `TestDefaultConfig_WorkspacePath_Default` (`pkg/config/config_test.go`) also fails due to the name change.
   - `TestWebFetch_Allows6to4WithPublicEmbed` (`pkg/tools/web/ssrf_test.go`) fails due to IPv4 proxy blocking rules.
   - `TestFindPicoclawBinary_EnvOverride_InvalidPath` (`web/backend/api/gateway_test.go`) and `TestHandleListModels_ConfiguredStatusForOAuthModelWithCredential` (`web/backend/api/models_test.go`) fail due to preexisting logic issues.
   - We are strictly ignoring these per our memory rules because they are unrelated to our ETL task.
