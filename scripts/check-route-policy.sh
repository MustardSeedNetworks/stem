#!/usr/bin/env bash
# check-route-policy.sh — capability-registry enforcement gate.
#
# Every API route MUST be registered through the capability registry
# (register / registerAll in internal/api/route.go), which composes its
# per-route policy — rate limit, authentication — in ONE canonical order.
# Hand-wrapping a route directly on the mux (s.mux.Handle*/s.handle of an
# "/api/..." literal) bypasses that composition and is how a mutating route
# silently ships without authentication — the regression that left
# POST /api/v1/reflector/config open (fixed in #398).
#
# This gate fails if any "/api/..." path is registered directly instead of via
# register(). register() installs routes with a variable path (rt.path), so it
# never matches; only direct literal registrations do. Non-/api introspection
# endpoints (/__version, /__capabilities, /health/*) are intentionally direct
# and are not /api/, so they are not matched.
#
# Run locally: scripts/check-route-policy.sh
set -euo pipefail

API_DIR="internal/api"

violations=$(grep -rnE 's\.(mux\.Handle(Func)?|handle)\("/api/' "$API_DIR"/*.go \
	| grep -v '_test.go' || true)

if [[ -n "$violations" ]]; then
	echo "❌ Route-policy gate: API routes must be registered through the"
	echo "   capability registry (registerAll/register in route.go), not via a"
	echo "   raw s.mux.Handle*/s.handle of an \"/api/...\" literal. A direct"
	echo "   registration skips the auth+rate-limit composition — the #398"
	echo "   reflector regression. Add a route{} entry to setupRoutes instead."
	echo ""
	echo "$violations"
	exit 1
fi

echo "✓ Route-policy gate: all /api routes go through the capability registry."

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
