#!/bin/bash
go install mvdan.cc/gofumpt@v0.7.0
go install github.com/daixiang0/gci@latest
go install github.com/segmentio/golines@latest

$(go env GOPATH)/bin/gofumpt -w cmd/picoclaw/internal/agent/create.go cmd/picoclaw-launcher-tui/main.go cmd/picoclaw-launcher-tui/internal/ui/app.go cmd/picoclaw-launcher-tui/internal/config/store.go cmd/picoclaw/internal/jules/jules.go
$(go env GOPATH)/bin/gofmt -w cmd/picoclaw/internal/jules/jules.go
$(go env GOPATH)/bin/gci write --skip-generated -s standard -s default -s 'prefix(jane)' web/backend/utils/banner.go pkg/tools/web/search.go pkg/tools/subagent_progress.go pkg/tools/browser.go
$(go env GOPATH)/bin/golines -m 120 -w cmd/picoclaw/internal/agent/create.go cmd/picoclaw-launcher-tui/main.go cmd/picoclaw-launcher-tui/internal/ui/app.go cmd/picoclaw-launcher-tui/internal/config/store.go cmd/picoclaw/internal/jules/jules.go web/backend/utils/banner.go pkg/tools/web/search.go pkg/tools/subagent_progress.go pkg/tools/browser.go
