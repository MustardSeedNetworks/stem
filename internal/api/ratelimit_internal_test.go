// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// TestTrimSpace tests the trimSpace function.
func TestTrimSpace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no whitespace",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "leading spaces",
			input: "   hello",
			want:  "hello",
		},
		{
			name:  "trailing spaces",
			input: "hello   ",
			want:  "hello",
		},
		{
			name:  "both sides spaces",
			input: "   hello   ",
			want:  "hello",
		},
		{
			name:  "leading tabs",
			input: "\t\thello",
			want:  "hello",
		},
		{
			name:  "trailing tabs",
			input: "hello\t\t",
			want:  "hello",
		},
		{
			name:  "mixed whitespace",
			input: " \t hello \t ",
			want:  "hello",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only spaces",
			input: "     ",
			want:  "",
		},
		{
			name:  "only tabs",
			input: "\t\t\t",
			want:  "",
		},
		{
			name:  "single character",
			input: "a",
			want:  "a",
		},
		{
			name:  "single character with spaces",
			input: "  a  ",
			want:  "a",
		},
		{
			name:  "internal spaces preserved",
			input: "  hello world  ",
			want:  "hello world",
		},
		{
			name:  "internal tabs preserved",
			input: "\thello\tworld\t",
			want:  "hello\tworld",
		},
		{
			name:  "IP address with spaces",
			input: "  192.168.1.1  ",
			want:  "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trimSpace(tt.input)
			if got != tt.want {
				t.Errorf("trimSpace(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
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

// TestGetClientIP tests the getClientIP function.
func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		xri        string
		want       string
	}{
		{
			name:       "simple remote addr",
			remoteAddr: "192.168.1.1:12345",
			xff:        "",
			xri:        "",
			want:       "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For single IP",
			remoteAddr: "127.0.0.1:12345",
			xff:        "203.0.113.195",
			xri:        "",
			want:       "203.0.113.195",
		},
		{
			name:       "X-Forwarded-For multiple IPs",
			remoteAddr: "127.0.0.1:12345",
			xff:        "203.0.113.195, 70.41.3.18, 150.172.238.178",
			xri:        "",
			want:       "203.0.113.195",
		},
		{
			name:       "X-Real-IP takes precedence over RemoteAddr",
			remoteAddr: "127.0.0.1:12345",
			xff:        "",
			xri:        "198.51.100.42",
			want:       "198.51.100.42",
		},
		{
			name:       "X-Forwarded-For takes precedence over X-Real-IP",
			remoteAddr: "127.0.0.1:12345",
			xff:        "203.0.113.195",
			xri:        "198.51.100.42",
			want:       "203.0.113.195",
		},
		{
			name:       "X-Forwarded-For with spaces",
			remoteAddr: "127.0.0.1:12345",
			xff:        "  203.0.113.195  ",
			xri:        "",
			want:       "203.0.113.195",
		},
		{
			name:       "Remote addr without port",
			remoteAddr: "192.168.1.1",
			xff:        "",
			xri:        "",
			want:       "192.168.1.1",
		},
		{
			name:       "IPv6 remote addr",
			remoteAddr: "[::1]:12345",
			xff:        "",
			xri:        "",
			want:       "::1",
		},
		{
			name:       "empty X-Forwarded-For fallback to X-Real-IP",
			remoteAddr: "127.0.0.1:12345",
			xff:        "",
			xri:        "10.0.0.1",
			want:       "10.0.0.1",
		},
		{
			name:       "whitespace-only X-Forwarded-For fallback to X-Real-IP",
			remoteAddr: "127.0.0.1:12345",
			xff:        "   ",
			xri:        "10.0.0.1",
			want:       "10.0.0.1",
		},
		{
			// Security: non-loopback peers MUST NOT be able to spoof their
			// source IP by sending forged X-Forwarded-For. Returns the actual
			// TCP peer instead.
			name:       "non-loopback peer cannot spoof X-Forwarded-For",
			remoteAddr: "203.0.113.50:12345",
			xff:        "1.2.3.4",
			xri:        "",
			want:       "203.0.113.50",
		},
		{
			// Same as above, but with X-Real-IP.
			name:       "non-loopback peer cannot spoof X-Real-IP",
			remoteAddr: "203.0.113.50:12345",
			xff:        "",
			xri:        "1.2.3.4",
			want:       "203.0.113.50",
		},
		{
			// IPv6 loopback is also a trusted source for forwarding headers.
			name:       "IPv6 loopback peer can forward client IP",
			remoteAddr: "[::1]:12345",
			xff:        "203.0.113.195",
			xri:        "",
			want:       "203.0.113.195",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				req.Header.Set("X-Real-IP", tt.xri)
			}

			got := getClientIP(req)
			if got != tt.want {
				t.Errorf("getClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
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
