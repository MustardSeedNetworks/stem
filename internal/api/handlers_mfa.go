// SPDX-License-Identifier: BUSL-1.1

package api

// Multi-factor authentication endpoints. Wave 3 (#85) adds TOTP and
// WebAuthn (passkeys) as second factors on top of the existing
// password login. The handlers integrate with the CSRF middleware via
// setupRoutes — every state-changing POST is registered behind
// s.handleAuthRateLimited (which is itself behind the CSRF middleware
// installed in Run()).
//
// CSRF: setup/verify/disable/login POSTs require a CSRF token in
// X-Csrf-Token. The login-finisher (POST /api/v1/auth/login/totp) is
// exempt because it carries the mfa_token, which acts as the
// pre-session credential proof of intent — the same role
// /api/v1/auth/login already plays. Its route therefore declares no
// CSRF (routes.go) for this reason.

import (
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/auth"
	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// MFA constants and request / response shapes. Kept private to api/
// so the public API surface remains documented via types.go.

// jsonFieldSuccess is the JSON key for the boolean success flag
// returned by every MFA mutation endpoint.
const jsonFieldSuccess = "success"

// totpSetupResponse is returned by POST /api/v1/auth/totp/setup. The
// QR PNG is base64-encoded for transport in JSON.
type totpSetupResponse struct {
	Secret          string `json:"secret"`
	ProvisioningURI string `json:"provisioningUri"`
	QRCodePNGBase64 string `json:"qrCodePngBase64"`
}

// totpVerifyRequest enables MFA after the user has scanned the QR and
// can produce a current code.
type totpVerifyRequest struct {
	Code string `json:"code"`
}

// totpDisableRequest re-authenticates with password + current code
// before tearing down MFA.
type totpDisableRequest struct {
	Password string `json:"password"`
	Code     string `json:"code"`
}

// loginMFARequest finishes a two-step login: caller presents the
// mfa_token they got from /api/v1/auth/login and the current TOTP
// code.
type loginMFARequest struct {
	MFAToken string `json:"mfaToken"`
	Code     string `json:"code"`
}

// mfaRequiredResponse is the body returned by /api/v1/auth/login when
// MFA is enrolled. The caller must POST to /api/v1/auth/login/totp
// (or /api/v1/auth/login/webauthn) with the embedded mfa_token to
// complete the flow.
type mfaRequiredResponse struct {
	MFARequired bool   `json:"mfaRequired"`
	MFAToken    string `json:"mfaToken"`
	Factor      string `json:"factor"`
}

// mfaStatusResponse is returned by GET /api/v1/auth/mfa/status. The
// UI uses this on the account/security page to decide whether to show
// the enroll or disable button.
type mfaStatusResponse struct {
	TOTPEnabled             bool `json:"totpEnabled"`
	WebAuthnRegistered      bool `json:"webauthnRegistered"`
	WebAuthnCredentialCount int  `json:"webauthnCredentialCount"`
}

// handleTOTPSetup mints a fresh TOTP secret and returns the
// provisioning artefacts (URI + QR PNG). The secret is staged on the
// auth manager but MFA is NOT enabled until handleTOTPVerify confirms
// the user can produce a valid code.
func (s *Server) handleTOTPSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}

	username := s.authManager.GetUsername()
	if username == "" {
		WriteError(w, ErrInternalError)
		return
	}

	setup, err := auth.GenerateTOTPSecret(username, auth.DefaultTOTPIssuer)
	if err != nil {
		logging.Error("Failed to generate TOTP secret", "error", err)
		WriteError(w, ErrInternalError)
		return
	}

	s.authManager.SetTOTPSecret(setup.Secret)
	logging.AuditMFASetup(r.Context(), r, username, logging.MFAFactorTOTP)

	writeJSON(w, totpSetupResponse{
		Secret:          setup.Secret,
		ProvisioningURI: setup.ProvisioningURI,
		QRCodePNGBase64: base64.StdEncoding.EncodeToString(setup.QRCodePNG),
	})
}

