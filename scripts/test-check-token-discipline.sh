#!/usr/bin/env bash
# test-check-token-discipline.sh — exercises the FAMILY_OPACITY_DIM rule in
# both directions.
#
# The rule's message is about a TEXT layer dimmed with opacity, where the label
# fails contrast while it is explaining itself. Its pattern used to match any
# `opacity-{0..60}` anywhere, and 13 of its 13 hits were things it does not
# mean: `disabled:` states on native controls (WCAG 1.4.3 exempts inactive
# components), opacity on SVG shapes, and `opacity-0` on a hidden element
# (#657). An advisory that can never reach zero stops being read.
#
# A rule nobody tests is a rule nobody knows the shape of, which is how it
# drifted from its own message in the first place.
set -euo pipefail

cd "$(dirname "$0")/.."

readonly PATTERN='(?<![-:\w])(?!(?<=className=")opacity-[0-9]+")opacity-(5|10|20|25|30|40|50|60)\b'

# Keep this in step with scripts/check-token-discipline.sh. A drift here is a
# test that passes while the gate does something else.
if ! grep -qF "$PATTERN" scripts/check-token-discipline.sh; then
  echo "FAIL: this test's pattern is not the one check-token-discipline.sh uses" >&2
  exit 1
fi

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

cat > "$work/cases.tsx" <<'EOF'
<span className="text-text-muted opacity-50">DIM_TRAILING</span>
<span className="opacity-50 text-text-muted">DIM_LEADING</span>
<circle className="opacity-25" cx="12" cy="12" r="10" />
<input className="border disabled:opacity-50" />
<button className="hover:opacity-60">x</button>
<div className="opacity-0 transition-opacity" />
EOF

hits=$(grep -oP "$PATTERN" "$work/cases.tsx" | wc -l | tr -d ' ')
matched_lines=$(grep -nP "$PATTERN" "$work/cases.tsx" || true)

fail=0
expect_flagged() {
  if ! grep -q "$1" <<<"$matched_lines"; then
    echo "FAIL: expected the rule to flag: $1"
    fail=1
  fi
}
expect_ignored() {
  if grep -q "$1" <<<"$matched_lines"; then
    echo "FAIL: expected the rule to ignore: $1"
    fail=1
  fi
}

# A dimmed text layer, whichever end of the class list the utility sits at.
expect_flagged DIM_TRAILING
expect_flagged DIM_LEADING

# Not a dimmed text layer.
expect_ignored 'className="opacity-25"'   # SVG shape, carries no text
expect_ignored 'disabled:opacity-50'      # inactive component, WCAG-exempt
expect_ignored 'hover:opacity-60'         # pointer affordance, transient
expect_ignored 'opacity-0'                # hidden, not dimmed

if [ "$hits" -ne 2 ]; then
  echo "FAIL: expected exactly 2 matches, got $hits"
  echo "$matched_lines"
  fail=1
fi

if [ "$fail" -ne 0 ]; then
  echo "test-check-token-discipline: FAILED"
  exit 1
fi

echo "test-check-token-discipline: PASS (2 real dims flagged, 4 non-dims ignored)"
