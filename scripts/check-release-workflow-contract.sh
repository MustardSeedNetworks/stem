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
goreleaser_config="${GORELEASER_CONFIG_PATH:-$repo_root/.goreleaser.yml}"
release_please_config="${RELEASE_PLEASE_CONFIG_PATH:-$repo_root/.github/release-please-config.json}"

require() {
  local pattern="$1"
  if ! grep -Fq -- "$pattern" "$workflow"; then
    echo "release workflow contract missing: $pattern" >&2
    exit 1
  fi
}

# require_pin asserts the SHAPE of a Renovate-managed version pin rather than
# its literal value. The value lives in two places — the workflow and this
# gate — and only the workflow is what Renovate edits, so a literal `require`
# here made every toolchain bump red by construction and unfixable by a rebase
# (PR #1298 sat red for exactly this reason; seed hit it first as seed#2622).
#
# The gate's purpose survives intact: it exists so nobody silently UN-pins the
# release toolchain, and "pinned to an exact version and a full digest" is a
# shape a bump satisfies and an un-pinning does not.
#
# The optional third argument is how many lines must match. Syft and Cosign are
# installed twice, once in the Linux goreleaser container and once on the macOS
# runner that builds the .pkg, and each copy needs its own pin.
require_pin() {
  local what="$1"
  local pattern="$2"
  local want="${3:-1}"
  local found
  found=$(grep -Ec -- "$pattern" "$workflow" || true)
  if [ "$found" -lt "$want" ]; then
    echo "release workflow contract: $what is missing or no longer pinned in $want place(s), found $found (want /$pattern/)" >&2
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
  "if: \${{ !cancelled() && github.event_name == 'push' && !inputs.dry_run && needs.goreleaser.result == 'success' && needs.macos-pkg.result == 'success' }}"

# The macOS .pkg is a required release artifact, so provenance waits for it and
# attests the subjects it hands over. Publishing it carries the same predicate
# as goreleaser's, and a snapshot only uploads the .pkg for inspection.
require "needs: [goreleaser, macos-pkg]"
require "needs.macos-pkg.outputs.hashes"
require "deploy/macos/build-pkg.sh"
require "stem-*-darwin-arm64.tar.gz"
require_step_condition "Install Syft and Cosign" \
  "if: \${{ github.event_name == 'push' && !inputs.dry_run }}"
require_step_condition "SBOM, sign, and publish the .pkg" \
  "if: \${{ github.event_name == 'push' && !inputs.dry_run }}"
require_step_condition "Capture provenance subjects including the .pkg" \
  "if: \${{ github.event_name == 'push' && !inputs.dry_run }}"
require_step_condition "Upload dry-run .pkg for inspection" \
  "if: \${{ github.event_name != 'push' || inputs.dry_run }}"
# The .pkg must be Developer ID signed, notarized and stapled on a tag push.
# The resolve step is what makes that fail closed: with any secret missing a
# push errors out, and only a non-push run may fall back to an unsigned build.
# Every signing step keys off its SIGN output, and cosign, the SBOM and the
# checksum come after stapling, which rewrites the package bytes.
require_step_condition "Import signing identities" "if: \${{ env.SIGN == 'true' }}"
require_step_condition "Check the signed package contract" "if: \${{ env.SIGN == 'true' }}"
require_step_condition "Notarize and staple the .pkg" "if: \${{ env.SIGN == 'true' }}"
require_step_condition "Remove the signing keychain" "if: \${{ always() && env.SIGN == 'true' }}"
require_step_condition "Resolve the signing mode" "elif [ \"\${EVENT}\" = \"push\" ] || [ \"\${SIGN_MACOS}\" = \"true\" ]; then"
require_step_condition "Resolve the signing mode" "SIGN_MACOS: \${{ inputs.sign_macos }}"
require "sign_macos:"
require_step_condition "Resolve the signing mode" "exit 1"
require_step_condition "Resolve the signing mode" "EVENT: \${{ github.event_name }}"
for secret in MACOS_SIGNING_P12_BASE64 MACOS_SIGNING_P12_PASSWORD APPLE_NOTARY_KEY_P8_BASE64 APPLE_NOTARY_KEY_ID APPLE_NOTARY_ISSUER_ID; do
  require_step_condition "Resolve the signing mode" "\${{ secrets.$secret }}"
done
require_step_condition "Build the .pkg" 'export STEM_SIGN_IDENTITY="Developer ID Application"'
require_step_condition "Build the .pkg" 'export STEM_INSTALLER_IDENTITY="Developer ID Installer"'
require_step_condition "Check the signed package contract" 'flags=0x10000(runtime)'
require_step_condition "Notarize and staple the .pkg" "xcrun notarytool submit"
require_step_condition "Notarize and staple the .pkg" '!= "Accepted"'
require_step_condition "Notarize and staple the .pkg" "xcrun stapler staple"
require_step_condition "Notarize and staple the .pkg" "xcrun stapler validate"
require_step_condition "Notarize and staple the .pkg" "spctl --assess --type install"
require_step_condition "Notarize and staple the .pkg" "pkgutil --check-signature"
require_step_condition "Import signing identities" "security import"
require_step_condition "Import signing identities" "Developer ID Installer"
require "security delete-keychain"
if ! awk '
  /- name: Notarize and staple the .pkg/ { notarize = NR }
  /- name: Install Syft and Cosign$/ { tools = NR }
  /- name: SBOM, sign, and publish the .pkg/ { publish = NR }
  END { exit !(notarize && tools && publish && notarize < tools && notarize < publish) }
' "$workflow"; then
  echo "release workflow contract: notarize and staple must run before the SBOM and cosign steps" >&2
  exit 1
fi

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
# Renovate owns these values; the gate owns their shape. See require_pin.
require_pin 'the goreleaser-cross image' \
  '^ +image: goreleaser/goreleaser-cross:v[0-9]+\.[0-9]+\.[0-9]+(-[0-9]+)?(-v[0-9.]+)?@sha256:[0-9a-f]{64}$'
require_pin 'SYFT_VERSION' '^ +SYFT_VERSION: "[0-9]+\.[0-9]+\.[0-9]+"$' 2
require_pin 'SYFT_SHA256' '^ +SYFT_SHA256: "[0-9a-f]{64}"$' 2
require_pin 'COSIGN_VERSION' '^ +COSIGN_VERSION: "v[0-9]+\.[0-9]+\.[0-9]+"$' 2
require_pin 'COSIGN_SHA256' '^ +COSIGN_SHA256: "[0-9a-f]{64}"$' 2
require "| sha256sum -c -"
require "| shasum -a 256 -c -"

# Releases are immutable once published, and an immutable release refuses
# every upload: v0.26.7 went out with zero assets because release-please
# published it before goreleaser ran (#1611). The release must therefore stay a
# draft from creation until the last asset is on it. Each of the three hand-offs
# is pinned, because losing any one of them reproduces the empty release.
if ! jq -e '.packages["."] | .draft == true and ."force-tag-creation" == true' \
  "$release_please_config" >/dev/null; then
  echo "release workflow contract: release-please must create a draft and force the tag ($release_please_config)" >&2
  exit 1
fi
for setting in 'draft: true' 'use_existing_draft: true'; do
  if ! grep -Eq "^  ${setting}\$" "$goreleaser_config"; then
    echo "release workflow contract: goreleaser release block is missing '$setting' ($goreleaser_config)" >&2
    exit 1
  fi
done
require "gh release edit \"\$TARGET_TAG\" --draft=false"
if [ "$(grep -E '^      - name:' "$workflow" | tail -n 1)" != "      - name: Publish the completed release" ]; then
  echo "release workflow contract: publishing the draft must be the workflow's last step" >&2
  exit 1
fi

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
