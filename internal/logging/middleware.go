// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package logging

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

const (
	// RequestIDHeader is the HTTP header used to pass request IDs.
	RequestIDHeader = "X-Request-ID"

	// requestIDBytes is the number of random bytes for request ID generation.
	requestIDBytes = 8

	// maxRequestIDLength is the maximum length of a client-provided request ID.
	maxRequestIDLength = 64
)

// RequestIDMiddleware generates a unique request ID for each incoming request
// and adds it to the request context. If the client sends an X-Request-ID header,
// that value is used instead (useful for distributed tracing).
// Client-provided IDs are validated (length/charset); invalid IDs are replaced
// with a generated one to prevent log spoofing or oversized headers.
//
// The request ID is available via RequestIDFromContext(r.Context()) and is
// automatically included in logs when using FromContext(ctx) to get the logger.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for existing request ID from upstream proxy/client
		requestID := r.Header.Get(RequestIDHeader)

		// Validate client-provided ID; fall back to generated if invalid
		if !isValidRequestID(requestID) {
			requestID = generateRequestID()
		}

		// Generate a new one if not provided
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Add request ID to response header for client correlation
		w.Header().Set(RequestIDHeader, requestID)

		// Add request ID to context
		ctx := WithRequestID(r.Context(), requestID)

		// Pass the modified request to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateRequestID creates a unique request ID using random bytes.
// Format: 16 hex characters (8 bytes of randomness).
// Falls back to time-based ID if crypto/rand fails (extremely rare).
func generateRequestID() string {
	b := make([]byte, requestIDBytes)
	_, err := rand.Read(b)
	if err != nil {
		// Crypto/rand failure is extremely rare but shouldn't crash the server.
		// Fall back to time-based ID (sufficient for request correlation).
		Get().ErrorContext(context.Background(), "crypto/rand failed, using time-based fallback", "error", err)
		return fmt.Sprintf("%016x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// isValidRequestID ensures a client-supplied request ID is reasonably bounded
// and uses a safe character set to avoid log/header abuse.
func isValidRequestID(id string) bool {
	if id == "" {
		return false
	}
	if len(id) > maxRequestIDLength {
		return false
	}
	for _, c := range id {
		if (c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' {
			continue
		}
		return false
	}
	return true
}

// Middleware logs HTTP requests with timing information.
// It captures the request method, path, status code, and duration.
//
// This middleware should be placed after RequestIDMiddleware so that
// the request ID is available in the logs.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code.
		wrapped := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
			wroteHeader:    false,
		}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log the request
		duration := time.Since(start)
		logger := FromContext(r.Context())

		// Skip logging for health checks and static assets
		if r.URL.Path == "/api/health" || r.URL.Path == "/health" {
			return
		}

		logger.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.status,
			"duration_ms", duration.Milliseconds(),
			"client_ip", GetClientIP(r),
			"user_agent", r.UserAgent(),
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter

	status      int
	wroteHeader bool
}

// WriteHeader captures the status code before calling the underlying WriteHeader.
func (w *responseWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(code)
}

// Write calls the underlying Write and sets status to 200 if not already set.
func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(b)
	if err != nil {
		return n, fmt.Errorf("response write failed: %w", err)
	}
	return n, nil
}

// Unwrap returns the underlying ResponseWriter, supporting http.ResponseController.
func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// Hijack implements http.Hijacker for WebSocket support.
// This allows the logging middleware to be used with WebSocket endpoints.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		conn, rw, err := h.Hijack()
		if err != nil {
			return nil, nil, fmt.Errorf("hijack failed: %w", err)
		}
		return conn, rw, nil
	}
	return nil, nil, errors.New("underlying ResponseWriter does not implement http.Hijacker")
}
