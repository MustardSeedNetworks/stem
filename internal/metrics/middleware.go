// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package metrics

import (
	"net/http"
	"strconv"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter

	statusCode int
}

// WriteHeader captures the status code before writing.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Middleware returns HTTP middleware that records request metrics.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code.
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default if WriteHeader not called.
		}

		// Call the next handler.
		next.ServeHTTP(wrapped, r)

		// Record metrics.
		duration := time.Since(start).Seconds()
		path := normalizePath(r.URL.Path)
		status := strconv.Itoa(wrapped.statusCode)

		RecordHTTPRequest(r.Method, path, status)
		ObserveHTTPDuration(r.Method, path, duration)
	})
}

// normalizePath reduces path cardinality by grouping dynamic segments.
// This prevents high-cardinality metrics from paths like /api/test/123.
func normalizePath(path string) string {
	// Keep API paths but normalize dynamic segments.
	switch {
	case len(path) > 12 && path[:12] == "/api/modules":
		return "/api/modules"
	case len(path) > 9 && path[:9] == "/api/test":
		return "/api/test"
	default:
		return path
	}
}
