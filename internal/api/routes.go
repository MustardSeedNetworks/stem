// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"io/fs"
	"net/http"
	"slices"

	"github.com/MustardSeedNetworks/foundation/pkg/httpserver/route"

	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// routes is the capability registry: every route stem serves and its policy.
// The Registrar composes each one in the canonical order, so a route cannot
// ship without the policy declared here, and the manifest (/__capabilities)
// and docs/openapi.yaml are generated from it. CSRF is declared on every route
// that takes a state-changing method, except the pre-session sign-in steps,
// which cannot carry a token; a request with no session passes CSRF and meets
// Auth instead.
func (s *Server) routes(reg *route.Registrar) []route.Route {
	get := []string{http.MethodGet}
	return slices.Concat(
		[]route.Route{
			// Infrastructure: unversioned deployment and audit introspection.
			{Path: "/__version", Handler: s.handleBuildVersion, Methods: get},
			{Path: "/health/live", Handler: s.handleHealthLive, Methods: get},
			{Path: "/health/ready", Handler: s.handleHealthReady, Methods: get},
			{Path: "/__capabilities", Handler: reg.ServeManifest, Methods: get},
		},
		s.apiRoutes(),
		s.authRoutes(),
		[]route.Route{
			// The embedded UI, falling back to index.html so client-side
			// routes survive a refresh. Hidden: a catch-all is not an API
			// operation.
			{Path: "/", Handler: uiHandler(), Hidden: true},
		},
	)
}

// apiRoutes are the product routes: status, settings, runs, the reflector,
// licensing and the module catalogue.
func (s *Server) apiRoutes() []route.Route {
	get := []string{http.MethodGet}
	post := []string{http.MethodPost}
	getPost := []string{http.MethodGet, http.MethodPost}
	return []route.Route{
		// Health & status.
		{Path: "/api/v1/health", Handler: s.handleHealth, Methods: get},
		// Telemetry (#340).
		{Path: "/api/v1/stats", Handler: s.handleStats, Methods: get, Auth: true, Limiter: limitAPI},
		// Platform capabilities — UI calls this pre-login to gate CGO-less builds.
		{Path: "/api/v1/capabilities", Handler: s.handleCapabilities, Methods: get},
		// Interfaces — discloses the host NIC inventory (#340).
		{Path: "/api/v1/interfaces", Handler: s.handleInterfaces, Methods: get, Auth: true, Limiter: limitAPI},
		// Settings + mode — POST writes config / flips the operating role.
		{
			Path:    "/api/v1/settings",
			Handler: s.handleSettings,
			Methods: getPost,
			Auth:    true,
			CSRF:    true,
			Limiter: limitAPI,
		},
		{Path: "/api/v1/mode", Handler: s.handleMode, Methods: getPost, Auth: true, CSRF: true, Limiter: limitAPI},
		// SSE stream — long-lived, unauthenticated (observational frames only).
		// If privileged frames are ever published here, set Auth.
		{Path: "/api/v1/events", Handler: s.handleSSEEvents, Methods: get},
		// Test execution.
		{
			Path:    "/api/v1/test/start",
			Handler: s.handleTestStart,
			Methods: post,
			Auth:    true,
			CSRF:    true,
			Limiter: limitAPI,
		},
		{
			Path:    "/api/v1/test/stop",
			Handler: s.handleTestStop,
			Methods: post,
			Auth:    true,
			CSRF:    true,
			Limiter: limitAPI,
		},
		{Path: "/api/v1/test/result", Handler: s.handleTestResult, Methods: get, Auth: true, Limiter: limitAPI},
		// Reflector — reconfigures/inspects the dataplane, requires auth (#398).
		{
			Path:    "/api/v1/reflector/config",
			Handler: s.handleReflectorConfig,
			Methods: getPost,
			Auth:    true,
			CSRF:    true,
			Limiter: limitAPI,
		},
		{Path: "/api/v1/reflector/stats", Handler: s.handleReflectorStats, Methods: get, Auth: true, Limiter: limitAPI},
		// License — entitlement state, so every route requires a session (#1317).
		// ADR-0009's amendment records why the trial route's pre-session
		// exemption was withdrawn.
		{
			Path:    "/api/v1/license",
			Handler: s.handleLicense,
			Methods: []string{http.MethodGet, http.MethodDelete}, Auth: true, CSRF: true, Limiter: limitAPI,
		},
		{
			Path:    "/api/v1/license/activate",
			Handler: s.handleLicenseActivate,
			Methods: post,
			Auth:    true,
			CSRF:    true,
			Limiter: limitAPI,
		},
		{
			Path:    "/api/v1/license/trial",
			Handler: s.handleLicenseTrial,
			Methods: getPost,
			Auth:    true,
			CSRF:    true,
			Limiter: limitAPI,
		},
		// Modules (public catalog).
		{Path: "/api/v1/modules", Handler: s.handleModules, Methods: get, Limiter: limitAPI},
		{Path: "/api/v1/modules/", Handler: s.handleModuleByName, Methods: get, Limiter: limitAPI},
	}
}

