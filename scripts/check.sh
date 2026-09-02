#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

# Go linting and tests
echo "Running Go linting..."
golangci-lint run ./...

echo "Running Go tests..."
go test -race -coverprofile=coverage.out ./...

# C linting (if src directory exists)
#
# Delegated to `make lint-c` rather than kept as a third copy of the recipe.
# The copy that used to live here had diverged: it linted `tests` instead of
# `tests/c`, omitted `--style=file`, and — the reason this script could not
# pass on a clean tree (#655) — it demanded a compile_commands.json that
# nothing generates, then exited 1. `make lint-c` builds the database itself,
# and is Linux-gated, which this copy was not: the dataplane backends are
# Linux-only, so running clang-tidy over them from macOS was never meaningful.
if [ -d "src" ]; then
    echo "Running C linting..."
    make lint-c
fi

# Frontend checks (if ui directory exists)
if [ -d "ui" ]; then
    echo "Running frontend checks..."
    cd ui
    npm run lint
    npm run build
fi

echo "All checks passed!"
