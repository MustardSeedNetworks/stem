#!/usr/bin/env bash
# check-route-policy.sh — capability-registry enforcement gate.
#
# Server side: every route goes through foundation's route.Registrar, which
# composes each route's policy (rate limit, auth, method gate, CSRF, body cap)
# in one canonical order. Hand-wrapping a route on a mux of our own is how a
# mutating route silently ships without authentication — the regression that
# left POST /api/v1/reflector/config open (fixed in #398). The rule is
# foundation's, run from the pinned module copy so the gate and the registrar
# it enforces move on one version.
#
# Run locally: scripts/check-route-policy.sh
set -euo pipefail

route_pkg_dir=$(go list -f '{{.Dir}}' github.com/MustardSeedNetworks/foundation/pkg/httpserver/route)
bash "$route_pkg_dir/check-route-policy.sh" internal/api

# --- Browser side: every API call carries the /v1 prefix ---------------------
#
# The same policy seen from the client. `apiVersionMiddleware` only stamps a
# response header — it rewrites nothing — so an unversioned path is not a 404:
# it falls through to the SPA handler and returns index.html with HTTP 200.
# `response.ok` is then true and the failure surfaces, if at all, as a JSON
# parse error inside a catch. That is exactly how `/api/modules` shipped a
# hardcoded module catalogue and how `/api/license*` left the License panel
# unable to activate anything (#1247) — twice, silently.
#
# Comment lines are skipped so prose about the historical paths stays legible,
# and so are test files: they assert against the production code this gate
# already covers (ModuleSelector.test.tsx pins that the unversioned path is NOT
# requested, which has to name it).
UI_SRC="ui/src"

ui_calls=$(grep -rnE "['\"\`]/api/(v[^1]|[^v])" "$UI_SRC" \
	--include='*.ts' --include='*.tsx' \
	| grep -v '\.test\.tsx\?:' \
	| grep -vE ':[0-9]+: *(\*|//|/\*)' || true)

if [[ -n "$ui_calls" ]]; then
	echo "❌ Route-policy gate: a UI request omits the /v1 prefix. An"
	echo "   unversioned /api path is served the SPA's index.html with HTTP"
	echo "   200, so the call silently succeeds and returns HTML (#1247)."
	echo "   Use /api/v1/... ."
	echo ""
	echo "$ui_calls"
	exit 1
fi

echo "✓ Route-policy gate: every UI API path is versioned."
