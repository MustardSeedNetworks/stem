// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"net/http"
)

// securityHeadersMiddleware adds security headers to all responses.
// These headers protect against common web vulnerabilities:
//   - HSTS: Enforces HTTPS connections (only on TLS)
//   - X-Frame-Options: Prevents clickjacking attacks
//   - X-Content-Type-Options: Prevents MIME-type sniffing
//   - X-XSS-Protection: Legacy XSS filter (for older browsers)
//   - Referrer-Policy: Controls referrer information leakage
//   - Content-Security-Policy: Restricts resource loading
//   - Permissions-Policy: Restricts browser features
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// HSTS: Only set on TLS connections to avoid redirect loops.
		// max-age=31536000 = 1 year, includeSubDomains for full coverage.
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// X-Frame-Options: DENY prevents any framing (clickjacking protection).
		w.Header().Set("X-Frame-Options", "DENY")

		// X-Content-Type-Options: Prevents MIME-type sniffing.
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection: Legacy XSS filter for older browsers.
		// Modern browsers rely on CSP instead.
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy: Only send origin on cross-origin requests.
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content-Security-Policy: Strict policy without 'unsafe-inline' (ported from Seed).
		// - default-src 'self': Only load resources from same origin.
		// - script-src 'self': Only allow same-origin scripts.
		// - style-src 'self': Only same-origin styles (no unsafe-inline for XSS protection).
		// - img-src 'self' data:: Allow same-origin images and data URIs.
		// - connect-src 'self': API calls and SSE connections.
		// - font-src 'self': Only same-origin fonts.
		// - object-src 'none': Block plugins (Flash, etc.).
		// - base-uri 'self': Restrict <base> tag.
		// - form-action 'self': Forms can only submit to same origin.
		// - frame-ancestors 'none': Prevent framing (like X-Frame-Options).
		csp := "default-src 'self'; " +
			"script-src 'self'; " +
			"style-src 'self'; " +
			"img-src 'self' data:; " +
			"connect-src 'self'; " +
			"font-src 'self'; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'; " +
			"frame-ancestors 'none'"
		w.Header().Set("Content-Security-Policy", csp)

		// Permissions-Policy: Restrict browser features.
		// Disable features that aren't needed for a network testing tool.
		permissions := "camera=(), " +
			"microphone=(), " +
			"geolocation=(), " +
			"payment=(), " +
			"usb=()"
		w.Header().Set("Permissions-Policy", permissions)

		next.ServeHTTP(w, r)
	})
}
