// SPDX-License-Identifier: BUSL-1.1

package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/logging"

	"golang.org/x/time/rate"
)

// TestAPIRateLimitFromEnv locks the override's one-directional contract: it may
// raise the standard API limit and may never lower or disable it, so no value
// of APIRateLimitEnv can leave the binary less rate-limited than its default.
func TestAPIRateLimitFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		set      bool
		envValue string
		want     int
	}{
		{name: "unset uses the compiled-in default", set: false, want: APIRateLimit},
		{name: "a higher value is honoured", set: true, envValue: "1000", want: 1000},
		{name: "surrounding whitespace is tolerated", set: true, envValue: "  1000\t", want: 1000},
		{
			name: "one above the default is honoured",
			set:  true, envValue: strconv.Itoa(APIRateLimit + 1), want: APIRateLimit + 1,
		},
		{name: "the default itself is a no-op", set: true, envValue: strconv.Itoa(APIRateLimit), want: APIRateLimit},
		{name: "a lower value cannot weaken the limiter", set: true, envValue: "1", want: APIRateLimit},
		{name: "zero cannot disable the limiter", set: true, envValue: "0", want: APIRateLimit},
		{name: "a negative value cannot disable the limiter", set: true, envValue: "-1", want: APIRateLimit},
		{name: "a non-numeric value falls back", set: true, envValue: "lots", want: APIRateLimit},
		{name: "an empty value falls back", set: true, envValue: "", want: APIRateLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				t.Setenv(APIRateLimitEnv, tt.envValue)
			}

			if got := apiRateLimitFromEnv(); got != tt.want {
				t.Errorf("apiRateLimitFromEnv() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestAuthRateLimitFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     int
	}{
		{name: "higher value raises the limit", envValue: "200", want: 200},
		{name: "default cannot be lowered", envValue: "4", want: AuthRateLimit},
		{name: "malformed value falls back", envValue: "many", want: AuthRateLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(AuthRateLimitEnv, tt.envValue)
			if got := authRateLimitFromEnv(); got != tt.want {
				t.Errorf("authRateLimitFromEnv() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNewAuthRateLimiterHonoursOverride(t *testing.T) {
	t.Setenv(AuthRateLimitEnv, strconv.Itoa(AuthRateLimit*4))
	rl := NewAuthRateLimiter(nil)
	defer rl.Stop()

	for i := range AuthRateLimit * 4 {
		if !rl.Allow("10.0.0.202") {
			t.Fatalf("request %d denied; raised burst should allow %d", i+1, AuthRateLimit*4)
		}
	}
}

// TestNewAPIRateLimiterHonoursOverride proves the override reaches the
// constructed limiter's burst, not merely the helper: the regression it guards
// is a bucket that still runs dry at the default after the env is set.
func TestNewAPIRateLimiterHonoursOverride(t *testing.T) {
	t.Setenv(APIRateLimitEnv, strconv.Itoa(APIRateLimit*4))

	rl := NewAPIRateLimiter(nil)
	defer rl.Stop()

	ip := "10.0.0.201"
	for i := range APIRateLimit * 4 {
		if !rl.Allow(ip) {
			t.Fatalf("request %d denied; the raised burst should allow %d", i+1, APIRateLimit*4)
		}
	}

	if rl.Allow(ip) {
		t.Error("request beyond the raised burst should be denied; the limiter must stay on")
	}
}

// TestCleanup tests the cleanup function.
func TestCleanup(t *testing.T) {
	t.Run("removes old visitors", func(t *testing.T) {
		rl := &RateLimiter{
			visitors:      make(map[string]*visitor),
			mu:            defaultRWMutex(),
			rate:          rate.Limit(1),
			burst:         1,
			done:          make(chan struct{}),
			globalLimiter: rate.NewLimiter(1, 1),
			maxVisitors:   MaxVisitors,
		}

		// Add visitors with old timestamps.
		oldTime := time.Now().Add(-VisitorTTL - time.Minute)
		rl.visitors["old-ip"] = &visitor{
			limiter:  rate.NewLimiter(1, 1),
			lastSeen: oldTime,
		}

		// Add a recent visitor.
		rl.visitors["new-ip"] = &visitor{
			limiter:  rate.NewLimiter(1, 1),
			lastSeen: time.Now(),
		}

		// Run cleanup.
		rl.cleanup()

		// Old visitor should be removed.
		rl.mu.RLock()
		_, oldExists := rl.visitors["old-ip"]
		_, newExists := rl.visitors["new-ip"]
		rl.mu.RUnlock()

		if oldExists {
			t.Error("Expected old visitor to be removed")
		}
		if !newExists {
			t.Error("Expected new visitor to still exist")
		}
	})

	t.Run("keeps all recent visitors", func(t *testing.T) {
		rl := &RateLimiter{
			visitors:      make(map[string]*visitor),
			mu:            defaultRWMutex(),
			rate:          rate.Limit(1),
			burst:         1,
			done:          make(chan struct{}),
			globalLimiter: rate.NewLimiter(1, 1),
			maxVisitors:   MaxVisitors,
		}

		// Add multiple recent visitors.
		for i := range 10 {
			ip := "192.168.1." + string(rune('0'+i))
			rl.visitors[ip] = &visitor{
				limiter:  rate.NewLimiter(1, 1),
				lastSeen: time.Now(),
			}
		}

		initialCount := len(rl.visitors)

		// Run cleanup.
		rl.cleanup()

		rl.mu.RLock()
		finalCount := len(rl.visitors)
		rl.mu.RUnlock()

		if finalCount != initialCount {
			t.Errorf("Expected %d visitors after cleanup, got %d", initialCount, finalCount)
		}
	})

	t.Run("removes all old visitors", func(t *testing.T) {
		rl := &RateLimiter{
			visitors:      make(map[string]*visitor),
			mu:            defaultRWMutex(),
			rate:          rate.Limit(1),
			burst:         1,
			done:          make(chan struct{}),
			globalLimiter: rate.NewLimiter(1, 1),
			maxVisitors:   MaxVisitors,
		}

		// Add multiple old visitors.
		oldTime := time.Now().Add(-VisitorTTL - time.Minute)
		for i := range 10 {
			ip := "10.0.0." + string(rune('0'+i))
			rl.visitors[ip] = &visitor{
				limiter:  rate.NewLimiter(1, 1),
				lastSeen: oldTime,
			}
		}

		// Run cleanup.
		rl.cleanup()

		rl.mu.RLock()
		finalCount := len(rl.visitors)
		rl.mu.RUnlock()

		if finalCount != 0 {
			t.Errorf("Expected 0 visitors after cleanup, got %d", finalCount)
		}
	})

	t.Run("handles empty map", func(t *testing.T) {
		rl := &RateLimiter{
			visitors:      make(map[string]*visitor),
			mu:            defaultRWMutex(),
			rate:          rate.Limit(1),
			burst:         1,
			done:          make(chan struct{}),
			globalLimiter: rate.NewLimiter(1, 1),
			maxVisitors:   MaxVisitors,
		}

		// Should not panic on empty map.
		rl.cleanup()

		rl.mu.RLock()
		count := len(rl.visitors)
		rl.mu.RUnlock()

		if count != 0 {
			t.Errorf("Expected 0 visitors, got %d", count)
		}
	})

	t.Run("boundary condition - exactly at TTL", func(t *testing.T) {
		rl := &RateLimiter{
			visitors:      make(map[string]*visitor),
			mu:            defaultRWMutex(),
			rate:          rate.Limit(1),
			burst:         1,
			done:          make(chan struct{}),
			globalLimiter: rate.NewLimiter(1, 1),
			maxVisitors:   MaxVisitors,
		}

		// Add visitor exactly at TTL boundary.
		boundaryTime := time.Now().Add(-VisitorTTL)
		rl.visitors["boundary-ip"] = &visitor{
			limiter:  rate.NewLimiter(1, 1),
			lastSeen: boundaryTime,
		}

		// Run cleanup.
		rl.cleanup()

		rl.mu.RLock()
		_, exists := rl.visitors["boundary-ip"]
		rl.mu.RUnlock()

		// Boundary visitor should be removed (Before check is exclusive).
		if exists {
			t.Error("Expected boundary visitor to be removed")
		}
	})
}

// The bucket key used to be derived here; it now has one home in
// internal/logging (SecurityClientIP), where the table that used to live in
// this file is kept. What is exercised here is that the limiter passes its
// configured trust list to it.
func TestAuthRateLimiterHonoursTrustedProxies(t *testing.T) {
	trusted, err := logging.ParseTrustedProxies("10.0.0.0/24")
	if err != nil {
		t.Fatalf("ParseTrustedProxies() = %v", err)
	}

	rl := NewAuthRateLimiter(trusted)
	defer rl.Stop()

	// Two clients arriving through the trusted proxy are two buckets, so
	// neither spends the other's burst.
	behindProxy := requestFrom("10.0.0.5:1", "198.51.100.7")
	if key := logging.SecurityClientIP(behindProxy, rl.trustedProxies); key != "198.51.100.7" {
		t.Errorf("bucket key behind the trusted proxy = %q, want the forwarded client", key)
	}

	untrusted := requestFrom("203.0.113.9:1", "198.51.100.7")
	if key := logging.SecurityClientIP(untrusted, rl.trustedProxies); key != "203.0.113.9" {
		t.Errorf("bucket key from an untrusted peer = %q, want the peer", key)
	}
}

// requestFrom builds a request from a peer, carrying a forwarding header.
func requestFrom(remoteAddr, xff string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.RemoteAddr = remoteAddr
	r.Header.Set("X-Forwarded-For", xff)
	return r
}

// TestCleanupLoop tests the cleanup loop starts and stops properly.
func TestCleanupLoop(t *testing.T) {
	t.Run("cleanup loop stops on done signal", func(_ *testing.T) {
		rl := NewRateLimiter(rate.Limit(1), 1)

		// Give cleanup loop time to start.
		time.Sleep(10 * time.Millisecond)

		// Stop should close the done channel.
		rl.Stop()

		// Give cleanup goroutine time to exit.
		time.Sleep(20 * time.Millisecond)

		// Test passes if we don't hang or panic.
	})
}

// TestVisitorLastSeenUpdated tests that lastSeen is updated on access.
func TestVisitorLastSeenUpdated(t *testing.T) {
	rl := NewRateLimiter(rate.Limit(1), 1)
	defer rl.Stop()

	ip := "10.10.10.10"

	// First access creates visitor.
	_ = rl.GetLimiter(ip)

	rl.mu.RLock()
	firstSeen := rl.visitors[ip].lastSeen
	rl.mu.RUnlock()

	// Small delay.
	time.Sleep(10 * time.Millisecond)

	// Second access should update lastSeen.
	_ = rl.GetLimiter(ip)

	rl.mu.RLock()
	secondSeen := rl.visitors[ip].lastSeen
	rl.mu.RUnlock()

	if !secondSeen.After(firstSeen) {
		t.Error("Expected lastSeen to be updated on second access")
	}
}

// defaultRWMutex returns a default [sync.RWMutex] for testing.
func defaultRWMutex() sync.RWMutex {
	return sync.RWMutex{}
}
