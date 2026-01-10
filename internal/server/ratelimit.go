// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/krisarmstrong/stem/internal/logging"
)

// Rate limiting configuration constants.
const (
	// AuthRateLimit is the rate limit for authentication endpoints (per minute).
	AuthRateLimit = 5
	// AuthBurstLimit is the burst limit for authentication endpoints.
	AuthBurstLimit = 5

	// APIRateLimit is the rate limit for standard API endpoints (per minute).
	APIRateLimit = 100
	// APIBurstLimit is the burst limit for standard API endpoints.
	APIBurstLimit = 100

	// CleanupInterval is how often to clean up old rate limiter entries.
	CleanupInterval = 10 * time.Minute
	// VisitorTTL is how long to keep a visitor's rate limiter after last access.
	VisitorTTL = 15 * time.Minute

	// MaxVisitors is the maximum number of unique IPs to track.
	// Prevents memory exhaustion from IP spoofing attacks.
	MaxVisitors = 10000

	// secondsPerMinute is used for converting per-minute rates to per-second.
	secondsPerMinute = 60.0

	// globalRateDivisor is how much stricter the global fallback limiter is.
	// When max visitors is exceeded, new IPs share a limiter that is 10x stricter.
	globalRateDivisor = 10

	// capacityThresholdHigh is 80% capacity - triggers moderate cleanup.
	capacityThresholdHigh = 80
	// capacityThresholdCritical is 90% capacity - triggers aggressive cleanup.
	capacityThresholdCritical = 90
	// percentDivisor is used for percentage calculations.
	percentDivisor = 100

	// ttlDivisorModerate reduces TTL by half at high capacity.
	ttlDivisorModerate = 2
	// ttlDivisorAggressive reduces TTL to quarter at critical capacity.
	ttlDivisorAggressive = 4
)

// visitor holds the rate limiter and last seen time for an IP.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter provides per-IP rate limiting using the token bucket algorithm.
type RateLimiter struct {
	visitors      map[string]*visitor
	mu            sync.RWMutex
	rate          rate.Limit
	burst         int
	done          chan struct{}
	globalLimiter *rate.Limiter // Fallback limiter when max visitors exceeded
	maxVisitors   int           // Maximum number of IPs to track
}

// NewRateLimiter creates a new rate limiter with the specified rate (events per second) and burst.
// The rate is specified as events per second; use rate.Every() for other intervals.
func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	// Create a global fallback limiter with more restrictive rate.
	// When max visitors is exceeded, all new IPs share this stricter limiter.
	globalRate := r / globalRateDivisor
	if globalRate < 1 {
		globalRate = 1
	}

	rl := &RateLimiter{
		visitors:      make(map[string]*visitor),
		mu:            sync.RWMutex{},
		rate:          r,
		burst:         burst,
		done:          make(chan struct{}),
		globalLimiter: rate.NewLimiter(globalRate, 1), // Very restrictive fallback
		maxVisitors:   MaxVisitors,
	}

	// Start background cleanup goroutine.
	go rl.cleanupLoop()

	return rl
}

// NewAuthRateLimiter creates a rate limiter configured for authentication endpoints.
// Limits to 5 requests per minute with burst of 5.
func NewAuthRateLimiter() *RateLimiter {
	// Convert per-minute rate to per-second for rate.Limit.
	r := rate.Limit(float64(AuthRateLimit) / secondsPerMinute)
	return NewRateLimiter(r, AuthBurstLimit)
}

// NewAPIRateLimiter creates a rate limiter configured for standard API endpoints.
// Limits to 100 requests per minute with burst of 100.
func NewAPIRateLimiter() *RateLimiter {
	// Convert per-minute rate to per-second for rate.Limit.
	r := rate.Limit(float64(APIRateLimit) / secondsPerMinute)
	return NewRateLimiter(r, APIBurstLimit)
}

