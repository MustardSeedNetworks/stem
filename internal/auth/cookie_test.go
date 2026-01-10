// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package auth_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/krisarmstrong/stem/internal/auth"
)

func TestDefaultCookieConfig_Secure(t *testing.T) {
	t.Parallel()

	config := auth.DefaultCookieConfig(true)

	if !config.Secure {
		t.Error("Secure should be true when passed true")
	}
	if config.SameSite != http.SameSiteStrictMode {
		t.Errorf("SameSite = %v, want SameSiteStrictMode", config.SameSite)
	}
	if config.Domain != "" {
		t.Errorf("Domain = %q, want empty string", config.Domain)
	}
	if config.Path != "/" {
		t.Errorf("Path = %q, want /", config.Path)
	}
}

func TestDefaultCookieConfig_Insecure(t *testing.T) {
	t.Parallel()

	config := auth.DefaultCookieConfig(false)

	if config.Secure {
		t.Error("Secure should be false when passed false")
	}
	if config.SameSite != http.SameSiteStrictMode {
		t.Errorf("SameSite = %v, want SameSiteStrictMode", config.SameSite)
	}
	if config.Path != "/" {
		t.Errorf("Path = %q, want /", config.Path)
	}
}

func TestSetAccessTokenCookie(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	config := auth.CookieConfig{
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Domain:   "example.com",
		Path:     "/api",
	}
	duration := 15 * time.Minute
	token := "test-access-token"

	auth.SetAccessTokenCookie(w, token, duration, config)

	resp := w.Result()
	defer resp.Body.Close()

	cookies := resp.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != auth.CookieNameAccess {
		t.Errorf("cookie name = %q, want %q", cookie.Name, auth.CookieNameAccess)
	}
	if cookie.Value != token {
		t.Errorf("cookie value = %q, want %q", cookie.Value, token)
	}
	if !cookie.Secure {
		t.Error("cookie should be secure")
	}
	if !cookie.HttpOnly {
		t.Error("cookie should be HttpOnly")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("cookie SameSite = %v, want SameSiteStrictMode", cookie.SameSite)
	}
	if cookie.Path != "/api" {
		t.Errorf("cookie path = %q, want /api", cookie.Path)
	}
	if cookie.Domain != "example.com" {
		t.Errorf("cookie domain = %q, want example.com", cookie.Domain)
	}
	if cookie.MaxAge != int(duration.Seconds()) {
		t.Errorf("cookie MaxAge = %d, want %d", cookie.MaxAge, int(duration.Seconds()))
	}
}

func TestSetRefreshTokenCookie(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	config := auth.CookieConfig{
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Domain:   "",
		Path:     "/",
	}
	duration := 7 * 24 * time.Hour
	token := "test-refresh-token"

	auth.SetRefreshTokenCookie(w, token, duration, config)

	resp := w.Result()
	defer resp.Body.Close()

	cookies := resp.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != auth.CookieNameRefresh {
		t.Errorf("cookie name = %q, want %q", cookie.Name, auth.CookieNameRefresh)
	}
	if cookie.Value != token {
		t.Errorf("cookie value = %q, want %q", cookie.Value, token)
	}
	if !cookie.Secure {
		t.Error("cookie should be secure")
	}
	if !cookie.HttpOnly {
		t.Error("cookie should be HttpOnly")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("cookie SameSite = %v, want SameSiteLaxMode", cookie.SameSite)
	}
	if cookie.MaxAge != int(duration.Seconds()) {
		t.Errorf("cookie MaxAge = %d, want %d", cookie.MaxAge, int(duration.Seconds()))
	}
}

func TestClearAuthCookies(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	config := auth.CookieConfig{
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Domain:   "",
		Path:     "/",
	}

	auth.ClearAuthCookies(w, config)

	resp := w.Result()
	defer resp.Body.Close()

	cookies := resp.Cookies()
	if len(cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(cookies))
	}

	// Check both cookies are cleared (MaxAge = -1).
	cookieNames := make(map[string]bool)
	for _, cookie := range cookies {
		cookieNames[cookie.Name] = true

		if cookie.Value != "" {
			t.Errorf("cookie %q value should be empty, got %q", cookie.Name, cookie.Value)
		}
		if cookie.MaxAge != -1 {
			t.Errorf("cookie %q MaxAge = %d, want -1", cookie.Name, cookie.MaxAge)
		}
		if !cookie.HttpOnly {
			t.Errorf("cookie %q should be HttpOnly", cookie.Name)
		}
	}

	if !cookieNames[auth.CookieNameAccess] {
		t.Errorf("expected access cookie %q to be cleared", auth.CookieNameAccess)
	}
	if !cookieNames[auth.CookieNameRefresh] {
		t.Errorf("expected refresh cookie %q to be cleared", auth.CookieNameRefresh)
	}
}

func TestGetAccessTokenFromCookie_Found(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.CookieNameAccess,
		Value: "test-token",
	})

	token, err := auth.GetAccessTokenFromCookie(req)
	if err != nil {
		t.Fatalf("GetAccessTokenFromCookie() error = %v", err)
	}
	if token != "test-token" {
		t.Errorf("token = %q, want test-token", token)
	}
}

