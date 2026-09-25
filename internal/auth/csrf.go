// SPDX-License-Identifier: BUSL-1.1

package auth

import (
	"net/http"
	"strings"

	"github.com/MustardSeedNetworks/foundation/pkg/csrf"
)

// CSRF token configuration.
const (
	// CSRFHeaderName is the HTTP header name for CSRF tokens.
	CSRFHeaderName = "X-Csrf-Token"
)

// GetSessionIDFromRequest derives the CSRF session key from the request's
// authenticated JWT. The JWT is extracted the same way the auth middleware
// reads it (GetTokenFromRequest), then hashed via foundation's SessionKey so
// the bearer plaintext is never stored in the manager. Returns "" when the
// request carries no token. Exported for the CSRF token endpoint, which mints
// under this key, and for the route Registrar's CSRF layer, which validates
// under it.
func GetSessionIDFromRequest(r *http.Request) string {
	token, _ := GetTokenFromRequest(r)
	// Only a JWT-shaped bearer (a browser session token) is CSRF-relevant.
	// A malformed value or a non-JWT bearer (e.g. an API token, which a
	// cross-site attacker cannot set) gets no session key, so the request
	// passes through to the auth layer — which returns 401 for an invalid
	// token — instead of being 403'd for a missing CSRF token.
	const jwtMinParts = 2
	if len(strings.Split(token, ".")) < jwtMinParts {
		return ""
	}
	return csrf.SessionKey(token)
}
