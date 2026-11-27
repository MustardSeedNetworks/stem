// SPDX-License-Identifier: BUSL-1.1

package logging

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
)

// RedactedPlaceholder is the string used to mask sensitive values in
// logs and error messages. Hoisted as a const so format changes are a
// single-line edit and tests don't need to repeat the literal.
const RedactedPlaceholder = "[REDACTED]"

// Sensitive field patterns that should always be redacted.
// Comprehensive patterns for: passwords, tokens, API keys, secrets, SSNs, credit cards, etc.
func redactionPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		// Passwords and credentials
		regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[=:]\s*[^\s&]+`),
		regexp.MustCompile(`(?i)(token|auth|api[_-]?key|secret)\s*[=:]\s*[^\s&]+`),
		regexp.MustCompile(`(?i)(bearer\s+)\S+`),
		regexp.MustCompile(`(?i)(basic\s+)\S+`),

		// API keys and tokens - common formats
		regexp.MustCompile(`(?i)(api[_-]?key|apikey|access[_-]?key)\s*[=:]\s*[a-zA-Z0-9_\-\.]+`),
		regexp.MustCompile(`(?i)(client[_-]?secret|client_id)\s*[=:]\s*[a-zA-Z0-9_\-.]+`),
		regexp.MustCompile(`(?i)(oauth[_-]?token|refresh[_-]?token)\s*[=:]\s*[a-zA-Z0-9_\-\.]+`),

		// AWS-style keys
		regexp.MustCompile(`(?i)(aws[_-]?access[_-]?key[_-]?id|aws[_-]?secret[_-]?access[_-]?key)\s*[=:]\s*[A-Z0-9]+`),
		regexp.MustCompile(`AKIA[0-9A-Z]{16}`), // AWS Access Key ID pattern

		// GitHub/GitLab tokens
		regexp.MustCompile(`(?i)(github[_-]?token|gh[ps]_[a-zA-Z0-9_]{36,})`),
		regexp.MustCompile(`(?i)(gitlab[_-]?token|glpat-[a-zA-Z0-9_\-]{20,})`),

		// Private keys
		regexp.MustCompile(`-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----[^-]*-----END\s+(RSA\s+)?PRIVATE\s+KEY-----`),
		regexp.MustCompile(`(?i)(private[_-]?key|privatekey)\s*[=:]\s*[^\s&]+`),

		// Social Security Numbers (US) - XXX-XX-XXXX format
		regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),

		// Email addresses (for privacy)
		regexp.MustCompile(`\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Z|a-z]{2,}\b`),

		// JWT tokens (base64.base64.base64 format)
		regexp.MustCompile(`eyJ[a-zA-Z0-9_\-]*\.eyJ[a-zA-Z0-9_\-]*\.[a-zA-Z0-9_\-]*`),

		// Generic secrets and credentials
		regexp.MustCompile(`(?i)(credential|credentials|auth[_-]?token)\s*[=:]\s*[^\s&]+`),
		regexp.MustCompile(`(?i)(passphrase|pin|shared[_-]?secret)\s*[=:]\s*[^\s&]+`),
	}
}

// Sensitive header names (case-insensitive).
// Extended to include authentication, credential, and privacy-sensitive headers.
func redactionHeaderSet() map[string]bool {
	return map[string]bool{
		// Authentication and credentials
		"authorization":       true,
		"x-api-key":           true,
		"x-auth-token":        true,
		"cookie":              true,
		"set-cookie":          true,
		"x-csrf-token":        true,
		"x-xsrf-token":        true,
		"proxy-authorization": true,
		"x-access-token":      true,
		"x-refresh-token":     true,
		"x-session-token":     true,
		"x-client-secret":     true,
		"x-client-id":         true,
		"x-oauth-token":       true,
		"apikey":              true,
		"api-key":             true,
		// Privacy-sensitive headers
		"x-forwarded-for":     true,
		"x-real-ip":           true,
		"x-client-ip":         true,
		"cf-connecting-ip":    true,
		"true-client-ip":      true,
		"x-cluster-client-ip": true,
		"forwarded":           true,
	}
}

// digitRun matches a maximal run of digits joined by single spaces, tabs or
// dashes. isCardNumber decides whether the run holds a card number.
var digitRun = regexp.MustCompile(`\d(?:[ \t\-]?\d)*`)

// RedactString removes sensitive data from a string.
func RedactString(s string) string {
	for _, pattern := range redactionPatterns() {
		s = pattern.ReplaceAllString(s, RedactedPlaceholder)
	}
	return digitRun.ReplaceAllStringFunc(s, func(m string) string {
		if isCardNumber(m) {
			return RedactedPlaceholder
		}
		return m
	})
}

const (
	cardMinDigits = 13
	cardMaxDigits = 19
	decimalBase   = 10
)

// issuer is a card network's numbering: the leading digits in [lo, hi]
// (prefixLen digits) and the lengths it issues. Lengths are kept as ranges, a
// little wider than the networks issue, so an ambiguous window is redacted.
// There is no row for a leading 1: no network issues it, which is also why a
// unix timestamp (1.7e12 ms, 1.7e15 us, 1.7e18 ns) never reads as a card.
type issuer struct {
	prefixLen, lo, hi, minLen, maxLen int
}

func issuers() []issuer {
	return []issuer{
		{1, 4, 4, 13, 19},       // Visa
		{2, 51, 55, 16, 16},     // Mastercard
		{4, 2221, 2720, 16, 16}, // Mastercard 2-series
		{2, 34, 34, 15, 15},     // American Express
		{2, 37, 37, 15, 15},     // American Express
		{4, 6011, 6011, 16, 19}, // Discover
		{2, 65, 65, 16, 19},     // Discover
		{3, 644, 649, 16, 19},   // Discover
		{2, 35, 35, 16, 19},     // JCB
		{2, 30, 30, 14, 19},     // Diners Club
		{2, 36, 36, 14, 19},     // Diners Club
		{2, 38, 39, 14, 19},     // Diners Club
		{2, 62, 62, 16, 19},     // UnionPay
		{2, 50, 50, 12, 19},     // Maestro
		{2, 56, 58, 12, 19},     // Maestro
		{4, 6304, 6304, 12, 19}, // Maestro
		{4, 6759, 6759, 12, 19}, // Maestro
		{4, 6761, 6763, 12, 19}, // Maestro
	}
}

// isCardNumber reports whether a maximal digit run contains a payment card
// number: any window of 13-19 consecutive digits that has an issuer prefix and
// a valid Luhn checksum. Every window is tried, so digits adjacent to a card (a
// year, an ID prefix) cannot hide it, and the whole run is redacted when one
// window qualifies. Luhn alone passes one random run in ten; the issuer prefix
// is what keeps timestamps and most run IDs readable.
func isCardNumber(run string) bool {
	digits := make([]int, 0, len(run))
	for _, r := range run {
		if r >= '0' && r <= '9' {
			digits = append(digits, int(r-'0'))
		}
	}
	for size := cardMinDigits; size <= min(cardMaxDigits, len(digits)); size++ {
		for start := 0; start+size <= len(digits); start++ {
			if w := digits[start : start+size]; hasIssuerPrefix(w) && luhnValid(w) {
				return true
			}
		}
	}
	return false
}

func hasIssuerPrefix(window []int) bool {
	for _, is := range issuers() {
		if len(window) < is.minLen || len(window) > is.maxLen {
			continue
		}
		prefix := 0
		for _, d := range window[:is.prefixLen] {
			prefix = prefix*decimalBase + d
		}
		if prefix >= is.lo && prefix <= is.hi {
			return true
		}
	}
	return false
}

func luhnValid(digits []int) bool {
	sum := 0
	for i, d := range slices.Backward(digits) {
		if (len(digits)-1-i)%2 == 1 {
			d *= 2
			d = d/decimalBase + d%decimalBase
		}
		sum += d
	}
	return sum%10 == 0
}

// RedactHeaders returns a map of headers with sensitive values redacted.
func RedactHeaders(headers http.Header) map[string]string {
	redacted := make(map[string]string)
	headerSet := redactionHeaderSet()
	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		if headerSet[lowerKey] {
			redacted[key] = RedactedPlaceholder
		} else {
			redacted[key] = strings.Join(values, ", ")
		}
	}
	return redacted
}

// RedactMap redacts sensitive fields in a map (useful for JSON logging).
func RedactMap(data map[string]any) map[string]any {
	redacted := make(map[string]any)
	for key, value := range data {
		lowerKey := strings.ToLower(key)
		if strings.Contains(lowerKey, FieldPassword) ||
			strings.Contains(lowerKey, "secret") ||
			strings.Contains(lowerKey, "token") ||
			strings.Contains(lowerKey, "key") ||
			strings.Contains(lowerKey, "auth") {
			redacted[key] = RedactedPlaceholder
		} else {
			// For string values, apply pattern-based redaction
			if strVal, ok := value.(string); ok {
				redacted[key] = RedactString(strVal)
			} else {
				redacted[key] = value
			}
		}
	}
	return redacted
}

// Logf is a safe logging function that redacts sensitive data.
// Note: Prefer using slog directly with the RedactingHandler for new code.
func Logf(format string, args ...any) {
	// Convert args to strings and redact.
	redactedArgs := make([]any, len(args))
	for i, arg := range args {
		switch v := arg.(type) {
		case string:
			redactedArgs[i] = RedactString(v)
		case http.Header:
			redactedArgs[i] = RedactHeaders(v)
		case map[string]any:
			redactedArgs[i] = RedactMap(v)
		default:
			redactedArgs[i] = arg
		}
	}
	Get().InfoContext(context.Background(), fmt.Sprintf(format, redactedArgs...))
}

// SafeError creates a safe error message with redacted content.
func SafeError(err error, context string) error {
	if err == nil {
		return nil
	}
	redactedMsg := RedactString(err.Error())
	return fmt.Errorf("%s: %s", context, redactedMsg)
}

// LogRequest logs an HTTP request with sensitive data redacted.
// Note: Prefer using Middleware for request logging in new code.
func LogRequest(r *http.Request, message string) {
	Get().InfoContext(r.Context(), message,
		"method", r.Method,
		"path", r.URL.Path,
		"client_ip", GetClientIP(r),
		"headers", RedactHeaders(r.Header),
	)
}

// GetClientIP extracts client IP from request for logging and display purposes.
//
// SECURITY WARNING:
// This function returns UNTRUSTED IP addresses that can be spoofed by malicious clients.
// It checks X-Forwarded-For and X-Real-IP headers which are trivially spoofed.
//
// TRUST MODEL:
// - X-Forwarded-For: UNTRUSTED - Any client can set this header to any value
// - X-Real-IP: UNTRUSTED - Any client can set this header to any value
// - r.RemoteAddr: TRUSTED - This is the actual TCP connection source
//
// USE CASES:
// - OK for logging/debugging (helps in reverse proxy scenarios)
// - OK for display in admin dashboards
// - NEVER use for rate limiting
// - NEVER use for access control or security decisions
// - NEVER use for ban lists or IP blocking.
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (UNTRUSTED - can be spoofed by clients).
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			clientIP := strings.TrimSpace(parts[0])
			Get().DebugContext(r.Context(), "Request includes X-Forwarded-For header (UNTRUSTED)",
				"xff_value", xff,
				"parsed_ip", clientIP,
				"remote_addr", r.RemoteAddr,
				"security_note", "XFF can be spoofed - only use for logging, not security decisions")
			return clientIP
		}
	}

	// Check X-Real-IP header (UNTRUSTED - can be spoofed by clients).
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		Get().DebugContext(r.Context(), "Request includes X-Real-IP header (UNTRUSTED)",
			"xri_value", xri,
			"remote_addr", r.RemoteAddr,
			"security_note", "X-Real-IP can be spoofed - only use for logging, not security decisions")
		return xri
	}

	// Fall back to RemoteAddr (TRUSTED - the only reliable source)
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		addr = addr[:idx]
	}
	return addr
}
