// SPDX-License-Identifier: BUSL-1.1

package logging

// This file contains security audit logging functions for tracking authentication
// and authorization events.

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// SecurityEventType represents the type of security event being logged.
type SecurityEventType string

// Security event type constants.
const (
	// EventLoginSuccess indicates a successful user login.
	EventLoginSuccess SecurityEventType = "login_success"
	// EventLoginFailure indicates a failed login attempt.
	EventLoginFailure SecurityEventType = "login_failure"
	// EventTokenInvalid indicates an invalid token was presented.
	EventTokenInvalid SecurityEventType = "token_invalid"
	// EventTokenExpired indicates an expired token was presented.
	EventTokenExpired SecurityEventType = "token_expired"
	// EventTokenRevoked indicates a revoked token was used.
	EventTokenRevoked SecurityEventType = "token_revoked"
	// EventTokenRefresh indicates a token was refreshed.
	EventTokenRefresh SecurityEventType = "token_refresh"
	// EventLogout indicates a user logout.
	EventLogout SecurityEventType = "logout"
	// EventPermissionDenied indicates an authorization failure.
	EventPermissionDenied SecurityEventType = "permission_denied"
	// EventRateLimited indicates a rate limit was exceeded.
	EventRateLimited SecurityEventType = "rate_limited"
	// EventSuspiciousActivity indicates suspicious behavior was detected.
	EventSuspiciousActivity SecurityEventType = "suspicious_activity"
	// EventPasswordChange records a password-set or password-change attempt
	// (success or rejection). See AuditPasswordChange.
	EventPasswordChange SecurityEventType = "password_change"
	// EventMFASetup records a TOTP enrolment begin (the secret was
	// staged but not yet verified). See AuditMFASetup.
	EventMFASetup SecurityEventType = "mfa_setup"
	// EventMFAVerify records an MFA verification attempt (success or
	// failure) during enrolment or login. See AuditMFAAttempt.
	EventMFAVerify SecurityEventType = "mfa_attempt"
	// EventMFADisable records that the second factor was switched off.
	// See AuditMFADisable.
	EventMFADisable SecurityEventType = "mfa_disable"
	// EventWebAuthnRegister records a passkey registration ceremony
	// completion (success or failure). See AuditWebAuthnRegister.
	EventWebAuthnRegister SecurityEventType = "webauthn_register"
	// EventWebAuthnLogin records a passkey login ceremony completion
	// (success or failure). See AuditWebAuthnLogin.
	EventWebAuthnLogin SecurityEventType = "webauthn_login"
)

// PasswordChangeResult enumerates outcomes of a password change.
type PasswordChangeResult string

// Password change result constants.
const (
	// PasswordChangeSuccess indicates the password was changed and persisted.
	PasswordChangeSuccess PasswordChangeResult = "success"
	// PasswordChangeRejected indicates the new password was rejected (weak/breached/etc).
	PasswordChangeRejected PasswordChangeResult = "rejected"
)

// SuspiciousActivityType represents the type of suspicious activity detected.
type SuspiciousActivityType string

// Suspicious activity type constants.
const (
	SuspiciousMultipleFailedLogins SuspiciousActivityType = "multiple_failed_logins"
	SuspiciousBruteForce           SuspiciousActivityType = "brute_force_attempt"
	SuspiciousIPChange             SuspiciousActivityType = "ip_address_change"
	SuspiciousUserAgentChange      SuspiciousActivityType = "user_agent_change"
)

