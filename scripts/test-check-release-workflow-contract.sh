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
assert_rejected "unpinned-builder-image" \
  "      image: goreleaser/goreleaser-cross:v1.27.0@sha256:3ce3506ee9179c4122ba0b5dc13ab564ff259fb65f45bfad005ddd5e4a3d326d" \
  "      image: goreleaser/goreleaser-cross:latest"

# An action on the signing path floats to a tag.
assert_rejected "unpinned-action" \
  "        uses: actions/attest-build-provenance@4d101475d8b20a2381f78447822ac1eab6504dd8 # v4.2.2" \
  "        uses: actions/attest-build-provenance@v4"

# A supply-chain download loses its checksum.
assert_rejected "syft-checksum-removed" \
  '          SYFT_SHA256: "d654f678b709eb53c393d38519d5ed7d2e57205529404018614cfefa0fb2b5ca"' \
  '          SYFT_SHA256: ""'

# A mutable latest-release lookup appears.
assert_rejected "mutable-latest-lookup" \
  "https://github.com/anchore/syft/releases/download/v\${SYFT_VERSION}/syft_\${SYFT_VERSION}_linux_amd64.tar.gz" \
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