// handleTOTPVerify finishes TOTP enrolment. Verifies the supplied
// code against the staged secret; on success, MFA is enabled.
func (s *Server) handleTOTPVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}

	var req totpVerifyRequest
	if !decodeJSONStrict(w, r, &req, maxRequestBodySize) {
		return
	}

	username := s.authManager.GetUsername()
	err := s.authManager.VerifyAndEnableTOTP(req.Code)
	if err != nil {
		logging.AuditMFAAttempt(r.Context(), r, username, logging.MFAFactorTOTP,
			logging.MFAResultFailure, err.Error())
		s.writeMFAError(w, err)
		return
	}
	logging.AuditMFAAttempt(r.Context(), r, username, logging.MFAFactorTOTP,
		logging.MFAResultSuccess, "enrolled")
	writeJSON(w, map[string]any{
		jsonFieldSuccess: true,
		"totpEnabled":    true,
	})
}

// handleTOTPDisable tears down MFA. Step-up: the caller must present
// both the current password AND a valid TOTP code. The password
// verification reuses the existing Authenticate flow.
func (s *Server) handleTOTPDisable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}

	var req totpDisableRequest
	if !decodeJSONStrict(w, r, &req, maxRequestBodySize) {
		return
	}

	username := s.authManager.GetUsername()

	// Step-up #1: password.
	if _, passwordErr := s.authManager.Authenticate(username, req.Password); passwordErr != nil {
		logging.AuditMFAAttempt(r.Context(), r, username, logging.MFAFactorTOTP,
			logging.MFAResultFailure, "password rejected")
		s.writeAuthError(w, passwordErr)
		return
	}

	// Step-up #2: current TOTP code.
	if codeErr := s.authManager.VerifyTOTPCode(req.Code); codeErr != nil {
		logging.AuditMFAAttempt(r.Context(), r, username, logging.MFAFactorTOTP,
			logging.MFAResultFailure, "totp rejected")
		s.writeMFAError(w, codeErr)
		return
	}

	s.authManager.DisableTOTP()
	logging.AuditMFADisable(r.Context(), r, username, logging.MFAFactorTOTP)

	writeJSON(w, map[string]any{
		jsonFieldSuccess: true,
		"totpEnabled":    false,
	})
}

// handleLoginTOTP finishes the two-step login. It validates the
// short-lived mfa_token issued by the password-stage login, then
// verifies the TOTP code, and on success issues full access + refresh
// tokens.
//
// This route declares no CSRF (routes.go): the
// mfa_token is a server-signed proof of the password stage, which is
// equivalent to /api/v1/auth/login's "credential is the intent". The
// rate limiter applied at registration time still caps brute-force.
func (s *Server) handleLoginTOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}

	var req loginMFARequest
	if !decodeJSONStrict(w, r, &req, maxRequestBodySize) {
		return
	}

	username, err := s.authManager.ValidateMFAPendingToken(r.Context(), req.MFAToken)
	if err != nil {
		logging.AuditMFAAttempt(r.Context(), r, "", logging.MFAFactorTOTP,
			logging.MFAResultFailure, "invalid mfa_token")
		s.writeMFAError(w, err)
		return
	}

	if codeErr := s.authManager.VerifyTOTPCode(req.Code); codeErr != nil {
		logging.AuditMFAAttempt(r.Context(), r, username, logging.MFAFactorTOTP,
			logging.MFAResultFailure, "wrong code")
		s.writeMFAError(w, codeErr)
		return
	}

	accessToken, err := s.authManager.IssueAccessToken(username)
	if err != nil {
		WriteError(w, ErrInternalError)
		return
	}
	refreshToken, err := s.authManager.GenerateRefreshToken(username)
	if err != nil {
		WriteError(w, ErrInternalError)
		return
	}
	sessionDuration := s.authManager.SessionDuration()
	auth.SetAccessTokenCookie(w, accessToken, sessionDuration, s.cookieConfig)
	auth.SetRefreshTokenCookie(w, refreshToken, sessionDuration*refreshMultiplier, s.cookieConfig)

	// Same CSRF rotation as the regular login.
	if newSessionID := sessionIDFromJWT(accessToken); newSessionID != "" {
		s.csrfManager.Revoke(newSessionID)
	}

	logging.AuditMFAAttempt(r.Context(), r, username, logging.MFAFactorTOTP,
		logging.MFAResultSuccess, "login completed")
	s.auditor.LoginSuccess(r.Context(), r, username, username)

	writeJSON(w, AuthLoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(sessionDuration).Unix(),
	})
}

