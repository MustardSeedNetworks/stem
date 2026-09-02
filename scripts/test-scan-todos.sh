#!/usr/bin/env bash
# test-scan-todos.sh — proves scan-todos.sh catches real markers and ignores
# the false-positive families that made four weekly summary issues worthless.
#
# It runs the real script against a throwaway repository rather than
# re-implementing the pattern, so a change to the scanner is actually exercised.
set -euo pipefail

cd "$(dirname "$0")/.."
readonly SCANNER="$PWD/scripts/scan-todos.sh"

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

mkdir -p "$work/scripts"
cp "$SCANNER" "$work/scripts/scan-todos.sh"

cat > "$work/should_match.go" <<'EOF'
package p

// TODO: wire the reflector shutdown path
// FIXME the retry budget is guessed
//HACK: works around a driver bug
/* XXX this allocation is unbounded */

/*
 * TODO: block-comment continuation form
 */
func f() {}
EOF

cat > "$work/should_not_match.go" <<'EOF'
package p

import "context"

const keyFormat = "XXXX-XXXX-XXXX-XXXX"

// Social Security Numbers (US) - XXX-XX-XXXX format
func f() {
	_ = context.TODO()
	_, _ = keyFormat, "  stem license --activate XXXX-XXXX-XXXX-XXXX"
}
EOF

cat > "$work/should_not_match.tsx" <<'EOF'
export const F = () => <input placeholder="XXXX-XXXX-XXXX-XXXX" />;
EOF

git -C "$work" init --quiet
git -C "$work" add -A
out=$("$work/scripts/scan-todos.sh")

fail=0
expect_hit() {
  if ! printf '%s\n' "$out" | grep -qF "$1"; then
    echo "FAIL: expected a hit for: $1"
    fail=1
  fi
}
expect_miss() {
  if printf '%s\n' "$out" | grep -qF "$1"; then
    echo "FAIL: expected NO hit for: $1"
    fail=1
  fi
}

expect_hit 'TODO: wire the reflector'
expect_hit 'FIXME the retry budget'
expect_hit 'HACK: works around'
expect_hit 'XXX this allocation'
expect_hit 'TODO: block-comment continuation'

expect_miss 'context.TODO()'
expect_miss 'XXXX-XXXX-XXXX-XXXX'
expect_miss 'Social Security Numbers'

if [ "$fail" -ne 0 ]; then
  echo "--- scanner output was ---"
  printf '%s\n' "$out"
  echo "test-scan-todos: FAILED"
  exit 1
fi

echo "test-scan-todos: PASS (5 real markers found, 3 false-positive families ignored)"
