# CI Analysis

The following test failures are preexisting issues and are unrelated to the current PR. They are safely ignored per the Code Review Rule:

- `cmd/picoclaw/main_test.go`: Fails due to expected "jane-ai" vs "picoclaw" output.
- `pkg/config/config_test.go`: Fails due to expected ".jane-ai" vs ".picoclaw" output.
- `pkg/tools/web/ssrf_test.go`: Fails due to IPv4 proxy blocking logic.
- `web/backend/api/...`: Fails due to existing model configuration testing logic.
