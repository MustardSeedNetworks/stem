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

# first_matching_line is for the Syft and Cosign pins, which appear once in the
# Linux container and once on the macOS runner; the first is the Linux one.
first_matching_line() {
  awk -v pattern="$1" '$0 ~ pattern { print; exit }' "$source_workflow"
}

# mutate copies $1 to $2 with the one occurrence of $OLD replaced by $NEW.
mutate() {
  python3 - "$1" "$2" <<'PY'
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
}

report() {
  local name="$1"
  shift
  if env "$@" "$checker" >/dev/null 2>&1; then
    echo "FAIL: contract accepted mutation: $name" >&2
    failures=$((failures + 1))
  else
    echo "ok: rejected $name"
  fi
}

assert_rejected() {
  local name="$1"
  local fixture="$fixture_dir/$name.yml"

  OLD="$2" NEW="$3" mutate "$source_workflow" "$fixture"
  report "$name" RELEASE_WORKFLOW_PATH="$fixture"
}

# assert_config_rejected mutates a release config the workflow depends on,
# handed to the checker through its path override ($2).
assert_config_rejected() {
  local name="$1"
  local var="$2"
  local source="$3"
  local fixture="$fixture_dir/$name.${source##*.}"

  OLD="$4" NEW="$5" mutate "$source" "$fixture"
  report "$name" "$var=$fixture"
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
  "    if: \${{ !cancelled() && github.event_name == 'push' && !inputs.dry_run && needs.goreleaser.result == 'success' && needs.macos-pkg.result == 'success' }}" \
  "    if: \${{ !cancelled() && needs.goreleaser.result == 'success' }}"

# The .pkg is signed and published on a pushed release only; a dispatch must
# not upload it to a release.
assert_rejected "pkg-publish-without-event-check" \
  "      - name: SBOM, sign, and publish the .pkg
        if: \${{ github.event_name == 'push' && !inputs.dry_run }}" \
  "      - name: SBOM, sign, and publish the .pkg
        if: \${{ !inputs.dry_run }}"

# The .pkg's provenance subjects are computed on a release only.
assert_rejected "pkg-subjects-without-event-check" \
  "      - name: Capture provenance subjects including the .pkg
        if: \${{ github.event_name == 'push' && !inputs.dry_run }}" \
  "      - name: Capture provenance subjects including the .pkg
        if: \${{ !inputs.dry_run }}"

# Provenance stops waiting for the .pkg, so it would ship unattested.
assert_rejected "provenance-ignores-pkg" \
  "    needs: [goreleaser, macos-pkg]" \
  "    needs: [goreleaser]"

assert_rejected \
  "pkg-signing-does-not-fail-closed-on-push" \
  "elif [ \"\${EVENT}\" = \"push\" ] || [ \"\${SIGN_MACOS}\" = \"true\" ]; then" \
  "elif [ \"\${EVENT}\" = \"never\" ] || [ \"\${SIGN_MACOS}\" = \"true\" ]; then"
assert_rejected \
  "pkg-signing-ignores-p12-secret" \
  "P12_BASE64: \${{ secrets.MACOS_SIGNING_P12_BASE64 }}
          P12_PASSWORD: \${{ secrets.MACOS_SIGNING_P12_PASSWORD }}
          NOTARY_KEY_BASE64: \${{ secrets.APPLE_NOTARY_KEY_P8_BASE64 }}" \
  "P12_BASE64: \${{ secrets.MACOS_SIGNING_P12_BASE64_RENAMED }}
          P12_PASSWORD: \${{ secrets.MACOS_SIGNING_P12_PASSWORD }}
          NOTARY_KEY_BASE64: \${{ secrets.APPLE_NOTARY_KEY_P8_BASE64 }}"
assert_rejected \
  "pkg-notarize-step-gated-off-dispatch" \
  $'      - name: Notarize and staple the .pkg\n        if: ${{ env.SIGN == \'true\' }}' \
  $'      - name: Notarize and staple the .pkg\n        if: ${{ env.SIGN == \'true\' && github.event_name == \'push\' }}'
assert_rejected \
  "pkg-notarize-removed" \
  "xcrun notarytool submit" \
  "xcrun notarytool removed"
assert_rejected \
  "pkg-notarize-verdict-unchecked" \
  '!= "Accepted"' \
  '!= "Anything"'
assert_rejected \
  "pkg-staple-removed" \
  "xcrun stapler staple \"dist/\${PKG}\"" \
  'echo no staple'
assert_rejected \
  "pkg-staple-not-validated" \
  "xcrun stapler validate \"dist/\${PKG}\"" \
  'echo no validate'
assert_rejected \
  "pkg-gatekeeper-assessment-removed" \
  'spctl --assess --type install --verbose=2' \
  'echo no spctl'
assert_rejected \
  "pkg-installer-identity-dropped" \
  'export STEM_INSTALLER_IDENTITY="Developer ID Installer"' \
  'export STEM_INSTALLER_IDENTITY_UNUSED=1'
assert_rejected \
  "pkg-hardened-runtime-unchecked" \
  'flags=0x10000(runtime)' \
  'flags=0x0(none)'
assert_rejected \
  "pkg-keychain-not-removed-always" \
  $'      - name: Remove the signing keychain\n        if: ${{ always() && env.SIGN == \'true\' }}' \
  $'      - name: Remove the signing keychain\n        if: ${{ env.SIGN == \'true\' }}'
assert_rejected \
  "pkg-notarized-after-publish" \
  "      - name: Capture provenance subjects including the .pkg" \
  $'      - name: Notarize and staple the .pkg\n        run: true\n\n      - name: Capture provenance subjects including the .pkg'

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
syft_checksum=$(first_matching_line '^[[:space:]]+SYFT_SHA256:')
assert_rejected "syft-checksum-removed" \
  "$syft_checksum" \
  '          SYFT_SHA256: ""'

# The macOS copy of the pin loses its checksum too.
macos_syft_checksum=$(awk '/^[[:space:]]+SYFT_SHA256:/ { n++; if (n == 2) { print; exit } }' "$source_workflow")
assert_rejected "macos-syft-checksum-removed" \
  "$macos_syft_checksum" \
  '          SYFT_SHA256: ""'

# A mutable latest-release lookup appears.
syft_download=$(first_matching_line 'anchore/syft/releases/download')
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

# A step added after publishing would upload into an immutable release.
assert_rejected "upload-after-publish" \
  "          gh release edit \"\$TARGET_TAG\" --draft=false --repo \"\${GITHUB_REPOSITORY}\"" \
  "          gh release edit \"\$TARGET_TAG\" --draft=false --repo \"\${GITHUB_REPOSITORY}\"

      - name: Attach release notes
        run: gh release upload \"\$TARGET_TAG\" NOTES.md"

# The draft is never published.
assert_rejected "draft-never-published" \
  "          gh release edit \"\$TARGET_TAG\" --draft=false --repo \"\${GITHUB_REPOSITORY}\"" \
  '          echo "release left as a draft"'

# release-please publishes the release itself, the v0.26.7 failure.
assert_config_rejected "release-please-publishes" RELEASE_PLEASE_CONFIG_PATH \
  .github/release-please-config.json \
  '      "draft": true,' \
  '      "draft": false,'

# Without a forced tag, a draft release creates no tag and release.yml never runs.
assert_config_rejected "release-please-no-tag" RELEASE_PLEASE_CONFIG_PATH \
  .github/release-please-config.json \
  '      "force-tag-creation": true,' \
  ''

# goreleaser creates a second release instead of filling the draft.
assert_config_rejected "goreleaser-ignores-draft" GORELEASER_CONFIG_PATH \
  .goreleaser.yml \
  '  use_existing_draft: true' \
  '  use_existing_draft: false'

# goreleaser publishes before provenance has run.
assert_config_rejected "goreleaser-publishes" GORELEASER_CONFIG_PATH \
  .goreleaser.yml \
  '  draft: true' \
  '  draft: false'

# The Windows zip stops carrying the installer, so nothing can register the service.
assert_config_rejected "windows-archive-without-installer" GORELEASER_CONFIG_PATH \
  .goreleaser.yml \
  '      - src: deploy/windows/install.ps1
        dst: install.ps1
' \
  ''

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
