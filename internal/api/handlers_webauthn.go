// SPDX-License-Identifier: BUSL-1.1

package api

// Passkey endpoints: enrolment for the signed-in operator and passwordless
// sign-in. foundation/pkg/passkey owns the ceremonies and the fleet's
// verification policy (discoverable credentials, user verification,
// single-use browser-bound challenges); this file owns stem's routes, the
// cookie that binds a ceremony to its browser, and the relying-party origin.

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/MustardSeedNetworks/foundation/pkg/passkey"

	"github.com/MustardSeedNetworks/stem/internal/auth"
	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// WebAuthnOriginEnv names the canonical HTTPS origin operators browse to,
// such as https://stem.example.com:8444. Its host is the relying-party ID.
// Unset, the origin is https://localhost on the port actually bound.
const WebAuthnOriginEnv = "STEM_WEBAUTHN_ORIGIN"

// passkeyCeremonyTTL matches foundation's ceremony timeout: the cookie lives
// no longer than the challenge it names.
const passkeyCeremonyTTL = 5 * time.Minute

// initPasskeys builds the relying party for the port the listener serves on,
// with a fresh ceremony store: a session is valid only with the service that
// issued it. A configured origin that foundation refuses disables passkeys
// rather than falling back to another origin.
func (s *Server) initPasskeys(port int) {
	origin := os.Getenv(WebAuthnOriginEnv)
	if origin == "" {
		origin = "https://localhost:" + strconv.Itoa(port)
	}
	var rpID string
	if parsed, err := url.Parse(origin); err == nil {
		rpID = parsed.Hostname()
	}
	service, err := passkey.New(passkey.Config{RPID: rpID, Origin: origin, DisplayName: "Stem"})
	if err != nil {
		logging.Warn("Passkeys disabled: the WebAuthn origin is not a valid relying party",
			"origin", origin, "error", err.Error())
	}
	s.passkeys = service
	s.passkeyCeremonies = passkey.NewStore()
}

// passkeyCookieScope returns a ceremony cookie's name and path. Each purpose
// has its own cookie scoped to its own finish route, so it travels with no
// other request.
func passkeyCookieScope(purpose passkey.Purpose) (string, string) {
	if purpose == passkey.Registration {
		return "stem_passkey_register", "/api/v1/auth/webauthn/register/finish"
	}
	return "stem_passkey_login", "/api/v1/auth/webauthn/login/finish"
}

func setPasskeyCookie(w http.ResponseWriter, purpose passkey.Purpose, value string, maxAge int) {
	name, path := passkeyCookieScope(purpose)
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// startPasskeyCeremony keeps the session server-side under a random browser
// secret and hands the browser only "<ceremony id>.<secret>" in an HttpOnly
// cookie. account is the user handle for enrolment and empty for sign-in.
func (s *Server) startPasskeyCeremony(
	w http.ResponseWriter, purpose passkey.Purpose, session *passkey.Session, account string,
) error {
	const secretBytes = 32
	secret := make([]byte, secretBytes)
	_, _ = rand.Read(secret) // crypto/rand.Read never returns an error
	browser := base64.RawURLEncoding.EncodeToString(secret)
	id, err := s.passkeyCeremonies.Add(*session, passkey.Binding{
		Browser: browser, AccountID: account, Purpose: purpose,
	})
	if err != nil {
		return err
	}
	setPasskeyCookie(w, purpose, id+"."+browser, int(passkeyCeremonyTTL.Seconds()))
	return nil
}

// takePasskeyCeremony consumes the ceremony the request's cookie names. The
// cookie is spent either way: a failed finish needs a fresh challenge.
func (s *Server) takePasskeyCeremony(
	w http.ResponseWriter, r *http.Request, purpose passkey.Purpose, account string,
) (passkey.Session, bool) {
	setPasskeyCookie(w, purpose, "", -1)
	name, _ := passkeyCookieScope(purpose)
	cookie, err := r.Cookie(name)
	if err != nil {
		return passkey.Session{}, false
	}
	id, browser, ok := strings.Cut(cookie.Value, ".")
	if !ok {
		return passkey.Session{}, false
	}
	return s.passkeyCeremonies.Take(id, passkey.Binding{
		Browser: browser, AccountID: account, Purpose: purpose,
	})
}

// handleWebAuthnRegisterBegin issues a creation challenge.
func (s *Server) handleWebAuthnRegisterBegin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}
	if s.passkeys == nil {
		s.writeMFAError(w, auth.ErrWebAuthnNotConfigured)
		return
	}
	account := s.authManager.PasskeyAccount()
	options, session, err := s.passkeys.BeginRegistration(account)
	if err == nil {
		err = s.startPasskeyCeremony(w, passkey.Registration, session, string(account.WebAuthnID()))
	}
	if err != nil {
		logging.AuditWebAuthnRegister(r.Context(), r, account.WebAuthnName(),
			logging.MFAResultFailure, err.Error())
		s.writeMFAError(w, err)
		return
	}
	writeJSON(w, options)
}

