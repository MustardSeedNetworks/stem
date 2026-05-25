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
    find src include tests -type f \( -name '*.c' -o -name '*.h' \) -exec clang-format --dry-run --Werror {} +
    if [ -f "build/compile_commands.json" ]; then
        clang_tidy_db="build"
    elif [ -f "compile_commands.json" ]; then
        clang_tidy_db="."
    else
        echo "compile_commands.json not found. Generate with: bear -- make dataplane c-test"
        exit 1
    fi
    find src include tests -type f -name '*.c' -exec clang-tidy -p "$clang_tidy_db" -warnings-as-errors=* {} +
fi

# Frontend checks (if ui directory exists)
if [ -d "ui" ]; then
    echo "Running frontend checks..."
    cd ui
    npm run lint
    npm run build
fi

echo "All checks passed!"
