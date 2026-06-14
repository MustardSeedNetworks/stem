#!/usr/bin/env bash
# check-json-casing.sh — JSON wire-casing discipline gate.
#
# Convention (mirrors seed's ADR-0010): JSON API wire tags are camelCase. The
# config-file format and SQL columns are snake_case by design and are NOT
# scanned here.
#
# Stem's own wire/domain tags are all camelCase today, so the baseline
# (scripts/json-casing-baseline.txt) is currently EMPTY. It is an
# EXTERNAL-BOUNDARY allow-list, not a debt list: the only entries that should
# ever be added are keys from an external contract we must parse verbatim
# (e.g. a vendor tool's `-json` output). Do NOT add OUR keys here — fix the
# casing to camelCase instead.
#
# This gate scans `json:"..."` struct tags in the scanned dirs for snake_case
# keys and compares them to the committed allow-list. It is a RATCHET:
#   - it FAILS if a NEW snake_case tag appears that is not in the allow-list, and
#   - it passes when violations only shrink.
#
# Regenerate the allow-list only when adding a genuinely-external contract key:
#   scripts/check-json-casing.sh --update
#
# Run locally: scripts/check-json-casing.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

BASELINE="scripts/json-casing-baseline.txt"
SCAN_DIRS=("internal/api" "internal/reflector")

# current_violations prints sorted "path\ttag" lines for every snake_case JSON
# tag in the scanned dirs (excluding tests). A snake_case tag is a json tag
# whose key contains an underscore between lowercase/alnum segments.
current_violations() {
	# `|| true`: grep exits 1 when there are zero snake_case tags (the healthy
	# state for a fully-camelCase repo); under `set -o pipefail` that would
	# otherwise abort the script. An empty result is success, not failure.
	grep -rnoE 'json:"[a-z][a-z0-9]*_[a-z0-9_]+[^"]*"' "${SCAN_DIRS[@]}" --include='*.go' 2>/dev/null \
		| grep -v '_test\.go:' \
		| sed -E 's/^([^:]+):[0-9]+:json:"([^",]+).*/\1\t\2/' \
		| sort -u || true
}

if [[ "${1:-}" == "--update" ]]; then
	current_violations >"$BASELINE"
	echo "Wrote $(wc -l <"$BASELINE" | tr -d ' ') baselined snake_case JSON tag(s) to $BASELINE"
	exit 0
fi

if [[ ! -f "$BASELINE" ]]; then
	echo "::error::$BASELINE missing — run scripts/check-json-casing.sh --update" >&2
	exit 1
fi

# New violations = current entries not present in the baseline.
new=$(comm -23 <(current_violations) <(sort -u "$BASELINE") || true)

if [[ -n "$new" ]]; then
	echo "::error::New snake_case JSON wire tag(s) introduced — use camelCase:" >&2
	echo "$new" | sed 's/^/  /' >&2
	echo "" >&2
	echo "If this is a protocol-mandated external key, grandfather it:" >&2
	echo "  scripts/check-json-casing.sh --update   # then commit the baseline" >&2
	exit 1
fi

remaining=$(current_violations | comm -12 - <(sort -u "$BASELINE") | wc -l | tr -d ' ')
echo "JSON casing gate OK — no new snake_case wire tags. ${remaining} external-contract tag(s) on the allow-list."