func TestGetAccessTokenFromCookie_NotFound(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := auth.GetAccessTokenFromCookie(req)
	if err == nil {
		t.Error("expected error when cookie not found")
	}
	if !errors.Is(err, auth.ErrCookieNotFound) {
		t.Errorf("error = %v, want ErrCookieNotFound", err)
	}
}

func TestGetRefreshTokenFromCookie_Found(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.CookieNameRefresh,
		Value: "test-refresh-token",
	})

	token, err := auth.GetRefreshTokenFromCookie(req)
	if err != nil {
		t.Fatalf("GetRefreshTokenFromCookie() error = %v", err)
	}
	if token != "test-refresh-token" {
		t.Errorf("token = %q, want test-refresh-token", token)
	}
}

func TestGetRefreshTokenFromCookie_NotFound(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := auth.GetRefreshTokenFromCookie(req)
	if err == nil {
		t.Error("expected error when cookie not found")
	}
	if !errors.Is(err, auth.ErrCookieNotFound) {
		t.Errorf("error = %v, want ErrCookieNotFound", err)
	}
}

func TestGetTokenFromRequest_FromCookie(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.CookieNameAccess,
		Value: "cookie-token",
	})

	token, source := auth.GetTokenFromRequest(req)
	if token != "cookie-token" {
		t.Errorf("token = %q, want cookie-token", token)
	}
	if source != "cookie" {
		t.Errorf("source = %q, want cookie", source)
	}
}

func TestGetTokenFromRequest_FromHeader(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer header-token")

	token, source := auth.GetTokenFromRequest(req)
	if token != "header-token" {
		t.Errorf("token = %q, want header-token", token)
	}
	if source != "header" {
		t.Errorf("source = %q, want header", source)
	}
}

func TestGetTokenFromRequest_CookiePriority(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.CookieNameAccess,
		Value: "cookie-token",
	})
	req.Header.Set("Authorization", "Bearer header-token")

	token, source := auth.GetTokenFromRequest(req)
	// Cookie should take priority over header.
	if token != "cookie-token" {
		t.Errorf("token = %q, want cookie-token (cookie priority)", token)
	}
	if source != "cookie" {
		t.Errorf("source = %q, want cookie", source)
	}
}

func TestGetTokenFromRequest_None(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	token, source := auth.GetTokenFromRequest(req)
	if token != "" {
		t.Errorf("token = %q, want empty string", token)
	}
	if source != "none" {
		t.Errorf("source = %q, want none", source)
	}
}

func TestGetTokenFromRequest_InvalidBearerFormat(t *testing.T) {
	t.Parallel()

	// Test with just "Bearer" (no token).
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer")

	token, source := auth.GetTokenFromRequest(req)
	if token != "" {
		t.Errorf("token = %q, want empty string for invalid bearer", token)
	}
	if source != "none" {
		t.Errorf("source = %q, want none", source)
	}
}

func TestGetTokenFromRequest_NonBearerAuth(t *testing.T) {
	t.Parallel()

	// Test with Basic auth (not Bearer).
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")

	token, source := auth.GetTokenFromRequest(req)
	if token != "" {
		t.Errorf("token = %q, want empty string for non-bearer auth", token)
	}
	if source != "none" {
		t.Errorf("source = %q, want none", source)
	}
}

func TestGetTokenFromRequest_BearerCaseInsensitive(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		header string
	}{
		{"lowercase", "bearer token-value"},
		{"uppercase", "BEARER token-value"},
		{"mixed", "BeArEr token-value"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", tc.header)

			token, source := auth.GetTokenFromRequest(req)
			if token != "token-value" {
				t.Errorf("token = %q, want token-value", token)
			}
			if source != "header" {
				t.Errorf("source = %q, want header", source)
			}
		})
	}
}

func TestGetTokenFromRequest_EmptyCookie(t *testing.T) {
	t.Parallel()

	// Empty cookie value should fall through to header check.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.CookieNameAccess,
		Value: "",
	})
	req.Header.Set("Authorization", "Bearer header-token")

	token, source := auth.GetTokenFromRequest(req)
	if token != "header-token" {
		t.Errorf("token = %q, want header-token", token)
	}
	if source != "header" {
		t.Errorf("source = %q, want header", source)
	}
}

func TestCookieConstants(t *testing.T) {
	t.Parallel()

	if auth.CookieNameAccess != "stem_access" {
		t.Errorf("CookieNameAccess = %q, want stem_access", auth.CookieNameAccess)
	}
	if auth.CookieNameRefresh != "stem_refresh" {
		t.Errorf("CookieNameRefresh = %q, want stem_refresh", auth.CookieNameRefresh)
	}
}

func TestSetAccessTokenCookie_DefaultConfig(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	config := auth.DefaultCookieConfig(false)
	duration := 5 * time.Minute
	token := "my-access-token"

	auth.SetAccessTokenCookie(w, token, duration, config)

	resp := w.Result()
	defer resp.Body.Close()

	cookies := resp.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Secure {
		t.Error("cookie should not be secure when DefaultCookieConfig(false)")
	}
	if cookie.Path != "/" {
		t.Errorf("cookie path = %q, want /", cookie.Path)
	}
}
