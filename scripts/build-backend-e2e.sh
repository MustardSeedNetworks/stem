#!/usr/bin/env bash
#
# Build bin/stem for the E2E suite and the phone-width gate, from the UI
# already built into internal/api/ui.
#
# Why this exists rather than `make build`: neither caller has what the
# Makefile's `go` target needs on Linux. The E2E job runs inside
# mcr.microsoft.com/playwright, which ships no make and no C toolchain; the
# phone-width gate runs on a bare runner of the reusable workflow in
# MustardSeedNetworks/.github, which builds the UI itself and has no setup
# step for the C dataplane. Neither needs it: nothing either drives sends test
# traffic, so the binary is built CGO_ENABLED=0.
#
# The ldflags MUST mirror the Makefile's LDFLAGS. Per the universal build
# contract, a binary built without them reports "unknown" from /__version.
# A shallow checkout (the phone-width gate's) has no tag to describe, so the
# version falls back to the bare commit; the E2E job fetches full history
# because smoke.spec.ts asserts a tagged version.
set -euo pipefail

cd "$(dirname "$0")/.."

VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION_PKG="github.com/MustardSeedNetworks/stem/internal/version"

if [ -z "$(find internal/api/ui -type f ! -name .gitkeep -print -quit)" ]; then
  echo "::error::internal/api/ui is empty — build the UI first; refusing to build a UI-less binary" >&2
  exit 1
fi
UI_BUILD_HASH=$(find internal/api/ui -type f -exec md5sum {} \; | sort | md5sum | cut -d' ' -f1)

echo "building e2e backend: version=${VERSION} commit=${COMMIT} uiBuildHash=${UI_BUILD_HASH}"

mkdir -p bin
CGO_ENABLED=0 go build -trimpath -buildvcs=false \
  -ldflags "-s -w \
    -X ${VERSION_PKG}.Version=${VERSION} \
    -X ${VERSION_PKG}.Commit=${COMMIT} \
    -X ${VERSION_PKG}.BuildTime=${BUILD_TIME} \
    -X ${VERSION_PKG}.UIBuildHash=${UI_BUILD_HASH}" \
  -o bin/stem ./cmd/stem/
