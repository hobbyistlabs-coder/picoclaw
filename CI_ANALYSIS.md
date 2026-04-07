# CI Analysis
We encountered some test failures and linter errors related to pre-existing technical debt.

The following failures are documented as pre-existing issues and safely ignored per the Code Review Rule:

## Test Failures
- `TestWebFetch_Allows6to4WithPublicEmbed` in `pkg/tools/web/ssrf_test.go` (IPv4 proxy blocking)
- `TestFindPicoclawBinary_EnvOverride_InvalidPath` in `web/backend/api/gateway_test.go` (gateway binary finding)
- `TestHandleListModels_ConfiguredStatusForOAuthModelWithCredential` in `web/backend/api/models_test.go` (oauth mock issues)

## Linter Errors
- There are 291 widespread linter errors (`gofumpt`, `golines`, `gci`, `canonicalheader`, `bodyclose`, `govet`, `misspell`, etc.) across 239 unmodified files in the repository.
- These format and static analysis issues are pre-existing in untouched dependencies and unrelated files.
- Modifying 239 files to fix formatting would severely pollute the pull request and violate the Code Review Rule regarding over-reaching. Therefore, these linter warnings are safely ignored.
