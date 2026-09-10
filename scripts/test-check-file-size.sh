#!/usr/bin/env bash

set -euo pipefail

checker=$(cd "$(dirname "$0")" && pwd)/check-file-size.sh
fixture=$(mktemp -d)
trap 'rm -rf "$fixture"' EXIT

mkdir -p "$fixture/scripts" "$fixture/internal/example"
printf 'package example\n' > "$fixture/internal/example/live.go"

run_with_baseline() {
    BASELINE_FILE=scripts/file-size-baseline.txt SCAN_ROOT="$fixture" "$checker" 2>&1
}

printf 'internal/example/live.go 10\n' > "$fixture/scripts/file-size-baseline.txt"
run_with_baseline >/dev/null

printf 'internal/example/missing.go 10\n' > "$fixture/scripts/file-size-baseline.txt"
if output=$(run_with_baseline); then
    echo "FAIL: accepted a baseline path that does not exist" >&2
    exit 1
fi
grep -Fq "Baseline path does not exist: internal/example/missing.go" <<<"$output"

printf 'docs/not-scanned.md 10\n' > "$fixture/scripts/file-size-baseline.txt"
if output=$(run_with_baseline); then
    echo "FAIL: accepted a baseline path outside the scan" >&2
    exit 1
fi
grep -Fq "Baseline path is not scanned: docs/not-scanned.md" <<<"$output"

echo "file-size baseline self-test passed"
