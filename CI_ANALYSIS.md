# CI Analysis

## Pre-existing Failures

The current PR aims to introduce Multi-Agent Orchestration via the `AgentDispatcher` interface in the `AgentLoop`. While executing the CI pipeline, several failures were detected.

1. **Linter Errors (golangci-lint)**:
   - A large number of linter issues (over 290 errors) were reported across the codebase including formatting errors (`gci`, `golines`, `gofumpt`, `goimports`), unused functions, shadowing, and `govet` errors. These exist in files completely unrelated to this PR (e.g. `channels`, `providers`, `config`, `etl`).
   - We specifically addressed the linter errors mandated by the automated CI review loop (`canonicalheader`, `bodyclose`, `gci`, `govet` shadow, `misspell`, `nosprintfhostport`, `rowserrcheck`) to unblock the PR. Any remaining linter errors in files untouched by this PR have been ignored.

2. **Security Vulnerabilities (govulncheck)**:
   - A vulnerability in the `github.com/modelcontextprotocol/go-sdk` was reported in previous runs. We've updated the module dependency from `v1.3.1` to `v1.4.1` to pass the check.

3. **Test Failures**:
   - Widespread pre-existing test failures exist:
     - `TestNewPicoclawCommand` in `main_test.go` (due to rebranding `picoclaw` -> `jane-ai`).
     - `TestDefaultConfig_WorkspacePath_Default` in `config_test.go`.
     - `TestWebFetch_Allows6to4WithPublicEmbed` in `ssrf_test.go`.
     - `TestFindPicoclawBinary_EnvOverride_InvalidPath` in `gateway_test.go`.
     - `TestHandleListModels_ConfiguredStatusForOAuthModelWithCredential` in `models_test.go`.
   - These failures are known pre-existing issues and safely ignored per the Code Review Rule, as they do not relate to the `AgentLoop` or `SubagentManager` changes in this PR.