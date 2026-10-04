#!/usr/bin/env bash
# =============================================================================
# deploy-validate.sh - Validate an installed stem release
# =============================================================================
#
# Build contract rule 5: after an install, the service answering on the host
# must be the artifact the release published. Asserts that /__version reports
# the release's version and commit and a non-empty uiBuildHash.
#
# Usage: scripts/deploy-validate.sh --host HOST --release vX.Y.Z [--port PORT]
#
# Normally run as `make deploy-validate HOST=... RELEASE=...`. The commit is
# read from the release tag, so the tag must exist locally (`git fetch --tags`).
# Release builds embed the version without its `v` and the full commit SHA.
# =============================================================================

set -euo pipefail

HOST=""
RELEASE=""
PORT=8444
MAX_RETRIES=5
RETRY_DELAY=3

usage() {
    echo "Usage: $0 --host HOST --release vX.Y.Z [--port PORT]" >&2
    exit 2
}

while [ $# -gt 0 ]; do
    case "$1" in
    --host) HOST="${2:-}"; shift 2 ;;
    --release) RELEASE="${2:-}"; shift 2 ;;
    --port) PORT="${2:-}"; shift 2 ;;
    *) usage ;;
    esac
done
[ -n "$HOST" ] && [ -n "$RELEASE" ] || usage

command -v jq >/dev/null || { echo "jq is required" >&2; exit 2; }

EXPECTED_COMMIT=$(git rev-parse --verify --quiet "refs/tags/${RELEASE}^{commit}") || {
    echo "✗ no local tag ${RELEASE} (git fetch --tags)" >&2
    exit 2
}
EXPECTED_VERSION="${RELEASE#v}"
ENDPOINT="https://${HOST}:${PORT}/__version"

echo "Target:   $ENDPOINT"
echo "Expected: version=$EXPECTED_VERSION commit=$EXPECTED_COMMIT"

# The service may still be starting after an install restarted it.
for attempt in $(seq 1 "$MAX_RETRIES"); do
    if RESPONSE=$(curl -sfk --max-time 5 "$ENDPOINT"); then
        break
    fi
    if [ "$attempt" -eq "$MAX_RETRIES" ]; then
        echo "✗ $ENDPOINT not responding after $MAX_RETRIES attempts" >&2
        exit 1
    fi
    sleep "$RETRY_DELAY"
done

ACTUAL_VERSION=$(jq -r '.version // ""' <<<"$RESPONSE")
ACTUAL_COMMIT=$(jq -r '.commit // ""' <<<"$RESPONSE")
ACTUAL_UI_HASH=$(jq -r '.uiBuildHash // ""' <<<"$RESPONSE")

failed=0
check() {
    if [ "$2" = "$3" ]; then
        echo "✓ $1: $3"
    else
        echo "✗ $1: expected=$2 actual=$3"
        failed=1
    fi
}
check version "$EXPECTED_VERSION" "$ACTUAL_VERSION"
check commit "$EXPECTED_COMMIT" "$ACTUAL_COMMIT"
case "$ACTUAL_UI_HASH" in
"" | unknown) echo "✗ uiBuildHash missing: the UI was not embedded"; failed=1 ;;
*) echo "✓ uiBuildHash: $ACTUAL_UI_HASH" ;;
esac

if [ "$failed" -ne 0 ]; then
    echo "DEPLOYMENT VALIDATION FAILED"
    exit 1
fi
echo "DEPLOYMENT VALIDATION PASSED"
