// SPDX-License-Identifier: BUSL-1.1

package api

// route.go wires stem's policy into the fleet's capability registry
// (foundation pkg/httpserver/route). Routes are declared as data in routes.go
// and the shared Registrar composes each one's policy — limiter,
// auth, method gate, CSRF, body cap — in its one canonical order, wrapped in
// request ID, access log and panic recovery. This file supplies only what the
// order does not decide: stem's error envelope, its auth middleware, its CSRF
// session key and its two limiters. Hand-wrapping routes is how
// POST /api/v1/reflector/config once shipped unauthenticated (#398);
// scripts/check-route-policy.sh keeps every route on the Registrar.

import (
	"net/http"
	"strings"

	"github.com/MustardSeedNetworks/foundation/pkg/csrf"
	"github.com/MustardSeedNetworks/foundation/pkg/httpserver/route"

	"github.com/MustardSeedNetworks/stem/internal/api/ratelimit"
	"github.com/MustardSeedNetworks/stem/internal/auth"
	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// Limiter names a route.Route.Limiter can select.
const (
	limitAuth = "auth" // authLimiter: credential and MFA endpoints (5/min)
	limitAPI  = "api"  // apiLimiter: every other metered route (100/min)
)

// newRegistrar builds the Registrar over this server's policy.
// It reads the limiters and the CSRF manager when called, so build it after
// they are set.
func (s *Server) newRegistrar() *route.Registrar {
	return route.New(route.Config{
		Error:        registrarError,
		MaxBodyBytes: maxRequestBodySize,
		Logger:       logging.Get(),
		Auth:         s.authMiddleware,
		CSRF:         s.csrfManager,
		SessionKey:   csrfSessionKey,
		Limiters: map[string]route.Middleware{
			limitAuth: s.authLimiter.Middleware,
			limitAPI:  s.apiLimiter.Middleware,
		},
	})
}

// csrfSessionKey keys CSRF tokens by the session the auth layer reads — the
// access-token cookie first, then a bearer header — which is the key
// /api/v1/auth/csrf-token mints under. A request with no session has nothing
// ambient to forge and passes on; behind Auth it has already been refused.
func csrfSessionKey(r *http.Request) (string, bool) {
	key := auth.GetSessionIDFromRequest(r)
	return key, key != ""
}

// registrarError renders the Registrar's own refusals (405, CSRF, a recovered
// panic) in stem's JSON envelope.
func registrarError(w http.ResponseWriter, _ *http.Request, status int, code, message string) {
	WriteError(w, NewError(status, registrarErrorCode(code), message))
}

// registrarErrorCode maps a Registrar code onto stem's vocabulary. The three
// CSRF causes keep their distinct codes, uppercased.
func registrarErrorCode(code string) ErrorCode {
	switch code {
	case "method_not_allowed":
		return ErrCodeMethodNotAllowed
	case "internal_server_error", "csrf_unavailable":
		return ErrCodeInternalError
	default:
		return ErrorCode(strings.ToUpper(code))
	}
}

// RouteManifest returns every route the registry declares, in registration
// order, without starting a daemon: registration only composes closures, so a
// Server carrying just its limiters and CSRF manager registers the full set.
// It is what cmd/stem-openapi documents.
func RouteManifest() []route.Policy {
	s := &Server{
		authLimiter: ratelimit.NewAuthRateLimiter(nil),
		apiLimiter:  ratelimit.NewAPIRateLimiter(nil),
		csrfManager: csrf.NewManager(),
	}
	defer s.authLimiter.Stop()
	defer s.apiLimiter.Stop()
	defer s.csrfManager.Stop()
	reg := s.newRegistrar()
	reg.RegisterAll(s.routes(reg))
	return reg.Policies()
}
