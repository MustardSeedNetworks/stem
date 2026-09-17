#!/usr/bin/env bash
# test-check-aria-label-i18n.sh — exercises check-aria-label-i18n.sh in both
# directions.
#
# A gate nobody tests is a gate nobody knows the shape of. This one is a grep,
# and a grep is exactly the kind of check that quietly stops matching: the
# repo it guards is at zero hits, so a pattern that broke would look
# indistinguishable from a clean tree.
set -euo pipefail

cd "$(dirname "$0")/.."

readonly GATE=scripts/check-aria-label-i18n.sh

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

mkdir -p "$work/src"
cp "$GATE" "$work/gate.sh"
cat > "$work/package.json" <<'EOF'
{ "name": "gate-fixture", "private": true }
EOF

run_gate() { (cd "$work" && ./gate.sh 2>&1); }

# --- RED: a literal English label must fail.
cat > "$work/src/Bad.tsx" <<'EOF'
export const Bad = () => <button type="button" aria-label="Close modal" />;
EOF
if run_gate >/dev/null 2>&1; then
  echo "FAIL: gate passed a literal English aria-label" >&2
  exit 1
fi
# pipefail would make `run_gate | grep` fail on the gate's own non-zero exit,
# so the output is captured first.
out=$(run_gate || true)
case "$out" in
  *Bad.tsx*) ;;
  *) echo "FAIL: gate did not name the offending file" >&2; exit 1 ;;
esac

# --- A test fixture's literal is out of scope.
mv "$work/src/Bad.tsx" "$work/src/Bad.test.tsx"
run_gate >/dev/null || { echo "FAIL: gate flagged a *.test.tsx fixture" >&2; exit 1; }
mv "$work/src/Bad.test.tsx" "$work/src/Bad.stories.tsx"
run_gate >/dev/null || { echo "FAIL: gate flagged a *.stories.tsx fixture" >&2; exit 1; }
rm "$work/src/Bad.stories.tsx"

# --- GREEN: the translated form passes.
cat > "$work/src/Good.tsx" <<'EOF'
export const Good = () => <button type="button" aria-label={t('accessibility.closeModal')} />;
EOF
run_gate >/dev/null || { echo "FAIL: gate rejected a t() aria-label" >&2; exit 1; }

# --- An identifier, not a sentence, stays out of scope.
cat > "$work/src/Iface.tsx" <<'EOF'
export const Iface = () => <span aria-label="eth0" />;
EOF
run_gate >/dev/null || { echo "FAIL: gate flagged a lowercase identifier label" >&2; exit 1; }

echo "test-check-aria-label-i18n: all cases pass"
