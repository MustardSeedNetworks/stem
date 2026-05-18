// SPDX-License-Identifier: BUSL-1.1

package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// testLogHandler captures log output for testing.
type testLogHandler struct {
	buf    *bytes.Buffer
	format string // "json" or "text"
}

func newTestLogHandler(format string) *testLogHandler {
	return &testLogHandler{
		buf:    &bytes.Buffer{},
		format: format,
	}
}

func (h *testLogHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *testLogHandler) Handle(_ context.Context, r slog.Record) error {
	if h.format == "json" {
		entry := make(map[string]any)
		entry["level"] = r.Level.String()
		entry["msg"] = r.Message
		entry["time"] = r.Time.Format(time.RFC3339)
		r.Attrs(func(a slog.Attr) bool {
			entry[a.Key] = a.Value.Any()
			return true
		})
		data, _ := json.Marshal(entry)
		h.buf.Write(data)
		h.buf.WriteByte('\n')
	} else {
		h.buf.WriteString(r.Level.String())
		h.buf.WriteString(" ")
		h.buf.WriteString(r.Message)
		r.Attrs(func(a slog.Attr) bool {
			h.buf.WriteString(" ")
			h.buf.WriteString(a.Key)
			h.buf.WriteString("=")
			h.buf.WriteString(a.Value.String())
			return true
		})
		h.buf.WriteByte('\n')
	}
	return nil
}

func (h *testLogHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *testLogHandler) WithGroup(_ string) slog.Handler {
	return h
}

func (h *testLogHandler) String() string {
	return h.buf.String()
}

func (h *testLogHandler) Reset() {
	h.buf.Reset()
}

// setupTestLogger sets up a test logger that captures output.
func setupTestLogger() *testLogHandler {
	handler := newTestLogHandler("json")
	logger := slog.New(handler)

	slog.SetDefault(logger)

	return handler
}

// createTestRequest creates a test HTTP request with common headers.
func createTestRequest(method, path string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "192.168.1.100:12345"
	return req
}

func TestLogSecurityEvent(t *testing.T) {
	handler := setupTestLogger()
	defer handler.Reset()

	ctx := context.Background()
	ctx = WithRequestID(ctx, "test-request-123")

	event := &SecurityEvent{
		Timestamp:      time.Time{},
		EventType:      EventLoginSuccess,
		UserID:         "user-456",
		Username:       "testuser",
		IPAddress:      "192.168.1.100",
		UserAgent:      "TestAgent/1.0",
		RequestID:      "",
		Resource:       "",
		Reason:         "",
		SuspiciousType: "",
		FailedAttempts: 0,
		WindowDuration: "",
	}

	LogSecurityEvent(ctx, event)

	output := handler.String()

	// Verify the log contains expected fields.
	if !strings.Contains(output, "security_audit") {
		t.Error("expected log message to contain 'security_audit'")
	}
	if !strings.Contains(output, "login_success") {
		t.Error("expected log to contain 'login_success' event type")
	}
	if !strings.Contains(output, "user-456") {
		t.Error("expected log to contain user ID")
	}
	if !strings.Contains(output, "testuser") {
		t.Error("expected log to contain username")
	}
	if !strings.Contains(output, "192.168.1.100") {
		t.Error("expected log to contain IP address")
	}
	if !strings.Contains(output, "TestAgent/1.0") {
		t.Error("expected log to contain user agent")
	}
	if !strings.Contains(output, "test-request-123") {
		t.Error("expected log to contain request ID from context")
	}
}

func TestLogSecurityEvent_NilEvent(t *testing.T) {
	handler := setupTestLogger()
	defer handler.Reset()

	// Should not panic or log anything.
	LogSecurityEvent(context.Background(), nil)

	if handler.String() != "" {
		t.Error("expected no log output for nil event")
	}
}

func TestLogSecurityEvent_SetsTimestamp(t *testing.T) {
	handler := setupTestLogger()
	defer handler.Reset()

	ctx := context.Background()
	event := &SecurityEvent{
		Timestamp:      time.Time{},
		EventType:      EventLoginSuccess,
		UserID:         "",
		Username:       "",
		IPAddress:      "127.0.0.1",
		UserAgent:      "Test",
		RequestID:      "",
		Resource:       "",
		Reason:         "",
		SuspiciousType: "",
		FailedAttempts: 0,
		WindowDuration: "",
	}

	before := time.Now()
	LogSecurityEvent(ctx, event)
	after := time.Now()

	// Verify timestamp was set.
	if event.Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
	if event.Timestamp.Before(before) || event.Timestamp.After(after) {
		t.Error("expected timestamp to be within test bounds")
	}
}

