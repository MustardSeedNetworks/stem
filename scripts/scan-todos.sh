#!/usr/bin/env bash
# scan-todos.sh — find real TODO/FIXME/HACK/XXX markers in source.
#
# Prints one `path:line:text` per marker to stdout; exit status is always 0.
# Used by .github/workflows/todo-tracker.yml to build the weekly summary.
#
# WHY THIS IS A SCRIPT AND NOT A GREP IN THE WORKFLOW
#
# The workflow used to run `grep -rn "TODO\|FIXME\|HACK\|XXX"`, which matches
# those letters ANYWHERE on a line. Every one of the 15 items it reported was
# a false positive:
#
#   XXXX-XXXX-XXXX-XXXX   license-key format placeholders (help text, CLI
#                         output, an input placeholder) — 9 hits
#   context.TODO()        the standard Go idiom for an unset context — 5 hits
#   XXX-XX-XXXX           an SSN format named in a redaction comment — 1 hit
#
# Four weeks of summary issues (#554, #650, #747, #926) all reported the same
# 15 non-TODOs, so the signal was zero and the issues were never actionable.
#
# A real marker is a comment whose FIRST token is the marker. That single
# constraint removes all three false-positive families: `context.TODO()` has no
# comment opener, `XXXX` has no word boundary after `XXX`, and the SSN comment
# opens with prose. It is deliberately conservative — a marker buried mid-prose
# is missed, which costs a reminder; a false positive costs the gate's
# credibility, which is what actually happened here.
#
# Run `scripts/test-scan-todos.sh` to exercise both directions.
set -euo pipefail

cd "$(dirname "$0")/.."

# Comment opener, optional space, marker as a whole word. `*` is anchored to
# line start so a multiplication does not read as a block-comment continuation.
readonly PATTERN='(//|/\*|^[[:space:]]*\*|#)[[:space:]]*(TODO|FIXME|HACK|XXX)\b'

rg --no-heading --line-number --color=never \
  --glob '*.go' --glob '*.ts' --glob '*.tsx' --glob '*.js' --glob '*.jsx' \
  --glob '!vendor/**' --glob '!node_modules/**' --glob '!dist/**' \
  --glob '!**/scan-todos.sh' \
  -e "$PATTERN" . 2>/dev/null | sed 's|^\./||' || true
