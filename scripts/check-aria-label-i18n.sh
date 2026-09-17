#!/usr/bin/env bash
# check-aria-label-i18n.sh — accessible names must come from the locale files.
#
# The *.i18n.test.tsx suites query visible text with getByText, which cannot
# see an attribute value. So an `aria-label="Close modal"` is invisible to
# every i18n test the repo has, and twelve of them shipped English into the
# `es` locale unnoticed (#1252) — including the language switcher itself,
# which announced "Select language" to a Spanish screen-reader user.
#
# The rule: an aria-label in application code is an expression, never a bare
# English string literal. `aria-label={t('accessibility.closeModal')}` passes;
# `aria-label="Close modal"` fails.
#
# Scope: *.tsx under ui/src/, EXCEPT tests and stories — a fixture's label is
# not shipped copy. A lowercase literal (aria-label="eth0") is out of scope by
# the same reasoning the issue used: it is an identifier, not a sentence.
#
# Run locally: scripts/check-aria-label-i18n.sh

set -uo pipefail

if [ -d "ui/src" ]; then
  TARGET="ui/src"
elif [ -d "src" ] && [ -f "package.json" ]; then
  TARGET="src"
else
  echo "ERROR: cannot find ui/src — run from repo root or ui/ directory" >&2
  exit 2
fi

readonly PATTERN='aria-label="[A-Z]'
readonly EXCLUDE_RE='\.(test|spec|stories|mock)\.tsx:'

hits=$(grep -rEn --include='*.tsx' "$PATTERN" "$TARGET" 2>/dev/null | grep -vE "$EXCLUDE_RE")

if [ -n "$hits" ]; then
  echo "FAIL: literal English aria-label(s) — route them through t() (#1252):" >&2
  echo "$hits" >&2
  exit 1
fi

echo "aria-label i18n gate: clean ($TARGET)"