// GetLimiter returns the rate limiter for the given IP address.
// Creates a new limiter if one doesn't exist for this IP.
// If max visitors is reached, returns a shared global limiter to prevent memory exhaustion.
func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	v, exists := rl.visitors[ip]
	visitorCount := len(rl.visitors)
	rl.mu.RUnlock()

	if exists {
		rl.mu.Lock()
		v.lastSeen = time.Now()
		rl.mu.Unlock()
		return v.limiter
	}

	// Check if we've hit the max visitors limit to prevent memory exhaustion.
	// New IPs will share a more restrictive global limiter.
	if visitorCount >= rl.maxVisitors {
		logging.Warn("Rate limiter at max capacity, using global limiter",
			"ip", ip,
			"maxVisitors", rl.maxVisitors,
		)
		return rl.globalLimiter
	}

	// Create new limiter for this IP.
	limiter := rate.NewLimiter(rl.rate, rl.burst)
	rl.mu.Lock()
	// Double-check under write lock to avoid race condition.
	if len(rl.visitors) >= rl.maxVisitors {
		rl.mu.Unlock()
		return rl.globalLimiter
	}
	rl.visitors[ip] = &visitor{
		limiter:  limiter,
		lastSeen: time.Now(),
	}
	rl.mu.Unlock()

	return limiter
}

// Allow checks if a request from the given IP should be allowed.
func (rl *RateLimiter) Allow(ip string) bool {
	return rl.GetLimiter(ip).Allow()
}

// cleanupLoop periodically removes old visitors to prevent memory leaks.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.done:
			return
		}
	}
}

// cleanup removes visitors that haven't been seen recently.
// When approaching max capacity, uses more aggressive TTL to free up space.
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	visitorCount := len(rl.visitors)
	ttl := VisitorTTL

	// Calculate capacity thresholds.
	criticalThreshold := rl.maxVisitors * capacityThresholdCritical / percentDivisor
	highThreshold := rl.maxVisitors * capacityThresholdHigh / percentDivisor

	// Use more aggressive cleanup when approaching max capacity.
	// At 90% capacity, use quarter TTL; at 80%, use half TTL.
	if visitorCount > criticalThreshold {
		ttl = VisitorTTL / ttlDivisorAggressive
	} else if visitorCount > highThreshold {
		ttl = VisitorTTL / ttlDivisorModerate
	}

	cutoff := time.Now().Add(-ttl)
	for ip, v := range rl.visitors {
		if v.lastSeen.Before(cutoff) {
			delete(rl.visitors, ip)
		}
	}

	// Log if we're still at high capacity after cleanup.
	newCount := len(rl.visitors)
	if newCount > highThreshold {
		logging.Warn("Rate limiter at high capacity after cleanup",
			"visitorCount", newCount,
			"maxVisitors", rl.maxVisitors,
			"ttlUsed", ttl.String(),
		)
	}
}

// Stop stops the cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.done)
}

// Middleware returns an HTTP middleware that applies rate limiting.
// Responds with 429 Too Many Requests when the limit is exceeded.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		if !rl.Allow(ip) {
			logging.Warn("Rate limit exceeded",
				"ip", ip,
				"path", r.URL.Path,
				"method", r.Method,
			)

			// Audit log the rate limit event.
			logging.AuditRateLimited(r.Context(), r, "", r.URL.Path, "1m")

			w.Header().Set("Retry-After", "60")
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the client IP from the request.
// Checks X-Forwarded-For and X-Real-IP headers before falling back to RemoteAddr.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (may contain comma-separated list).
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the list (client's original IP).
		ip := xff
		for i := range xff {
			if xff[i] == ',' {
				ip = xff[:i]
				break
			}
		}
		ip = trimSpace(ip)
		if ip != "" {
			return ip
		}
	}

	// Check X-Real-IP header.
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return trimSpace(xri)
	}

	// Fall back to RemoteAddr, stripping the port.
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr might not have a port.
		return r.RemoteAddr
	}
	return ip
}

// trimSpace removes leading and trailing whitespace from a string.
func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}

	return s[start:end]
}