// handleMFAStatus returns the current MFA enrolment state. The UI
// uses this on the account/security page to decide which buttons to
// show.
func (s *Server) handleMFAStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
		return
	}
	writeJSON(w, mfaStatusResponse{
		TOTPEnabled:             s.authManager.TOTPEnabled(),
		WebAuthnRegistered:      s.authManager.HasWebAuthnCredentials(),
		WebAuthnCredentialCount: s.authManager.WebAuthnCredentialCount(),
	})
}

// writeMFAError translates auth-package MFA errors into appropriate
// HTTP responses.
func (s *Server) writeMFAError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrTOTPInvalidCode),
		errors.Is(err, auth.ErrTOTPEmptyCode),
		errors.Is(err, auth.ErrTOTPEmptySecret),
		errors.Is(err, auth.ErrTOTPNotEnabled):
		WriteInvalidRequest(w, "Invalid or missing MFA code")
	case errors.Is(err, auth.ErrTOTPAlreadyEnabled):
		WriteError(w, ConflictError("TOTP is already enabled"))
	case errors.Is(err, auth.ErrMFATokenInvalid),
		errors.Is(err, auth.ErrMFATokenExpired):
		WriteAuthError(w, err)
	case errors.Is(err, auth.ErrWebAuthnNoCredentials):
		WriteInvalidRequest(w, "WebAuthn ceremony invalid")
	case errors.Is(err, auth.ErrWebAuthnNotConfigured):
		WriteError(w, &Error{
			HTTPStatus: http.StatusServiceUnavailable,
			Code:       ErrCodeServiceUnavailable,
			Message:    "WebAuthn not configured",
		})
	default:
		logging.Error("MFA error", "error", err)
		WriteError(w, ErrInternalError)
	}
}

// loginWithMFAGate is the MFA-aware shim called from setupRoutes for
// the regular /api/v1/auth/login endpoint. It runs the existing
// password verification and, if MFA is enrolled for the account, intercepts the
// happy path to return a 200 with `mfa_required=true` instead of
// minting the full token pair.
func (s *Server) loginWithMFAGate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}

	var req AuthLoginRequest
	if !decodeJSONStrict(w, r, &req, maxRequestBodySize) {
		return
	}
	// Empty credentials are a malformed request, not a failed login. Without
	// this they reached AuthenticateWithRefresh, failed, and were recorded as
	// an audit event and a failed-login attempt — so an unauthenticated
	// caller could fill the suspicious-activity tracker with requests that
	// were never credentials.
	if !validateStruct(w, &req) {
		return
	}

	// Step 1: verify the password.
	accessToken, refreshToken, err := s.authManager.AuthenticateWithRefresh(req.Username, req.Password)
	if err != nil {
		s.auditor.LoginFailure(r.Context(), r, req.Username, err.Error())
		s.writeAuthError(w, err)
		return
	}

	// Step 2: if MFA is enabled, hold the full tokens back and issue
	// a short-lived mfa_pending JWT instead. The caller must POST
	// to /api/v1/auth/login/totp (or /webauthn) to finish.
	if s.authManager.TOTPEnabled() {
		mfaToken, mfaErr := s.authManager.GenerateMFAPendingToken(req.Username)
		if mfaErr != nil {
			WriteError(w, ErrInternalError)
			return
		}
		logging.AuditMFAAttempt(r.Context(), r, req.Username, logging.MFAFactorTOTP,
			logging.MFAResultSuccess, "password verified, mfa pending")
		writeJSON(w, mfaRequiredResponse{
			MFARequired: true,
			MFAToken:    mfaToken,
			Factor:      string(logging.MFAFactorTOTP),
		})
		return
	}

	// No MFA enrolled — proceed as before (full token pair).
	sessionDuration := s.authManager.SessionDuration()
	auth.SetAccessTokenCookie(w, accessToken, sessionDuration, s.cookieConfig)
	auth.SetRefreshTokenCookie(w, refreshToken, sessionDuration*refreshMultiplier, s.cookieConfig)
	if newSessionID := sessionIDFromJWT(accessToken); newSessionID != "" {
		s.csrfManager.Revoke(newSessionID)
	}
	s.auditor.LoginSuccess(r.Context(), r, req.Username, req.Username)
	writeJSON(w, AuthLoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(sessionDuration).Unix(),
	})
}
