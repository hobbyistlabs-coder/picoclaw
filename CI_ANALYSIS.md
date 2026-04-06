# CI Failure Analysis

## Pre-existing Test Failures
The following test failures are pre-existing and unrelated to the current changes. They are safely ignored per the codebase's Code Review Rule:

1. `main_test.go`: `TestNewPicoclawCommand` fails due to a project rename from 'picoclaw' to 'jane-ai'. The test still expects the old strings.
2. `config_test.go`: `TestDefaultConfig_WorkspacePath_Default` fails for the same reason, expecting `.picoclaw` instead of `.jane-ai`.
3. `gateway_test.go`: `TestFindPicoclawBinary_EnvOverride_InvalidPath` fails as it's looking for the old `picoclaw` binary name.
4. `models_test.go`: `TestHandleListModels_ConfiguredStatusForOAuthModelWithCredential` fails due to a pre-existing issue with OAuth mocking.
5. `ssrf_test.go`: `TestWebFetch_Allows6to4WithPublicEmbed` fails due to an IPv4 proxy blocking issue.

These issues should be resolved in a separate PR dedicated to fixing pre-existing test failures across the codebase.
