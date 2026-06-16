// SPDX-License-Identifier: BUSL-1.1

package api

// route.go is the capability registry: API routes are declared as data and a
// single register() composes their per-route middleware in one canonical order.
// This replaces hand-wrapping each route at registration (s.handleAuthRateLimited
// / s.handleRateLimited / raw s.mux.Handle), where the wrapper nesting was applied
// inconsistently and could be forgotten — the regression that left
// POST /api/v1/reflector/config unauthenticated (fixed in #398). With the
// registry, a route cannot be installed without its policy, and
// scripts/check-route-policy.sh enforces that every API route goes through here.

import (
	"net/http"
	"slices"
	"strings"

	"github.com/MustardSeedNetworks/stem/internal/api/ratelimit"
)

// route declares an API route and its per-route policy. CSRF is enforced
// globally (CSRFMiddleware), so it is NOT part of this policy. Authentication is
// per-route (the auth flag) because stem mixes pre-session endpoints (login,
// setup, recovery, license, module list) with authenticated ones.
type route struct {
	// path is the full request path (e.g. "/api/v1/mode").
	path string
	// handler is the terminal handler for the route.
	handler http.HandlerFunc
	// methods, when non-empty, gates the route to these HTTP methods (405 + an
	// Allow header otherwise). Empty preserves the handler's own method dispatch.
	methods []string
	// maxBodyBytes caps the request body (DoS guard) via http.MaxBytesReader,
	// applied before the handler reads. 0 means the default (maxRequestBodySize).
	// Set explicitly only for a route that must accept a larger or tighter body.
	maxBodyBytes int64
	// auth requires a valid token (authMiddleware) before the handler runs.
	auth bool
	// limiter is the rate limiter the route is wrapped in — s.authLimiter for
	// auth-sensitive endpoints (5/min), s.apiLimiter for the rest. nil = none
	// (long-lived SSE / unauthenticated introspection).
	limiter *ratelimit.RateLimiter
}

// methodGate rejects any method outside allowed with 405 + an Allow header.
func (s *Server) methodGate(allowed []string, next http.HandlerFunc) http.HandlerFunc {
	allowHeader := strings.Join(allowed, ", ")
	return func(w http.ResponseWriter, r *http.Request) {
		if slices.Contains(allowed, r.Method) {
			next(w, r)
			return
		}
		w.Header().Set("Allow", allowHeader)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// bodyLimited caps the request body before the handler reads it.
func bodyLimited(limit int64, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, limit)
		next(w, r)
	}
}

// register installs rt on the mux, composing middleware in ONE canonical order
// for every route: rateLimit (outermost) → auth → methodGate → bodyLimit →
// handler (body cap closest to the handler so r.Body is capped before any read).
// Composing here rather than at each call site makes the policy declarative and
// uniform, records it in the manifest for /__capabilities, and is the single
// choke point the route-policy CI gate enforces.
func (s *Server) register(rt route) {
	if rt.maxBodyBytes == 0 {
		rt.maxBodyBytes = maxRequestBodySize
	}
	s.routeManifest = append(s.routeManifest, rt)

	h := rt.handler
	// bodyLimit innermost so r.Body is capped before the handler reads.
	h = bodyLimited(rt.maxBodyBytes, h)
	if len(rt.methods) > 0 {
		h = s.methodGate(rt.methods, h)
	}

	var handler http.Handler = h
	if rt.auth {
		handler = s.authMiddleware(h)
	}
	if rt.limiter != nil {
		handler = rt.limiter.Middleware(handler)
	}
	s.mux.Handle(rt.path, handler)
}

// registerAll installs a slice of routes through register().
func (s *Server) registerAll(routes []route) {
	for _, rt := range routes {
		s.register(rt)
	}
}

// routePolicyView is the JSON projection of a route's policy for the
// /__capabilities manifest (the handler func itself is not exposed).
type routePolicyView struct {
	Path         string   `json:"path"`
	Methods      []string `json:"methods,omitempty"`
	MaxBodyBytes int64    `json:"maxBodyBytes,omitempty"`
	Auth         bool     `json:"auth"`
	RateLimited  bool     `json:"rateLimited"`
}

// handleRoutePolicyManifest serves the route-policy manifest: every route
// registered through register() with its auth + rate-limit policy. No auth —
// like /__version it is a deployment/audit introspection surface. (Distinct from
// /api/v1/capabilities, which reports platform feature availability.)
func (s *Server) handleRoutePolicyManifest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	views := make([]routePolicyView, 0, len(s.routeManifest))
	for _, rt := range s.routeManifest {
		views = append(views, routePolicyView{
			Path:         rt.path,
			Methods:      rt.methods,
			MaxBodyBytes: rt.maxBodyBytes,
			Auth:         rt.auth,
			RateLimited:  rt.limiter != nil,
		})
	}
	writeJSON(w, views)
}
