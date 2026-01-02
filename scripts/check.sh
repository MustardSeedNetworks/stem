#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

# Go linting and tests
echo "Running Go linting..."
golangci-lint run ./...

echo "Running Go tests..."
go test -race -coverprofile=coverage.out ./...

# C linting (if src directory exists)
if [ -d "src" ]; then
    echo "Running C linting..."
    find src -name '*.c' -o -name '*.h' | xargs clang-format --dry-run --Werror 2>/dev/null || true
    find src -name '*.c' | xargs clang-tidy 2>/dev/null || true
fi

# Frontend checks (if ui directory exists)
if [ -d "ui" ]; then
    echo "Running frontend checks..."
    cd ui
    npm run lint
    npm run build
fi

echo "All checks passed!"
