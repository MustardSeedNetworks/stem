// SPDX-License-Identifier: BUSL-1.1

package logging

// MFA-specific audit helpers — Wave 3 (#85). Kept in a separate file
// to avoid further bloat of audit.go and to make the additive nature
// of Wave 3 obvious in diffs.

import (
	"context"
	"net/http"
	"time"
)

// MFAResult enumerates the outcome of an MFA verification.
type MFAResult string

// MFA result constants.
const (
	// MFAResultSuccess indicates the second factor verified successfully.
	MFAResultSuccess MFAResult = "success"
	// MFAResultFailure indicates the second factor was rejected.
	MFAResultFailure MFAResult = "failure"
)

// MFAFactor identifies which second-factor mechanism produced an event.
type MFAFactor string

// MFA factor constants.
const (
	// MFAFactorTOTP identifies time-based one-time-password events.
	MFAFactorTOTP MFAFactor = "totp"
	// MFAFactorWebAuthn identifies WebAuthn / passkey events.
	MFAFactorWebAuthn MFAFactor = "webauthn"
)

// AuditMFASetup records the start of a TOTP enrolment: a fresh secret
// was generated and returned to the UI, but the user has not yet
// proven they configured their authenticator. Pair with AuditMFAAttempt
// (result=success) when enrolment completes.
func AuditMFASetup(ctx context.Context, r *http.Request, username string, factor MFAFactor) {
	LogSecurityEvent(ctx, &SecurityEvent{
		Timestamp:         time.Time{},
		EventType:         EventMFASetup,
		UserID:            username,
		Username:          username,
		IPAddress:         GetClientIP(r),
		UserAgent:         r.UserAgent(),
		RequestID:         "",
		Resource:          string(factor),
		Reason:            "",
		SuspiciousType:    "",
		FailedAttempts:    0,
		WindowDuration:    "",
		Result:            "",
		RejectReason:      "",
		PreviousAlgorithm: "",
	})
}

// AuditMFAAttempt records a second-factor verification result. Called
// from both enrolment-verify and login-verify paths; the `factor`
// argument differentiates TOTP from WebAuthn so dashboards can break
// out the counts.
func AuditMFAAttempt(
	ctx context.Context,
	r *http.Request,
	username string,
	factor MFAFactor,
	result MFAResult,
	reason string,
) {
	LogSecurityEvent(ctx, &SecurityEvent{
		Timestamp:         time.Time{},
		EventType:         EventMFAVerify,
		UserID:            username,
		Username:          username,
		IPAddress:         GetClientIP(r),
		UserAgent:         r.UserAgent(),
		RequestID:         "",
		Resource:          string(factor),
		Reason:            reason,
		SuspiciousType:    "",
		FailedAttempts:    0,
		WindowDuration:    "",
		Result:            string(result),
		RejectReason:      "",
		PreviousAlgorithm: "",
	})
}

// AuditMFADisable records that the second factor was turned off.
// Step-up authentication (password + current code) happens in the
// HTTP handler before this is called; the audit entry records the
// final mutation.
func AuditMFADisable(ctx context.Context, r *http.Request, username string, factor MFAFactor) {
	LogSecurityEvent(ctx, &SecurityEvent{
		Timestamp:         time.Time{},
		EventType:         EventMFADisable,
		UserID:            username,
		Username:          username,
		IPAddress:         GetClientIP(r),
		UserAgent:         r.UserAgent(),
		RequestID:         "",
		Resource:          string(factor),
		Reason:            "",
		SuspiciousType:    "",
		FailedAttempts:    0,
		WindowDuration:    "",
		Result:            "",
		RejectReason:      "",
		PreviousAlgorithm: "",
	})
}

// AuditWebAuthnRegister records the result of a WebAuthn passkey
// registration ceremony. `result` distinguishes a successful enrolment
// from a verification failure; `reason` is the operator-friendly
// rejection message on failure.
func AuditWebAuthnRegister(
	ctx context.Context,
	r *http.Request,
	username string,
	result MFAResult,
	reason string,
) {
	LogSecurityEvent(ctx, &SecurityEvent{
		Timestamp:         time.Time{},
		EventType:         EventWebAuthnRegister,
		UserID:            username,
		Username:          username,
		IPAddress:         GetClientIP(r),
		UserAgent:         r.UserAgent(),
		RequestID:         "",
		Resource:          string(MFAFactorWebAuthn),
		Reason:            reason,
		SuspiciousType:    "",
		FailedAttempts:    0,
		WindowDuration:    "",
		Result:            string(result),
		RejectReason:      "",
		PreviousAlgorithm: "",
	})
}

// AuditWebAuthnLogin records the result of a WebAuthn login ceremony.
// Mirrors AuditWebAuthnRegister; kept distinct so dashboards can
// separate the two event streams.
func AuditWebAuthnLogin(
	ctx context.Context,
	r *http.Request,
	username string,
	result MFAResult,
	reason string,
) {
	LogSecurityEvent(ctx, &SecurityEvent{
		Timestamp:         time.Time{},
		EventType:         EventWebAuthnLogin,
		UserID:            username,
		Username:          username,
		IPAddress:         GetClientIP(r),
		UserAgent:         r.UserAgent(),
		RequestID:         "",
		Resource:          string(MFAFactorWebAuthn),
		Reason:            reason,
		SuspiciousType:    "",
		FailedAttempts:    0,
		WindowDuration:    "",
		Result:            string(result),
		RejectReason:      "",
		PreviousAlgorithm: "",
	})
}
