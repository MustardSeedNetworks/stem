#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
go run ./scripts/requestfields internal/api scripts/request-field-read-baseline.txt
