#!/usr/bin/env bash
# Mutation test for check-release-workflow-contract.sh.
#
# A guard nobody tests is a guard nobody knows works. Each case below breaks
# one release invariant in a copy of the workflow and asserts the checker
# rejects it; the last case asserts it still accepts the real file, so a
# checker that rejects everything cannot pass either.

set -euo pipefail

source_workflow=".github/workflows/release.yml"
checker="./scripts/check-release-workflow-contract.sh"
fixture_dir=$(mktemp -d)
trap 'rm -rf "$fixture_dir"' EXIT

failures=0

single_matching_line() {
  local pattern="$1" matches
  matches=$(awk -v pattern="$pattern" '$0 ~ pattern { print }' "$source_workflow")
  if [ "$(wc -l <<<"$matches" | tr -d ' ')" -ne 1 ]; then
    echo "mutation source must occur once: $pattern" >&2
    exit 1
  fi
  printf '%s' "$matches"
}

assert_rejected() {
  local name="$1"
  local old="$2"
  local new="$3"
  local fixture="$fixture_dir/$name.yml"

  OLD="$old" NEW="$new" python3 - "$source_workflow" "$fixture" <<'PY'
import os
import pathlib
import sys

source = pathlib.Path(sys.argv[1]).read_text()
old = os.environ["OLD"]
count = source.count(old)
if count != 1:
    raise SystemExit(f"mutation source occurs {count} times, want 1: {old!r}")
pathlib.Path(sys.argv[2]).write_text(source.replace(old, os.environ["NEW"], 1))
PY

  if RELEASE_WORKFLOW_PATH="$fixture" "$checker" >/dev/null 2>&1; then
    echo "FAIL: contract accepted mutation: $name" >&2
    failures=$((failures + 1))
  else
    echo "ok: rejected $name"
  fi
}

# The publish predicate loses its event check, so a workflow_dispatch could
# publish. This is the seed regression that motivated the gate.
assert_rejected "publish-without-event-check" \
  "      - name: Run goreleaser (publish)
        if: \${{ github.event_name == 'push' && !inputs.dry_run }}" \
  "      - name: Run goreleaser (publish)
        if: \${{ !inputs.dry_run }}"

# The snapshot predicate stops being the publish predicate's complement, so
# some runs would do both and some neither.
assert_rejected "snapshot-predicate-drift" \
  "      - name: Run goreleaser (snapshot/dry-run)
        if: \${{ github.event_name != 'push' || inputs.dry_run }}" \
  "      - name: Run goreleaser (snapshot/dry-run)
        if: \${{ inputs.dry_run }}"

# The dispatch refusal disappears.
assert_rejected "dispatch-refusal-removed" \
  "        if: \${{ github.event_name == 'workflow_dispatch' && !inputs.dry_run }}" \
  "        if: \${{ false }}"

# Provenance would attest a snapshot.
assert_rejected "provenance-condition-loosened" \
  "    if: \${{ !cancelled() && github.event_name == 'push' && !inputs.dry_run && needs.goreleaser.result == 'success' }}" \
  "    if: \${{ !cancelled() && needs.goreleaser.result == 'success' }}"

# --skip=validate returns to the publish path.
assert_rejected "publish-skips-validation" \
  "        run: goreleaser release --clean" \
  "        run: goreleaser release --clean --skip=validate"

# The dirty-tree assertion is removed, which is what stopped --skip=validate
# coming back.
assert_rejected "workspace-assertion-removed" \
  "      - name: Assert the workspace is clean before goreleaser" \
  "      - name: Formerly asserted the workspace was clean"

# The builder image floats.
builder_image=$(single_matching_line '^[[:space:]]+image: goreleaser/goreleaser-cross:')
assert_rejected "unpinned-builder-image" \
  "$builder_image" \
  "      image: goreleaser/goreleaser-cross:latest"

# An action on the signing path floats to a tag.
attestation_action=$(single_matching_line '^[[:space:]]+uses: actions/attest-build-provenance@')
assert_rejected "unpinned-action" \
  "$attestation_action" \
  "        uses: actions/attest-build-provenance@v4"

# A supply-chain download loses its checksum.
syft_checksum=$(single_matching_line '^[[:space:]]+SYFT_SHA256:')
assert_rejected "syft-checksum-removed" \
  "$syft_checksum" \
  '          SYFT_SHA256: ""'

# A mutable latest-release lookup appears.
syft_download=$(single_matching_line 'anchore/syft/releases/download')
syft_download=${syft_download#*\"}
syft_download=${syft_download%\"}
assert_rejected "mutable-latest-lookup" \
  "$syft_download" \
  "https://github.com/anchore/syft/releases/latest/syft_linux_amd64.tar.gz"

# Workflow-level permissions stop being read-only.
assert_rejected "write-permissions-at-workflow-level" \
  "permissions:
  contents: read" \
  "permissions:
  contents: write"

# And the guard must still accept the real workflow — a checker that rejects
# everything would pass every case above.
if ! "$checker" >/dev/null 2>&1; then
  echo "FAIL: contract rejected the real workflow" >&2
  failures=$((failures + 1))
else
  echo "ok: accepted the real workflow"
fi

if [ "$failures" -ne 0 ]; then
  echo "$failures contract self-test failure(s)" >&2
  exit 1
fi

echo "release workflow contract self-test passed"