func TestAuditLoginSuccess(t *testing.T) {
	handler := setupTestLogger()
	ResetFailedLoginTracker()
	defer handler.Reset()

	req := createTestRequest(http.MethodPost, "/api/auth/login")
	ctx := context.Background()

	AuditLoginSuccess(ctx, req, "user-123", "admin")

	output := handler.String()
	if !strings.Contains(output, "login_success") {
		t.Error("expected login_success event type")
	}
	if !strings.Contains(output, "user-123") {
		t.Error("expected user ID in log")
	}
	if !strings.Contains(output, "admin") {
		t.Error("expected username in log")
	}
}

func TestAuditLoginSuccess_ClearsFailedAttempts(t *testing.T) {
	ResetFailedLoginTracker()
	handler := setupTestLogger()
	defer handler.Reset()

	req := createTestRequest(http.MethodPost, "/api/auth/login")
	ctx := context.Background()
	ipAddress := GetClientIP(req)

	// Record some failed attempts first.
	for range 3 {
		AuditLoginFailure(ctx, req, "testuser", "wrong password")
	}

	// Verify attempts were recorded.
	if GetFailedLoginTracker().GetAttemptCount(ipAddress) != 3 {
		t.Error("expected 3 failed attempts recorded")
	}

	// Successful login should clear attempts.
	AuditLoginSuccess(ctx, req, "user-123", "testuser")

	if GetFailedLoginTracker().GetAttemptCount(ipAddress) != 0 {
		t.Error("expected failed attempts to be cleared after successful login")
	}
}

func TestAuditLoginFailure(t *testing.T) {
	handler := setupTestLogger()
	ResetFailedLoginTracker()
	defer handler.Reset()

	req := createTestRequest(http.MethodPost, "/api/auth/login")
	ctx := context.Background()

	triggered := AuditLoginFailure(ctx, req, "baduser", "invalid credentials")

	output := handler.String()
	if !strings.Contains(output, "login_failure") {
		t.Error("expected login_failure event type")
	}
	if !strings.Contains(output, "baduser") {
		t.Error("expected username in log")
	}
	if !strings.Contains(output, "invalid credentials") {
		t.Error("expected reason in log")
	}
	if triggered {
		t.Error("first failure should not trigger suspicious activity")
	}
}

func TestAuditLoginFailure_TriggersSuspiciousActivity(t *testing.T) {
	handler := setupTestLogger()
	ResetFailedLoginTracker()
	defer handler.Reset()

	req := createTestRequest(http.MethodPost, "/api/auth/login")
	ctx := context.Background()

	// Generate enough failures to trigger suspicious activity.
	var triggered bool
	for range FailedLoginThreshold {
		triggered = AuditLoginFailure(ctx, req, "attacker", "wrong password")
	}

	if !triggered {
		t.Error("expected suspicious activity to be triggered after threshold")
	}

	output := handler.String()
	if !strings.Contains(output, "suspicious_activity") {
		t.Error("expected suspicious_activity event in log")
	}
	if !strings.Contains(output, "multiple_failed_logins") {
		t.Error("expected multiple_failed_logins suspicious type")
	}
}

func TestAuditTokenInvalid(t *testing.T) {
	handler := setupTestLogger()
	defer handler.Reset()

	req := createTestRequest(http.MethodGet, "/api/test/start")
	ctx := context.Background()

	AuditTokenInvalid(ctx, req, "malformed JWT")

	output := handler.String()
	if !strings.Contains(output, "token_invalid") {
		t.Error("expected token_invalid event type")
	}
	if !strings.Contains(output, "malformed JWT") {
		t.Error("expected reason in log")
	}
}

func TestAuditTokenExpired(t *testing.T) {
	handler := setupTestLogger()
	defer handler.Reset()

	req := createTestRequest(http.MethodGet, "/api/test/start")
	ctx := context.Background()

	AuditTokenExpired(ctx, req, "user-expired")

	output := handler.String()
	if !strings.Contains(output, "token_expired") {
		t.Error("expected token_expired event type")
	}
	if !strings.Contains(output, "user-expired") {
		t.Error("expected user ID in log")
	}
}

func TestAuditTokenRevoked(t *testing.T) {
	handler := setupTestLogger()
	defer handler.Reset()

	req := createTestRequest(http.MethodGet, "/api/test/start")
	ctx := context.Background()

	AuditTokenRevoked(ctx, req, "user-revoked")

	output := handler.String()
	if !strings.Contains(output, "token_revoked") {
		t.Error("expected token_revoked event type")
	}
	if !strings.Contains(output, "user-revoked") {
		t.Error("expected user ID in log")
	}
}

func TestAuditTokenRefresh(t *testing.T) {
	handler := setupTestLogger()
	defer handler.Reset()

	req := createTestRequest(http.MethodPost, "/api/auth/refresh")
	ctx := context.Background()

	AuditTokenRefresh(ctx, req, "user-refresh")

	output := handler.String()
	if !strings.Contains(output, "token_refresh") {
		t.Error("expected token_refresh event type")
	}
	if !strings.Contains(output, "user-refresh") {
		t.Error("expected user ID in log")
	}
}

