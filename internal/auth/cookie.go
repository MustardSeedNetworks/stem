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

// CookieConfig holds cookie security settings.
type CookieConfig struct {
	// Secure sets the Secure flag (HTTPS only)
	Secure bool

	// SameSite sets the SameSite attribute
	SameSite http.SameSite

	// Domain sets the cookie domain
	Domain string

	// Path sets the cookie path
	Path string
}

// DefaultCookieConfig returns secure defaults.
// SameSite=Strict prevents CSRF attacks by blocking cookies in cross-site contexts.
func DefaultCookieConfig(secure bool) CookieConfig {
	return CookieConfig{
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Domain:   "", // Current domain
		Path:     "/",
	}
}

// SetAccessTokenCookie sets the access token as an httpOnly cookie.
func SetAccessTokenCookie(w http.ResponseWriter, token string, duration time.Duration, config CookieConfig) {
	cookie := &http.Cookie{
		Name:     CookieNameAccess,
		Value:    token,
		Path:     config.Path,
		Domain:   config.Domain,
		Expires:  time.Now().Add(duration),
		MaxAge:   int(duration.Seconds()),
		Secure:   config.Secure,
		HttpOnly: true, // Prevent JavaScript access (XSS protection)
		SameSite: config.SameSite,
	}
	http.SetCookie(w, cookie)
}

// SetRefreshTokenCookie sets the refresh token as an httpOnly cookie.
func SetRefreshTokenCookie(w http.ResponseWriter, token string, duration time.Duration, config CookieConfig) {
	cookie := &http.Cookie{
		Name:     CookieNameRefresh,
		Value:    token,
		Path:     config.Path,
		Domain:   config.Domain,
		Expires:  time.Now().Add(duration),
		MaxAge:   int(duration.Seconds()),
		Secure:   config.Secure,
		HttpOnly: true,
		SameSite: config.SameSite,
	}
	http.SetCookie(w, cookie)
}

// ClearAuthCookies removes both access and refresh token cookies.
func ClearAuthCookies(w http.ResponseWriter, config CookieConfig) {
	clearCookie := func(name string) {
		cookie := &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     config.Path,
			Domain:   config.Domain,
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
			Secure:   config.Secure,
			HttpOnly: true,
			SameSite: config.SameSite,
		}
		http.SetCookie(w, cookie)
	}

	clearCookie(CookieNameAccess)
	clearCookie(CookieNameRefresh)
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