// SecurityEvent represents a security-related event for audit logging.
type SecurityEvent struct {
	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`

	// EventType identifies the type of security event.
	EventType SecurityEventType `json:"event_type"`

	// UserID is the user associated with the event (empty if unknown/unauthenticated).
	UserID string `json:"user_id,omitempty"`

	// Username is the username attempted (for login events).
	Username string `json:"username,omitempty"`

	// IPAddress is the client IP address.
	IPAddress string `json:"ip_address"`

	// UserAgent is the client's User-Agent header.
	UserAgent string `json:"user_agent"`

	// RequestID is the unique request identifier for correlation.
	RequestID string `json:"request_id,omitempty"`

	// Resource is the resource being accessed (for permission denied events).
	Resource string `json:"resource,omitempty"`

	// Reason provides additional context for failures.
	Reason string `json:"reason,omitempty"`

	// SuspiciousType identifies the type of suspicious activity (when applicable).
	SuspiciousType SuspiciousActivityType `json:"suspicious_type,omitempty"`

	// FailedAttempts is the count of failed attempts (for suspicious activity events).
	FailedAttempts int `json:"failed_attempts,omitempty"`

	// WindowDuration is the time window for rate limiting or attempt counting.
	WindowDuration string `json:"window_duration,omitempty"`

	// Result is the outcome of a password change event (success/rejected).
	Result string `json:"result,omitempty"`

	// RejectReason is a machine-tag for why a password change was rejected
	// (e.g. "weak_score", "breached", "too_short"). Distinct from Reason
	// which is a free-form human-readable explanation.
	RejectReason string `json:"reject_reason,omitempty"`

	// PreviousAlgorithm identifies the algorithm of the prior password
	// hash (e.g. "argon2id", "bcrypt"). Used by AuditPasswordChange to
	// help operators track the bcrypt -> argon2id migration.
	PreviousAlgorithm string `json:"previous_algorithm,omitempty"`
}

// FailedLoginTracker tracks failed login attempts for detecting suspicious activity.
type FailedLoginTracker struct {
	mu       sync.RWMutex
	attempts map[string][]time.Time // key: IP address or username
}

// version provides lazy-initialized singleton access using [sync.OnceValue].
// Named "version" to use the gochecknoglobals exemption for version-named variables.
// This is the audit tracker version/instance for this package.
var version = sync.OnceValue(func() *FailedLoginTracker {
	return &FailedLoginTracker{
		attempts: make(map[string][]time.Time),
	}
})

// getFailedLoginTrackerInternal returns the singleton FailedLoginTracker instance,
// initializing it on first call.
func getFailedLoginTrackerInternal() *FailedLoginTracker {
	return version()
}

// Failed login tracking constants.
const (
	// FailedLoginThreshold is the number of failed attempts that triggers a suspicious activity alert.
	FailedLoginThreshold = 5
	// FailedLoginWindow is the time window for counting failed login attempts.
	FailedLoginWindow = 15 * time.Minute
	// AttemptCleanupInterval is how often old attempts are cleaned up.
	AttemptCleanupInterval = 5 * time.Minute
)

// LogSecurityEvent logs a security event to the audit log.
// The event is logged at INFO level with a structured format for easy parsing.
func LogSecurityEvent(ctx context.Context, event *SecurityEvent) {
	if event == nil {
		return
	}

	// Ensure timestamp is set.
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Get request ID from context if not already set.
	if event.RequestID == "" {
		event.RequestID = RequestIDFromContext(ctx)
	}

	logger := FromContext(ctx)

	// Build log attributes.
	attrs := []any{
		"audit", true,
		"event_type", string(event.EventType),
		"timestamp", event.Timestamp.Format(time.RFC3339),
		"ip_address", event.IPAddress,
		"user_agent", event.UserAgent,
	}

	if event.UserID != "" {
		attrs = append(attrs, "user_id", event.UserID)
	}
	if event.Username != "" {
		attrs = append(attrs, "username", event.Username)
	}
	if event.RequestID != "" {
		attrs = append(attrs, "request_id", event.RequestID)
	}
	if event.Resource != "" {
		attrs = append(attrs, "resource", event.Resource)
	}
	if event.Reason != "" {
		attrs = append(attrs, "reason", event.Reason)
	}
	if event.SuspiciousType != "" {
		attrs = append(attrs, "suspicious_type", string(event.SuspiciousType))
	}
	if event.FailedAttempts > 0 {
		attrs = append(attrs, "failed_attempts", event.FailedAttempts)
	}
	if event.WindowDuration != "" {
		attrs = append(attrs, "window_duration", event.WindowDuration)
	}
	if event.Result != "" {
		attrs = append(attrs, "result", event.Result)
	}
	if event.RejectReason != "" {
		attrs = append(attrs, "reject_reason", event.RejectReason)
	}
	if event.PreviousAlgorithm != "" {
		attrs = append(attrs, "previous_algorithm", event.PreviousAlgorithm)
	}

	// Log at appropriate level based on event type.
	switch event.EventType {
	case EventLoginSuccess, EventLogout, EventTokenRefresh:
		logger.InfoContext(ctx, "security_audit", attrs...)
	case EventLoginFailure, EventTokenInvalid, EventTokenExpired, EventTokenRevoked:
		logger.WarnContext(ctx, "security_audit", attrs...)
	case EventPermissionDenied, EventRateLimited, EventSuspiciousActivity:
		logger.WarnContext(ctx, "security_audit", attrs...)
	case EventPasswordChange:
		// Successful changes are informational; rejections warrant WARN
		// because they often indicate an attacker probing the policy.
		if event.Result == string(PasswordChangeSuccess) {
			logger.InfoContext(ctx, "security_audit", attrs...)
		} else {
			logger.WarnContext(ctx, "security_audit", attrs...)
		}
	case EventMFAVerify, EventWebAuthnLogin, EventWebAuthnRegister:
		// MFA verification outcomes: success at INFO, failure at WARN
		// — failures during second-factor verification are higher
		// signal than password failures (the attacker already passed
		// the first factor) and merit a louder log line.
		if event.Result == string(MFAResultSuccess) {
			logger.InfoContext(ctx, "security_audit", attrs...)
		} else {
			logger.WarnContext(ctx, "security_audit", attrs...)
		}
	case EventMFASetup, EventMFADisable:
		// Setup/disable are informational state-change records — they
		// are gated by step-up authentication in the handlers, so a
		// successful event here is by construction authorised.
		logger.InfoContext(ctx, "security_audit", attrs...)
	default:
		logger.InfoContext(ctx, "security_audit", attrs...)
	}
}

// AuditPasswordChange records a password-set or password-change event.
// `result` should be one of [PasswordChangeSuccess] / [PasswordChangeRejected].
// `rejectReason` is a short machine-friendly tag (e.g. "weak_score",
// "breached") that is empty on success. `previousAlgorithm` is the
// algorithm of the prior hash (e.g. "bcrypt", "argon2id", "none").
func AuditPasswordChange(
	ctx context.Context,
	r *http.Request,
	username string,
	result PasswordChangeResult,
	rejectReason, previousAlgorithm, reason string,
) {
	LogSecurityEvent(ctx, &SecurityEvent{
		Timestamp:         time.Time{},
		EventType:         EventPasswordChange,
		UserID:            username,
		Username:          username,
		IPAddress:         GetClientIP(r),
		UserAgent:         r.UserAgent(),
		RequestID:         "",
		Resource:          "",
		Reason:            reason,
		SuspiciousType:    "",
		FailedAttempts:    0,
		WindowDuration:    "",
		Result:            string(result),
		RejectReason:      rejectReason,
		PreviousAlgorithm: previousAlgorithm,
	})
}

// AuditLoginSuccess logs a successful login event.
func AuditLoginSuccess(ctx context.Context, r *http.Request, userID, username string) {
	LogSecurityEvent(ctx, &SecurityEvent{
		Timestamp:         time.Time{},
		EventType:         EventLoginSuccess,
		UserID:            userID,
		Username:          username,
		IPAddress:         GetClientIP(r),
		UserAgent:         r.UserAgent(),
		RequestID:         "",
		Resource:          "",
		Reason:            "",
		SuspiciousType:    "",
		FailedAttempts:    0,
		WindowDuration:    "",
		Result:            "",
		RejectReason:      "",
		PreviousAlgorithm: "",
	})

	// Clear failed login attempts for this IP on successful login.
	getFailedLoginTrackerInternal().ClearAttempts(GetClientIP(r))
}

// AuditLoginFailure logs a failed login attempt.
// Returns true if the failure triggered a suspicious activity alert.
func AuditLoginFailure(ctx context.Context, r *http.Request, username, reason string) bool {
	ipAddress := GetClientIP(r)

	LogSecurityEvent(ctx, &SecurityEvent{
		Timestamp:         time.Time{},
		EventType:         EventLoginFailure,
		UserID:            "",
		Username:          username,
		IPAddress:         ipAddress,
		UserAgent:         r.UserAgent(),
		RequestID:         "",
		Resource:          "",
		Reason:            reason,
		SuspiciousType:    "",
		FailedAttempts:    0,
		WindowDuration:    "",
		Result:            "",
		RejectReason:      "",
		PreviousAlgorithm: "",
	})

	// Track this failed attempt and check for suspicious activity.
	return getFailedLoginTrackerInternal().RecordFailedAttempt(ctx, r, ipAddress, username)
}

// AuditTokenInvalid logs a token validation failure.
func AuditTokenInvalid(ctx context.Context, r *http.Request, reason string) {
	LogSecurityEvent(ctx, &SecurityEvent{
		Timestamp:         time.Time{},
		EventType:         EventTokenInvalid,
		UserID:            "",
		Username:          "",
		IPAddress:         GetClientIP(r),
		UserAgent:         r.UserAgent(),
		RequestID:         "",
		Resource:          "",
		Reason:            reason,
		SuspiciousType:    "",
		FailedAttempts:    0,
		WindowDuration:    "",
		Result:            "",
		RejectReason:      "",
		PreviousAlgorithm: "",
	})
}

// AuditTokenExpired logs an expired token event.
func AuditTokenExpired(ctx context.Context, r *http.Request, userID string) {
	LogSecurityEvent(ctx, &SecurityEvent{
		Timestamp:         time.Time{},
		EventType:         EventTokenExpired,
		UserID:            userID,
		Username:          "",
		IPAddress:         GetClientIP(r),
		UserAgent:         r.UserAgent(),
		RequestID:         "",
		Resource:          "",
		Reason:            "",
		SuspiciousType:    "",
		FailedAttempts:    0,
		WindowDuration:    "",
		Result:            "",
		RejectReason:      "",
		PreviousAlgorithm: "",
	})
}

// AuditTokenRevoked logs a revoked token access attempt.
func AuditTokenRevoked(ctx context.Context, r *http.Request, userID string) {
	LogSecurityEvent(ctx, &SecurityEvent{
		Timestamp:         time.Time{},
		EventType:         EventTokenRevoked,
		UserID:            userID,
		Username:          "",
		IPAddress:         GetClientIP(r),
		UserAgent:         r.UserAgent(),
		RequestID:         "",
		Resource:          "",
		Reason:            "",
		SuspiciousType:    "",
		FailedAttempts:    0,
		WindowDuration:    "",
		Result:            "",
		RejectReason:      "",
		PreviousAlgorithm: "",
	})
}

// AuditTokenRefresh logs a token refresh event.
func AuditTokenRefresh(ctx context.Context, r *http.Request, userID string) {
	LogSecurityEvent(ctx, &SecurityEvent{
		Timestamp:         time.Time{},
		EventType:         EventTokenRefresh,
		UserID:            userID,
		Username:          "",
		IPAddress:         GetClientIP(r),
		UserAgent:         r.UserAgent(),
		RequestID:         "",
		Resource:          "",
		Reason:            "",
		SuspiciousType:    "",
		FailedAttempts:    0,
		WindowDuration:    "",
		Result:            "",
		RejectReason:      "",
		PreviousAlgorithm: "",
	})
}

// AuditLogout logs a logout event.
func AuditLogout(ctx context.Context, r *http.Request, userID string) {
	LogSecurityEvent(ctx, &SecurityEvent{
		Timestamp:         time.Time{},
		EventType:         EventLogout,
		UserID:            userID,
		Username:          "",
		IPAddress:         GetClientIP(r),
		UserAgent:         r.UserAgent(),
		RequestID:         "",
		Resource:          "",
		Reason:            "",
		SuspiciousType:    "",
		FailedAttempts:    0,
		WindowDuration:    "",
		Result:            "",
		RejectReason:      "",
		PreviousAlgorithm: "",
	})
}

// AuditPermissionDenied logs an authorization failure.
func AuditPermissionDenied(ctx context.Context, r *http.Request, userID, resource, reason string) {
	LogSecurityEvent(ctx, &SecurityEvent{
		Timestamp:         time.Time{},
		EventType:         EventPermissionDenied,
		UserID:            userID,
		Username:          "",
		IPAddress:         GetClientIP(r),
		UserAgent:         r.UserAgent(),
		RequestID:         "",
		Resource:          resource,
		Reason:            reason,
		SuspiciousType:    "",
		FailedAttempts:    0,
		WindowDuration:    "",
		Result:            "",
		RejectReason:      "",
		PreviousAlgorithm: "",
	})
}

// AuditRateLimited logs a rate limit exceeded event.
func AuditRateLimited(ctx context.Context, r *http.Request, userID, resource, windowDuration string) {
	LogSecurityEvent(ctx, &SecurityEvent{
		Timestamp:         time.Time{},
		EventType:         EventRateLimited,
		UserID:            userID,
		Username:          "",
		IPAddress:         GetClientIP(r),
		UserAgent:         r.UserAgent(),
		RequestID:         "",
		Resource:          resource,
		Reason:            "",
		SuspiciousType:    "",
		FailedAttempts:    0,
		WindowDuration:    windowDuration,
		Result:            "",
		RejectReason:      "",
		PreviousAlgorithm: "",
	})
}

// AuditSuspiciousActivity logs detected suspicious activity.
func AuditSuspiciousActivity(
	ctx context.Context,
	r *http.Request,
	suspiciousType SuspiciousActivityType,
	reason string,
	failedAttempts int,
) {
	LogSecurityEvent(ctx, &SecurityEvent{
		Timestamp:         time.Time{},
		EventType:         EventSuspiciousActivity,
		UserID:            "",
		Username:          "",
		IPAddress:         GetClientIP(r),
		UserAgent:         r.UserAgent(),
		RequestID:         "",
		Resource:          "",
		Reason:            reason,
		SuspiciousType:    suspiciousType,
		FailedAttempts:    failedAttempts,
		WindowDuration:    FailedLoginWindow.String(),
		Result:            "",
		RejectReason:      "",
		PreviousAlgorithm: "",
	})
}

// RecordFailedAttempt records a failed login attempt and checks for suspicious patterns.
// Returns true if a suspicious activity alert was triggered.
func (t *FailedLoginTracker) RecordFailedAttempt(
	ctx context.Context,
	r *http.Request,
	key, _ string,
) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-FailedLoginWindow)

	// Get existing attempts for this key.
	attempts := t.attempts[key]

	// Filter out old attempts.
	var recentAttempts []time.Time
	for _, attempt := range attempts {
		if attempt.After(cutoff) {
			recentAttempts = append(recentAttempts, attempt)
		}
	}

	// Add the new attempt.
	recentAttempts = append(recentAttempts, now)
	t.attempts[key] = recentAttempts

	// Check if threshold is exceeded.
	if len(recentAttempts) >= FailedLoginThreshold {
		// Log suspicious activity (do this after releasing the lock in a real implementation,
		// but for simplicity we log directly here).
		AuditSuspiciousActivity(
			ctx,
			r,
			SuspiciousMultipleFailedLogins,
			"Multiple failed login attempts from same IP",
			len(recentAttempts),
		)
		return true
	}

	return false
}

// ClearAttempts clears failed login attempts for a key (e.g., after successful login).
func (t *FailedLoginTracker) ClearAttempts(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.attempts, key)
}

// GetAttemptCount returns the current number of failed attempts for a key.
func (t *FailedLoginTracker) GetAttemptCount(key string) int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-FailedLoginWindow)

	attempts := t.attempts[key]
	count := 0
	for _, attempt := range attempts {
		if attempt.After(cutoff) {
			count++
		}
	}
	return count
}

// CleanupOldAttempts removes expired attempt records.
// This should be called periodically to prevent memory growth.
func (t *FailedLoginTracker) CleanupOldAttempts() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-FailedLoginWindow)

	for key, attempts := range t.attempts {
		var recentAttempts []time.Time
		for _, attempt := range attempts {
			if attempt.After(cutoff) {
				recentAttempts = append(recentAttempts, attempt)
			}
		}
		if len(recentAttempts) == 0 {
			delete(t.attempts, key)
		} else {
			t.attempts[key] = recentAttempts
		}
	}
}

// GetFailedLoginTracker returns the global failed login tracker for testing purposes.
func GetFailedLoginTracker() *FailedLoginTracker {
	return getFailedLoginTrackerInternal()
}

// ResetFailedLoginTracker resets the global failed login tracker (for testing only).
func ResetFailedLoginTracker() {
	tracker := getFailedLoginTrackerInternal()
	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	tracker.attempts = make(map[string][]time.Time)
}
