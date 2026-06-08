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
