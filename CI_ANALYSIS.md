# CI Analysis

The most recent CI run failed with numerous errors spanning the entire codebase (`pkg/channels/`, `cmd/picoclaw/`, `web/backend/`, etc.).

According to the project's **Code Review Rule**, I should generally ignore preexisting issues (like linter warnings, misspellings, and unrelated test failures) in files unrelated to my current task. Attempting to fix all 291 linter issues across dozens of files would severely pollute this pull request and violate the rule.

I am explicitly ignoring:
1. `TestNewPicoclawCommand` test failure (`main_test.go`) due to the "picoclaw" vs "jane-ai" rebrand.
2. `TestWebFetch_Allows6to4WithPublicEmbed` test failure (`ssrf_test.go`) due to known private IPv4 blocking rules.
3. OAuth and Config test failures.
4. Linter issues like `canonicalheader` (`x-goog-api-key`), `bodyclose`, `misspell` (`cancelled`), `govet` shadow declarations, and formatting errors in unmodified files.

I have strictly applied fixes only to the file I modified: `pkg/etl/pipeline.go` by running `gci` and `gofumpt` to resolve the formatting error.