func TestAuditLogout(t *testing.T) {
	handler := setupTestLogger()
	defer handler.Reset()

	req := createTestRequest(http.MethodPost, "/api/auth/logout")
	ctx := context.Background()

	AuditLogout(ctx, req, "user-logout")

	output := handler.String()
	if !strings.Contains(output, "logout") {
		t.Error("expected logout event type")
	}
	if !strings.Contains(output, "user-logout") {
		t.Error("expected user ID in log")
	}
}

func TestAuditPermissionDenied(t *testing.T) {
	handler := setupTestLogger()
	defer handler.Reset()

	req := createTestRequest(http.MethodPost, "/api/admin/settings")
	ctx := context.Background()

	AuditPermissionDenied(ctx, req, "user-denied", "/api/admin/settings", "insufficient privileges")

	output := handler.String()
	if !strings.Contains(output, "permission_denied") {
		t.Error("expected permission_denied event type")
	}
	if !strings.Contains(output, "user-denied") {
		t.Error("expected user ID in log")
	}
	if !strings.Contains(output, "/api/admin/settings") {
		t.Error("expected resource in log")
	}
	if !strings.Contains(output, "insufficient privileges") {
		t.Error("expected reason in log")
	}
}

func TestAuditRateLimited(t *testing.T) {
	handler := setupTestLogger()
	defer handler.Reset()

	req := createTestRequest(http.MethodPost, "/api/auth/login")
	ctx := context.Background()

	AuditRateLimited(ctx, req, "user-limited", "/api/auth/login", "1m")

	output := handler.String()
	if !strings.Contains(output, "rate_limited") {
		t.Error("expected rate_limited event type")
	}
	if !strings.Contains(output, "user-limited") {
		t.Error("expected user ID in log")
	}
	if !strings.Contains(output, "/api/auth/login") {
		t.Error("expected resource in log")
	}
	if !strings.Contains(output, "1m") {
		t.Error("expected window duration in log")
	}
}

func TestAuditSuspiciousActivity(t *testing.T) {
	handler := setupTestLogger()
	defer handler.Reset()

	req := createTestRequest(http.MethodPost, "/api/auth/login")
	ctx := context.Background()

	AuditSuspiciousActivity(
		ctx,
		req,
		SuspiciousMultipleFailedLogins,
		"possible brute force attack",
		10,
	)

	output := handler.String()
	if !strings.Contains(output, "suspicious_activity") {
		t.Error("expected suspicious_activity event type")
	}
	if !strings.Contains(output, "multiple_failed_logins") {
		t.Error("expected suspicious type in log")
	}
	if !strings.Contains(output, "possible brute force attack") {
		t.Error("expected reason in log")
	}
	if !strings.Contains(output, "10") {
		t.Error("expected failed attempts count in log")
	}
}

func TestFailedLoginTracker_RecordAndCount(t *testing.T) {
	ResetFailedLoginTracker()
	tracker := GetFailedLoginTracker()

	req := createTestRequest(http.MethodPost, "/api/auth/login")
	ctx := context.Background()
	key := "test-ip-1"

	// Initially should be 0.
	if count := tracker.GetAttemptCount(key); count != 0 {
		t.Errorf("expected 0 attempts, got %d", count)
	}

	// Record some attempts.
	for range 3 {
		tracker.RecordFailedAttempt(ctx, req, key, "testuser")
	}

	if count := tracker.GetAttemptCount(key); count != 3 {
		t.Errorf("expected 3 attempts, got %d", count)
	}
}

func TestFailedLoginTracker_ClearAttempts(t *testing.T) {
	ResetFailedLoginTracker()
	tracker := GetFailedLoginTracker()

	req := createTestRequest(http.MethodPost, "/api/auth/login")
	ctx := context.Background()
	key := "test-ip-2"

	// Record some attempts.
	for range 3 {
		tracker.RecordFailedAttempt(ctx, req, key, "testuser")
	}

	if count := tracker.GetAttemptCount(key); count != 3 {
		t.Errorf("expected 3 attempts, got %d", count)
	}

	// Clear attempts.
	tracker.ClearAttempts(key)

	if count := tracker.GetAttemptCount(key); count != 0 {
		t.Errorf("expected 0 attempts after clear, got %d", count)
	}
}