// authRoutes are sign-in, session, second-factor, first-run setup and
// password recovery.
func (s *Server) authRoutes() []route.Route {
	get := []string{http.MethodGet}
	post := []string{http.MethodPost}
	return []route.Route{
		// Authentication (strict authLimiter). Login is pre-session: the
		// credential is the proof of intent and there is no token to carry.
		{Path: "/api/v1/auth/login", Handler: s.loginWithMFAGate, Methods: post, Limiter: limitAuth},
		{Path: "/api/v1/auth/logout", Handler: s.handleAuthLogout, Methods: post, CSRF: true, Limiter: limitAPI},
		// refresh is checked only while the access token is still presented;
		// its normal case has none, and SameSite=Strict cookies block the
		// browser CSRF vector — accepted defense-in-depth edge, see ADR-0009.
		{Path: "/api/v1/auth/refresh", Handler: s.handleAuthRefresh, Methods: post, CSRF: true, Limiter: limitAuth},
		{Path: "/api/v1/auth/csrf-token", Handler: s.handleAuthCSRF, Methods: get, Auth: true, Limiter: limitAPI},
		// MFA — TOTP management requires auth; the login finisher does not (it
		// presents an mfa_token from the password stage as proof of intent, and
		// is pre-session in the same way login is).
		{
			Path:    "/api/v1/auth/totp/setup",
			Handler: s.handleTOTPSetup,
			Methods: post,
			Auth:    true,
			CSRF:    true,
			Limiter: limitAuth,
		},
		{
			Path:    "/api/v1/auth/totp/verify",
			Handler: s.handleTOTPVerify,
			Methods: post,
			Auth:    true,
			CSRF:    true,
			Limiter: limitAuth,
		},
		{
			Path:    "/api/v1/auth/totp/disable",
			Handler: s.handleTOTPDisable,
			Methods: post,
			Auth:    true,
			CSRF:    true,
			Limiter: limitAuth,
		},
		{Path: "/api/v1/auth/login/totp", Handler: s.handleLoginTOTP, Methods: post, Limiter: limitAuth},
		{Path: "/api/v1/auth/mfa/status", Handler: s.handleMFAStatus, Methods: get, Auth: true, Limiter: limitAPI},
		// WebAuthn — register requires auth; login does not (the assertion
		// proves identity, and the caller has no session yet).
		{
			Path:    "/api/v1/auth/webauthn/register/begin",
			Handler: s.handleWebAuthnRegisterBegin,
			Methods: post, Auth: true, CSRF: true, Limiter: limitAuth,
		},
		{
			Path:    "/api/v1/auth/webauthn/register/finish",
			Handler: s.handleWebAuthnRegisterFinish,
			Methods: post, Auth: true, CSRF: true, Limiter: limitAuth,
		},
		{
			Path:    "/api/v1/auth/webauthn/login/begin",
			Handler: s.handleWebAuthnLoginBegin,
			Methods: post,
			Limiter: limitAuth,
		},
		{
			Path:    "/api/v1/auth/webauthn/login/finish",
			Handler: s.handleWebAuthnLoginFinish,
			Methods: post,
			Limiter: limitAuth,
		},
		// First-time setup (pre-session, no auth).
		{Path: "/api/v1/setup/status", Handler: s.handleSetupStatus, Methods: get, Limiter: limitAPI},
		{Path: "/api/v1/setup/complete", Handler: s.handleSetupComplete, Methods: post, CSRF: true, Limiter: limitAuth},
		// Password recovery (pre-session, no auth).
		{Path: "/api/v1/recovery/status", Handler: s.handleRecoveryStatus, Methods: get, Limiter: limitAPI},
		{
			Path:    "/api/v1/recovery/complete",
			Handler: s.handleRecoveryComplete,
			Methods: post,
			CSRF:    true,
			Limiter: limitAuth,
		},
		{Path: "/api/v1/recovery/instructions", Handler: s.handleRecoveryInstructions, Methods: get, Limiter: limitAPI},
	}
}

// uiHandler serves the embedded UI, or a page saying it was not built.
func uiHandler() http.HandlerFunc {
	staticFS, err := fs.Sub(staticFiles, "ui")
	if err != nil {
		logging.Warn("Could not load embedded UI", "error", err)
		return serveFallbackUIPage
	}
	return spaFallbackHandler(staticFS)
}
