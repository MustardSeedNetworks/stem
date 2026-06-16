// SPDX-License-Identifier: BUSL-1.1

// Black-box tests for the ratelimit package. Uses only the exported API.
// White-box tests (internal helpers) live in limiter_internal_test.go.
package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/api/ratelimit"
)

func TestNewRateLimiter(t *testing.T) {
	rl := ratelimit.NewRateLimiter(1, 1)
	defer rl.Stop()

	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}
}

func TestNewAuthRateLimiter(t *testing.T) {
	rl := ratelimit.NewAuthRateLimiter()
	defer rl.Stop()

	if rl == nil {
		t.Fatal("NewAuthRateLimiter returned nil")
	}
}

func TestNewAPIRateLimiter(t *testing.T) {
	rl := ratelimit.NewAPIRateLimiter()
	defer rl.Stop()

	if rl == nil {
		t.Fatal("NewAPIRateLimiter returned nil")
	}
}

func TestRateLimiterGetLimiter(t *testing.T) {
	rl := ratelimit.NewRateLimiter(1, 1)
	defer rl.Stop()

	// Get limiter for an IP.
	limiter1 := rl.GetLimiter("192.168.1.1")
	if limiter1 == nil {
		t.Fatal("GetLimiter returned nil")
	}

	// Same IP should return the same limiter.
	limiter2 := rl.GetLimiter("192.168.1.1")
	if limiter1 != limiter2 {
		t.Error("Expected same limiter for same IP")
	}

	// Different IP should return a different limiter.
	limiter3 := rl.GetLimiter("192.168.1.2")
	if limiter1 == limiter3 {
		t.Error("Expected different limiter for different IP")
	}
}

func TestRateLimiterAllow(t *testing.T) {
	// Create a rate limiter that allows 2 requests with burst of 2.
	rl := ratelimit.NewRateLimiter(2, 2)
	defer rl.Stop()

	ip := "10.0.0.1"

	// First two requests should be allowed (burst).
	if !rl.Allow(ip) {
		t.Error("First request should be allowed")
	}
	if !rl.Allow(ip) {
		t.Error("Second request should be allowed")
	}

	// Third request should be denied (burst exhausted, rate not replenished).
	if rl.Allow(ip) {
		t.Error("Third request should be denied")
	}
}

func TestRateLimiterAllowDifferentIPs(t *testing.T) {
	// Create a rate limiter with burst of 1.
	rl := ratelimit.NewRateLimiter(1, 1)
	defer rl.Stop()

	ip1 := "10.0.0.1"
	ip2 := "10.0.0.2"

	// First request from each IP should be allowed.
	if !rl.Allow(ip1) {
		t.Error("Request from IP1 should be allowed")
	}
	if !rl.Allow(ip2) {
		t.Error("Request from IP2 should be allowed")
	}

	// Second request from each IP should be denied.
	if rl.Allow(ip1) {
		t.Error("Second request from IP1 should be denied")
	}
	if rl.Allow(ip2) {
		t.Error("Second request from IP2 should be denied")
	}
}

func TestRateLimiterMiddleware(t *testing.T) {
	// Create a rate limiter with burst of 2.
	rl := ratelimit.NewRateLimiter(2, 2)
	defer rl.Stop()

	// Create a simple handler.
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with rate limiting middleware.
	wrapped := rl.Middleware(handler)

	// First two requests should succeed.
	for i := range 2 {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, w.Code)
		}
	}

	// Third request should be rate limited.
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}

	// Check Retry-After header.
	retryAfter := w.Header().Get("Retry-After")
	if retryAfter != "60" {
		t.Errorf("Expected Retry-After: 60, got %s", retryAfter)
	}
}

func TestRateLimiterMiddlewareXForwardedFor(t *testing.T) {
	// Create a rate limiter with burst of 1.
	rl := ratelimit.NewRateLimiter(1, 1)
	defer rl.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware(handler)

	// Request with X-Forwarded-For header.
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.195, 70.41.3.18, 150.172.238.178")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Second request from same X-Forwarded-For should be denied.
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "127.0.0.1:12346"
	req2.Header.Set("X-Forwarded-For", "203.0.113.195, 70.41.3.18")
	w2 := httptest.NewRecorder()

	wrapped.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429 for same X-Forwarded-For IP, got %d", w2.Code)
	}
}

func TestRateLimiterMiddlewareXRealIP(t *testing.T) {
	// Create a rate limiter with burst of 1.
	rl := ratelimit.NewRateLimiter(1, 1)
	defer rl.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware(handler)

	// Request with X-Real-IP header.
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Real-IP", "198.51.100.42")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Second request from same X-Real-IP should be denied.
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "127.0.0.1:12346"
	req2.Header.Set("X-Real-IP", "198.51.100.42")
	w2 := httptest.NewRecorder()

	wrapped.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429 for same X-Real-IP, got %d", w2.Code)
	}
}

func TestRateLimiterConcurrency(t *testing.T) {
	t.Parallel()

	// Create a rate limiter with high burst to avoid rate limiting during test.
	rl := ratelimit.NewRateLimiter(1000, 1000)
	defer rl.Stop()

	var wg sync.WaitGroup
	numGoroutines := 100
	requestsPerGoroutine := 10

	// Spawn multiple goroutines making concurrent requests.
	for range numGoroutines {
		wg.Go(func() {
			ip := "192.168.1.1" // All use the same IP to stress the sync.

			for range requestsPerGoroutine {
				rl.Allow(ip)
			}
		})
	}

	wg.Wait()
	// If we get here without deadlock or panic, the test passes.
}

func TestRateLimiterStop(t *testing.T) {
	t.Parallel()

	rl := ratelimit.NewRateLimiter(1, 1)

	// Stop should not panic.
	rl.Stop()

	// Small delay to ensure goroutine has time to exit.
	time.Sleep(10 * time.Millisecond)
}

func TestAuthRateLimiterLimits(t *testing.T) {
	rl := ratelimit.NewAuthRateLimiter()
	defer rl.Stop()

	ip := "10.0.0.100"

	// Auth limiter has burst of 5, so first 5 requests should succeed.
	for i := range 5 {
		if !rl.Allow(ip) {
			t.Errorf("Auth request %d should be allowed within burst", i+1)
		}
	}

	// 6th request should be denied.
	if rl.Allow(ip) {
		t.Error("6th auth request should be denied (burst exhausted)")
	}
}

func TestAPIRateLimiterLimits(t *testing.T) {
	rl := ratelimit.NewAPIRateLimiter()
	defer rl.Stop()

	ip := "10.0.0.200"

	// API limiter has burst of 100, so first 100 requests should succeed.
	for i := range 100 {
		if !rl.Allow(ip) {
			t.Errorf("API request %d should be allowed within burst", i+1)
		}
	}

	// 101st request should be denied.
	if rl.Allow(ip) {
		t.Error("101st API request should be denied (burst exhausted)")
	}
}

// Benchmark tests.
func BenchmarkRateLimiterAllow(b *testing.B) {
	rl := ratelimit.NewRateLimiter(1000000, 1000000) // High limit to avoid rate limiting.
	defer rl.Stop()

	for b.Loop() {
		rl.Allow("192.168.1.1")
	}
}

func BenchmarkRateLimiterMiddleware(b *testing.B) {
	rl := ratelimit.NewRateLimiter(1000000, 1000000)
	defer rl.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := rl.Middleware(handler)

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
	}
}
