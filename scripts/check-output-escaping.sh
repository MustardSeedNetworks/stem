#!/usr/bin/env bash
# check-output-escaping.sh — output-escaping / XSS regression gate (#343).
#
# Audited 2026-05-28: every fmt.Fprintf(w, ...) site in this repo is
# either CLI display (internal/help, ANSI -> stdout) or a literal SSE
# wire-format comment in internal/api. No site renders user-supplied
# data as HTML. This gate keeps it that way.
#
# Two checks:
#   1. No raw innerHTML injection in ui/src (the React XSS vector).
#   2. No fmt.Fprintf to an http.ResponseWriter in internal/api/ that
#      interpolates a value (%s/%v/%q/%d) — HTTP responses must use the
#      JSON encoder or html/template, never raw format strings. SSE
#      comment lines (literal, no format verb) are fine and don't match.
#
# Run locally: scripts/check-output-escaping.sh

set -uo pipefail

FAIL=0

# Pattern written as a regex char class so this gate file does not itself
# contain the contiguous banned token.
INNER_HTML_RE='dangerously[S]etInnerHTML'

# --- Check 1: raw innerHTML injection in UI app code -------------------
UI_DIR=""
if [ -d "ui/src" ]; then
  UI_DIR="ui/src"
elif [ -d "src" ] && [ -f "package.json" ]; then
  UI_DIR="src"
fi

if [ -n "$UI_DIR" ]; then
  DANGER=$(grep -rEn "$INNER_HTML_RE" "$UI_DIR" \
    --include='*.tsx' --include='*.ts' 2>/dev/null \
    | grep -vE '\.(test|spec|stories)\.(ts|tsx):' || true)
  if [ -n "$DANGER" ]; then
    echo "============================================================"
    echo "[XSS] raw innerHTML injection found in UI app code:"
    echo "$DANGER"
    echo "Use plain text/JSX. If HTML rendering is unavoidable, sanitize"
    echo "with a vetted sanitizer (e.g. DOMPurify) and justify it in review."
    echo ""
    FAIL=1
  fi
fi

# --- Check 2: value-interpolating Fprintf to ResponseWriter in api/ ----
if [ -d "internal/api" ]; then
  # Match fmt.Fprintf(w, "...%s..." — a format verb interpolated into an
  # HTTP response. Literal SSE comments (": heartbeat\n\n") have no verb
  # and are not flagged.
  HTTP_FMT=$(grep -rEn 'fmt\.Fprintf\(w,[^)]*%[svqd]' internal/api 2>/dev/null \
    | grep -v '_test.go' || true)
  if [ -n "$HTTP_FMT" ]; then
    echo "============================================================"
    echo "[XSS] value-interpolating fmt.Fprintf(w, ...) in internal/api —"
    echo "HTTP responses must use json.NewEncoder(w).Encode(...) or"
    echo "html/template, not raw format strings (auto-escaping is bypassed):"
    echo "$HTTP_FMT"
    echo ""
    echo "If this is a deliberately-safe wire format (e.g. SSE with a"
    echo "server-controlled enum), refactor to a literal or add an"
    echo "explicit allow-list entry here with justification."
    echo ""
    FAIL=1
  fi
fi

if [ "$FAIL" -ne 0 ]; then
  echo "FAIL: output-escaping gate (#343)."
  exit 1
fi

echo "OK: no raw innerHTML injection and no value-interpolating Fprintf in internal/api."
