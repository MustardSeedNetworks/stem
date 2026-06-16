// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"net/http"

	"golang.org/x/time/rate"

	apiratelimit "github.com/MustardSeedNetworks/stem/internal/api/ratelimit"
)

// RateLimiter is a type alias for the leaf package's type so existing callers
// in server.go, route.go, and test files need no changes in this slice.
type RateLimiter = apiratelimit.RateLimiter

// Rate-limiting constants forwarded from the leaf package.
const (
	AuthRateLimit   = apiratelimit.AuthRateLimit
	AuthBurstLimit  = apiratelimit.AuthBurstLimit
	APIRateLimit    = apiratelimit.APIRateLimit
	APIBurstLimit   = apiratelimit.APIBurstLimit
	CleanupInterval = apiratelimit.CleanupInterval
	VisitorTTL      = apiratelimit.VisitorTTL
	MaxVisitors     = apiratelimit.MaxVisitors
)

// NewRateLimiter creates a new rate limiter. Delegates to the leaf package.
func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	return apiratelimit.NewRateLimiter(r, burst)
}

// NewAuthRateLimiter creates a rate limiter for authentication endpoints.
// Delegates to the leaf package.
func NewAuthRateLimiter() *RateLimiter {
	return apiratelimit.NewAuthRateLimiter()
}

// NewAPIRateLimiter creates a rate limiter for standard API endpoints.
// Delegates to the leaf package.
func NewAPIRateLimiter() *RateLimiter {
	return apiratelimit.NewAPIRateLimiter()
}

// getClientIP extracts the client IP from the request. Delegates to the leaf
// package's exported ClientIP so handlers_recovery.go compiles without change.
func getClientIP(r *http.Request) string {
	return apiratelimit.ClientIP(r)
}
