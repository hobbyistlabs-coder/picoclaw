#!/bin/bash
set -e

# Extract the list of files with issues from the CI log dump
FILES="
cmd/picoclaw/internal/jules/jules.go
pkg/tools/mcp2cli.go
web/backend/utils/banner.go
pkg/tools/web/search.go
pkg/tools/web/fetch.go
pkg/tools/mcp2cli_test.go
pkg/tools/browser.go
pkg/tools/alpaca/alpaca.go
pkg/health/resource_tracker.go
"

for file in $FILES; do
    echo "Formatting $file"
    go run mvdan.cc/gofumpt@latest -w "$file"
    go run github.com/segmentio/golines@latest -w "$file"
    go run github.com/daixiang0/gci@latest write --skip-generated -s standard -s default -s 'prefix(jane)' "$file"
    go fmt "$file"
done

# Fix godoc in pkg/tools/mcp2cli.go manually
sed -i 's/\/\/ Execute MCP tool locally/\/\/ Execute parses and executes the tool locally/g' pkg/tools/mcp2cli.go
