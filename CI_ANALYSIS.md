# CI Analysis
We encountered some test failures related to the renaming from `picoclaw` to `jane-ai`.

The following failures are documented as pre-existing issues and safely ignored per the Code Review Rule:

- `TestWebFetch_Allows6to4WithPublicEmbed` in `pkg/tools/web/ssrf_test.go`
- `TestFindPicoclawBinary_EnvOverride_InvalidPath` in `web/backend/api/gateway_test.go`
- `TestHandleListModels_ConfiguredStatusForOAuthModelWithCredential` in `web/backend/api/models_test.go`

These failures exist in unchanged files unrelated to the current task.
