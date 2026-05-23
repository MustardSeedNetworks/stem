// SPDX-License-Identifier: BUSL-1.1

package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Cookie names for authentication tokens.
const (
	// CookieNameAccess is the name of the access token cookie.
	CookieNameAccess = "stem_access"

	// CookieNameRefresh is the name of the refresh token cookie.
	CookieNameRefresh = "stem_refresh"
)

// CookieConfig holds cookie scope settings.
//
// Security attributes (Secure, HttpOnly, SameSite) are hardcoded on every
// auth cookie because the daemon enforces HTTPS for all auth endpoints
// (HTTP listener redirects to HTTPS; no auth flow ever runs over plain
// HTTP). They are intentionally NOT configurable.
type CookieConfig struct {
	// Domain sets the cookie domain
	Domain string

	// Path sets the cookie path
	Path string
}

// DefaultCookieConfig returns the standard cookie scope for stem.
func DefaultCookieConfig() CookieConfig {
	return CookieConfig{
		Domain: "", // Current domain
		Path:   "/",
	}
}

// newAuthCookie builds an http.Cookie with stem's hardcoded auth-cookie
// security baseline. Centralising the literals here keeps every auth
// cookie identical (and lets gosec G124 see Secure/HttpOnly/SameSite are
// all set unconditionally).
func newAuthCookie(name, value string, expires time.Time, maxAge int, config CookieConfig) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     config.Path,
		Domain:   config.Domain,
		Expires:  expires,
		MaxAge:   maxAge,
		Secure:   true,                    // HTTPS-only (HTTP listener redirects, never serves auth)
		HttpOnly: true,                    // Prevent JavaScript access (XSS protection)
		SameSite: http.SameSiteStrictMode, // Block cross-site contexts (CSRF)
	}
}

// SetAccessTokenCookie sets the access token as an httpOnly cookie.
func SetAccessTokenCookie(w http.ResponseWriter, token string, duration time.Duration, config CookieConfig) {
	http.SetCookie(w, newAuthCookie(CookieNameAccess, token, time.Now().Add(duration), int(duration.Seconds()), config))
}

// SetRefreshTokenCookie sets the refresh token as an httpOnly cookie.
func SetRefreshTokenCookie(w http.ResponseWriter, token string, duration time.Duration, config CookieConfig) {
	http.SetCookie(
		w,
		newAuthCookie(CookieNameRefresh, token, time.Now().Add(duration), int(duration.Seconds()), config),
	)
}

// ClearAuthCookies removes both access and refresh token cookies.
func ClearAuthCookies(w http.ResponseWriter, config CookieConfig) {
	for _, name := range []string{CookieNameAccess, CookieNameRefresh} {
		http.SetCookie(w, newAuthCookie(name, "", time.Unix(0, 0), -1, config))
	}
}

// ErrCookieNotFound indicates the requested cookie was not found.
var ErrCookieNotFound = errors.New("cookie not found")

// GetAccessTokenFromCookie extracts the access token from cookies.
func GetAccessTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(CookieNameAccess)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", ErrCookieNotFound
		}
		return "", fmt.Errorf("get access cookie: %w", err)
	}
	return cookie.Value, nil
}

// GetRefreshTokenFromCookie extracts the refresh token from cookies.
func GetRefreshTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(CookieNameRefresh)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", ErrCookieNotFound
		}
		return "", fmt.Errorf("get refresh cookie: %w", err)
	}
	return cookie.Value, nil
}

// GetTokenFromRequest tries to extract token from request in order of preference:
// 1. Cookie (most secure).
// 2. Authorization header (Bearer token - fallback for API clients).
// Returns the token and the source ("cookie", "header", or "none").
func GetTokenFromRequest(r *http.Request) (string, string) {
	// Try cookie first (most secure).
	token, err := GetAccessTokenFromCookie(r)
	if err == nil && token != "" {
		return token, "cookie"
	}

	// Try Authorization header (API client fallback).
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		const bearerPrefix = "Bearer "
		if len(authHeader) > len(bearerPrefix) && strings.EqualFold(authHeader[:len(bearerPrefix)], bearerPrefix) {
			return authHeader[len(bearerPrefix):], "header"
		}
	}

	return "", "none"
}
