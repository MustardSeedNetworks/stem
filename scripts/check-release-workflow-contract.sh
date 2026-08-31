#!/usr/bin/env bash
# check-release-workflow-contract.sh — release-mode and supply-chain regression gate.
#
# The release workflow decides, from an `if:` expression, whether a run signs
# and publishes artifacts or produces a throwaway snapshot. Nothing else checks
# those expressions: a run that publishes when it should not still goes green,
# because publishing IS the success path. seed added the equivalent gate after
# exactly that, and it caught two regressions in its first week.
#
# Ported from seed's, against stem's own predicates and pins.

set -euo pipefail

workflow="${RELEASE_WORKFLOW_PATH:-.github/workflows/release.yml}"
# Local composite refs (`uses: ./.github/actions/...`) resolve against the repo
# root, not the workflow's directory. Overridable so the self-test can stage a
# mutated composite without touching the tree.
repo_root="${RELEASE_REPO_ROOT:-.}"

require() {
  local pattern="$1"
  if ! grep -Fq -- "$pattern" "$workflow"; then
    echo "release workflow contract missing: $pattern" >&2
    exit 1
  fi
}

# require_step_condition pins a condition to the step that must carry it.
# A bare `require` cannot: the publish predicate appears on more than one step,
# so dropping it from one of them would still match elsewhere and pass.
require_step_condition() {
  local step="$1"
  local condition="$2"

  if ! awk -v step="- name: $step" -v cond="$condition" '
    index($0, step) { in_step = 1; next }
    in_step && index($0, cond) { found = 1 }
    in_step && /^      - name:/ { exit }
    END { exit !found }
  ' "$workflow"; then
    echo "release workflow contract: step \"$step\" is missing its condition: $condition" >&2
    exit 1
  fi
}

# require_job_condition pins a condition to the job that must carry it, the way
# require_step_condition does for steps.
require_job_condition() {
  local job="$1"
  local condition="$2"

  if ! awk -v job="  $job:" -v cond="$condition" '
    $0 == job { in_job = 1; next }
    in_job && index($0, cond) { found = 1 }
    in_job && /^  [a-z0-9-]+:$/ { exit }
    END { exit !found }
  ' "$workflow"; then
    echo "release workflow contract: job \"$job\" is missing its condition: $condition" >&2
    exit 1
  fi
}

# Every external action reachable from the release path must be SHA-pinned.
# Local composites are followed rather than skipped: the Node/npm pin lives in
# .github/actions/setup-node, so skipping them would leave the actions it calls
# unchecked on the path that produces signed, attested artifacts.
validate_action_pins() {
  local file="$1"
  local line
  local ref
  local composite

  while IFS= read -r line; do
    ref=$(awk '{ for (i = 1; i <= NF; i++) if ($i == "uses:") { print $(i + 1); exit } }' <<<"$line")
    if [[ "$ref" == ./* ]]; then
      composite="$repo_root/${ref#./}/action.yml"
      if [[ ! -f "$composite" ]]; then
        echo "release workflow contract references a missing composite: $ref" >&2
        exit 1
      fi
      validate_action_pins "$composite"
      continue
    fi
    if [[ ! "$ref" =~ ^[^@[:space:]]+@[0-9a-f]{40}$ ]]; then
      echo "release workflow contract has mutable action reference: $line" >&2
      exit 1
    fi
  done < <(grep -E '^[[:space:]]*(-[[:space:]]+)?uses:' "$file")
}

# Publishing is reserved for pushed tags: only a push can satisfy the verify-tag
# assertion that the commit passed CI Complete, and a v* tag can be created on
# any commit. Both halves are pinned -- the publish predicate and the snapshot
# predicate that must be its exact complement -- so a change that drops the
# event check from one of them cannot pass silently.
require_step_condition "Run goreleaser (publish)" \
  "if: \${{ github.event_name == 'push' && !inputs.dry_run }}"
require_step_condition "Capture artifact hashes for SLSA provenance" \
  "if: \${{ github.event_name == 'push' && !inputs.dry_run }}"
require_step_condition "Run goreleaser (snapshot/dry-run)" \
  "if: \${{ github.event_name != 'push' || inputs.dry_run }}"
require_step_condition "Upload dry-run artifact bundle for inspection" \
  "if: \${{ github.event_name != 'push' || inputs.dry_run }}"
require_step_condition "Refuse a manual dispatch that asks to publish" \
  "if: \${{ github.event_name == 'workflow_dispatch' && !inputs.dry_run }}"
require_job_condition "provenance" \
  "if: \${{ !cancelled() && github.event_name == 'push' && !inputs.dry_run && needs.goreleaser.result == 'success' }}"

# goreleaser runs with validation on when publishing. --skip=validate on this
# path is how a release stops checking it is building a clean tree at the
# expected tag; it sat here once already.
require "run: goreleaser release --clean"
require "run: goreleaser release --snapshot --clean --skip=publish,announce"

# Checked separately from the require above, which is a substring match and so
# is satisfied by "goreleaser release --clean --skip=validate" -- the exact
# regression this is here to stop. The self-test found that hole.
#
# The snapshot path legitimately skips publish and announce; nothing may skip
# validate, which is what checks the build is running from a clean tree at the
# expected tag. Comment lines are excluded -- the workflow explains in prose
# why --skip=validate was removed, and that explanation must not trip the gate
# that keeps it removed.
if grep -nE -- '^[[:space:]]*[^#[:space:]].*--skip=[a-z,]*validate' "$workflow"; then
  echo "release workflow contract: goreleaser must not skip validation" >&2
  exit 1
fi

# The dirty-tree assertion that replaced --skip=validate. Without it, the next
# person to hit a dirty workspace gets a goreleaser error whose easiest fix is
# to skip validation again.
require "- name: Assert the workspace is clean before goreleaser"

# Pinned toolchain and checksum-verified downloads on the signing path.
require 'image: goreleaser/goreleaser-cross:v1.27.0@sha256:3ce3506ee9179c4122ba0b5dc13ab564ff259fb65f45bfad005ddd5e4a3d326d'
require 'SYFT_VERSION: "1.46.0"'
require 'SYFT_SHA256: "d654f678b709eb53c393d38519d5ed7d2e57205529404018614cfefa0fb2b5ca"'
require 'COSIGN_VERSION: "v3.1.3"'
require 'COSIGN_SHA256: "4629c757b7618056f8ddd7e2625ae9fdd94c0372a65049520bc7d9df9efc7f71"'
require "| sha256sum -c -"

validate_action_pins "$workflow"

if grep -Fq '/releases/latest' "$workflow"; then
  echo "release workflow contract contains a mutable latest-release lookup" >&2
  exit 1
fi

if ! awk '
  /^permissions:$/ { top_permissions = 1; next }
  top_permissions && /^  contents: read$/ { contents_read = 1; next }
  top_permissions && /^  [a-z-]+:/ { unexpected = 1; next }
  top_permissions && /^[^ ]/ { exit }
  END { exit !(top_permissions && contents_read && !unexpected) }
' "$workflow"; then
  echo "release workflow contract missing read-only workflow permissions" >&2
  exit 1
fi

echo "release workflow contract OK"