// handleWebAuthnRegisterFinish validates the attestation and stores
// the credential on success.
func (s *Server) handleWebAuthnRegisterFinish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}
	if s.passkeys == nil {
		s.writeMFAError(w, auth.ErrWebAuthnNotConfigured)
		return
	}
	account := s.authManager.PasskeyAccount()
	username := account.WebAuthnName()
	session, ok := s.takePasskeyCeremony(w, r, passkey.Registration, string(account.WebAuthnID()))
	if !ok {
		logging.AuditWebAuthnRegister(r.Context(), r, username,
			logging.MFAResultFailure, "no matching ceremony")
		WriteInvalidRequest(w, "WebAuthn ceremony invalid")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	credential, err := s.passkeys.FinishRegistration(account, session, r)
	if err != nil {
		logging.AuditWebAuthnRegister(r.Context(), r, username,
			logging.MFAResultFailure, err.Error())
		WriteInvalidRequest(w, "Invalid WebAuthn registration response")
		return
	}
	s.authManager.SavePasskey(*credential)
	logging.AuditWebAuthnRegister(r.Context(), r, username,
		logging.MFAResultSuccess, "registered")
	writeJSON(w, map[string]any{
		jsonFieldSuccess: true,
		"credentialId":   base64.RawURLEncoding.EncodeToString(credential.ID),
	})
}

// handleWebAuthnLoginBegin issues a username-free assertion challenge.
func (s *Server) handleWebAuthnLoginBegin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}
	if s.passkeys == nil {
		s.writeMFAError(w, auth.ErrWebAuthnNotConfigured)
		return
	}
	username := s.authManager.GetUsername()
	if !s.authManager.HasWebAuthnCredentials() {
		logging.AuditWebAuthnLogin(r.Context(), r, username,
			logging.MFAResultFailure, auth.ErrWebAuthnNoCredentials.Error())
		s.writeMFAError(w, auth.ErrWebAuthnNoCredentials)
		return
	}
	options, session, err := s.passkeys.BeginLogin()
	if err == nil {
		err = s.startPasskeyCeremony(w, passkey.Login, session, "")
	}
	if err != nil {
		logging.AuditWebAuthnLogin(r.Context(), r, username,
			logging.MFAResultFailure, err.Error())
		s.writeMFAError(w, err)
		return
	}
	writeJSON(w, options)
}

// handleWebAuthnLoginFinish validates an assertion. On success, the
// browser has proven possession of a registered passkey; the user
// receives a full access + refresh token pair just like the password
// path.
func (s *Server) handleWebAuthnLoginFinish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}
	if s.passkeys == nil {
		s.writeMFAError(w, auth.ErrWebAuthnNotConfigured)
		return
	}
	username := s.authManager.GetUsername()
	session, ok := s.takePasskeyCeremony(w, r, passkey.Login, "")
	if !ok {
		logging.AuditWebAuthnLogin(r.Context(), r, username,
			logging.MFAResultFailure, "no matching ceremony")
		WriteInvalidRequest(w, "WebAuthn ceremony invalid")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	_, credential, err := s.passkeys.FinishLogin(s.authManager.ResolvePasskey, session, r)
	if err != nil {
		logging.AuditWebAuthnLogin(r.Context(), r, username,
			logging.MFAResultFailure, err.Error())
		WriteError(w, ErrAuthFailed)
		return
	}
	s.authManager.SavePasskey(*credential)

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
	if newSessionID := sessionIDFromJWT(accessToken); newSessionID != "" {
		s.csrfManager.Revoke(newSessionID)
	}

	logging.AuditWebAuthnLogin(r.Context(), r, username, logging.MFAResultSuccess, "login completed")
	s.auditor.LoginSuccess(r.Context(), r, username, username)

	writeJSON(w, AuthLoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(sessionDuration).Unix(),
	})
}