func TestFailedLoginTracker_CleanupOldAttempts(t *testing.T) {
	ResetFailedLoginTracker()
	tracker := GetFailedLoginTracker()

	// Manually insert an old attempt.
	tracker.mu.Lock()
	oldTime := time.Now().Add(-FailedLoginWindow - time.Minute)
	tracker.attempts["old-ip"] = []time.Time{oldTime}
	recentTime := time.Now()
	tracker.attempts["recent-ip"] = []time.Time{recentTime}
	tracker.mu.Unlock()

	// Run cleanup.
	tracker.CleanupOldAttempts()

	// Old IP should be removed.
	if count := tracker.GetAttemptCount("old-ip"); count != 0 {
		t.Errorf("expected old attempts to be cleaned up, got %d", count)
	}

	// Recent IP should remain.
	if count := tracker.GetAttemptCount("recent-ip"); count != 1 {
		t.Errorf("expected recent attempt to remain, got %d", count)
	}
}

func TestSecurityEventTypes(t *testing.T) {
	// Verify all event types are distinct.
	eventTypes := []SecurityEventType{
		EventLoginSuccess,
		EventLoginFailure,
		EventTokenInvalid,
		EventTokenExpired,
		EventTokenRevoked,
		EventTokenRefresh,
		EventLogout,
		EventPermissionDenied,
		EventRateLimited,
		EventSuspiciousActivity,
	}

	seen := make(map[SecurityEventType]bool)
	for _, et := range eventTypes {
		if seen[et] {
			t.Errorf("duplicate event type: %s", et)
		}
		seen[et] = true

		// Verify event type is not empty.
		if string(et) == "" {
			t.Error("event type should not be empty")
		}
	}
}

func TestSuspiciousActivityTypes(t *testing.T) {
	// Verify all suspicious activity types are distinct.
	types := []SuspiciousActivityType{
		SuspiciousMultipleFailedLogins,
		SuspiciousBruteForce,
		SuspiciousIPChange,
		SuspiciousUserAgentChange,
	}

	seen := make(map[SuspiciousActivityType]bool)
	for _, st := range types {
		if seen[st] {
			t.Errorf("duplicate suspicious activity type: %s", st)
		}
		seen[st] = true

		// Verify type is not empty.
		if string(st) == "" {
			t.Error("suspicious activity type should not be empty")
		}
	}
}

func TestLogSecurityEvent_LogLevels(t *testing.T) {
	tests := []struct {
		eventType     SecurityEventType
		expectedLevel string
	}{
		{EventLoginSuccess, "INFO"},
		{EventLogout, "INFO"},
		{EventTokenRefresh, "INFO"},
		{EventLoginFailure, "WARN"},
		{EventTokenInvalid, "WARN"},
		{EventTokenExpired, "WARN"},
		{EventTokenRevoked, "WARN"},
		{EventPermissionDenied, "WARN"},
		{EventRateLimited, "WARN"},
		{EventSuspiciousActivity, "WARN"},
	}

	for _, tt := range tests {
		t.Run(string(tt.eventType), func(t *testing.T) {
			handler := setupTestLogger()
			defer handler.Reset()

			event := &SecurityEvent{
				Timestamp:      time.Time{},
				EventType:      tt.eventType,
				UserID:         "",
				Username:       "",
				IPAddress:      "127.0.0.1",
				UserAgent:      "Test",
				RequestID:      "",
				Resource:       "",
				Reason:         "",
				SuspiciousType: "",
				FailedAttempts: 0,
				WindowDuration: "",
			}

			LogSecurityEvent(context.Background(), event)

			output := handler.String()
			if !strings.Contains(output, tt.expectedLevel) {
				t.Errorf("expected log level %s for event %s, got: %s",
					tt.expectedLevel, tt.eventType, output)
			}
		})
	}
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 192.168.1.1")
	req.RemoteAddr = "127.0.0.1:8080"

	ip := GetClientIP(req)
	if ip != "10.0.0.1" {
		t.Errorf("expected '10.0.0.1', got '%s'", ip)
	}
}

func TestGetClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "10.0.0.2")
	req.RemoteAddr = "127.0.0.1:8080"

	ip := GetClientIP(req)
	if ip != "10.0.0.2" {
		t.Errorf("expected '10.0.0.2', got '%s'", ip)
	}
}

func TestGetClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.50:12345"

	ip := GetClientIP(req)
	if ip != "192.168.1.50" {
		t.Errorf("expected '192.168.1.50', got '%s'", ip)
	}
}

func TestResetFailedLoginTracker(t *testing.T) {
	tracker := GetFailedLoginTracker()
	req := createTestRequest(http.MethodPost, "/api/auth/login")
	ctx := context.Background()

	// Add some attempts.
	tracker.RecordFailedAttempt(ctx, req, "test-ip", "user")

	// Reset.
	ResetFailedLoginTracker()

	// Should be empty now.
	if count := tracker.GetAttemptCount("test-ip"); count != 0 {
		t.Errorf("expected 0 after reset, got %d", count)
	}
}
