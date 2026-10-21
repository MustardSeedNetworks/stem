// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-webauthn/webauthn/protocol"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/api/ratelimit"
)

// setupAuthTestServer listens on 8444, so its relying party is localhost
// at that port.
const (
	passkeyRPID   = "localhost"
	passkeyOrigin = "https://localhost:8444"
)

// passkeyFixture is a signed-in server; one ceremony test drives more auth
// requests than the 5/min default allows from one address.
type passkeyFixture struct {
	server *api.Server
	token  string
}

func newPasskeyFixture(t *testing.T) *passkeyFixture {
	t.Helper()
	t.Setenv(ratelimit.AuthRateLimitEnv, "100")
	s := setupAuthTestServer(t)
	token, _ := getAuthToken(t, s)
	return &passkeyFixture{server: s, token: token}
}

// post sends body to path, signed in with the fixture's session when
// signedIn, carrying cookie as the browser would.
func (f *passkeyFixture) post(
	t *testing.T, path string, body []byte, signedIn bool, cookie *http.Cookie,
) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if signedIn {
		authorizeWithCSRF(t, f.server, req, f.token)
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	f.server.ServeHTTP(w, req)
	return w
}

func ceremonyCookie(t *testing.T, w *httptest.ResponseRecorder, name, path string) *http.Cookie {
	t.Helper()
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			if c.Path != path || !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteStrictMode {
				t.Errorf("%s cookie = %+v, want HttpOnly, Secure, SameSite=Strict on %s", name, c, path)
			}
			return c
		}
	}
	t.Fatalf("no %s cookie in %v", name, w.Result().Cookies())
	return nil
}

func requireStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatalf("status = %d, want %d: %s", w.Code, want, w.Body.String())
	}
}

func beginPasskeyRegistration(
	t *testing.T, f *passkeyFixture,
) (protocol.PublicKeyCredentialCreationOptions, *http.Cookie) {
	t.Helper()
	w := f.post(t, "/api/v1/auth/webauthn/register/begin", nil, true, nil)
	requireStatus(t, w, http.StatusOK)
	cookie := ceremonyCookie(t, w, "stem_passkey_register", "/api/v1/auth/webauthn/register/finish")
	var creation protocol.CredentialCreation
	if err := json.Unmarshal(w.Body.Bytes(), &creation); err != nil {
		t.Fatal(err)
	}
	return creation.Response, cookie
}

// enrolPasskey runs registration through the production routes and returns
// the authenticator holding the new credential.
func enrolPasskey(t *testing.T, f *passkeyFixture) *softAuthenticator {
	t.Helper()
	options, cookie := beginPasskeyRegistration(t, f)
	if options.User.Name != authTestUsername {
		t.Errorf("user name = %q, want %q", options.User.Name, authTestUsername)
	}
	selection := options.AuthenticatorSelection
	if selection.ResidentKey != protocol.ResidentKeyRequirementRequired ||
		selection.UserVerification != protocol.VerificationRequired {
		t.Errorf("authenticatorSelection = %+v, want a required discoverable, user-verified credential", selection)
	}
	handle, err := base64.RawURLEncoding.DecodeString(options.User.ID.(string))
	if err != nil {
		t.Fatal(err)
	}

	authenticator := newSoftAuthenticator(t, passkeyRPID, passkeyOrigin)
	body := authenticator.register(t, options.Challenge.String(), handle)
	requireStatus(t, f.post(t, "/api/v1/auth/webauthn/register/finish", body, true, cookie), http.StatusOK)

	// The challenge is single-use: replaying the same response fails.
	requireStatus(t, f.post(t, "/api/v1/auth/webauthn/register/finish", body, true, cookie), http.StatusBadRequest)
	return authenticator
}

func beginPasskeyLogin(t *testing.T, f *passkeyFixture) (string, *http.Cookie) {
	t.Helper()
	w := f.post(t, "/api/v1/auth/webauthn/login/begin", []byte(`{}`), false, nil)
	requireStatus(t, w, http.StatusOK)
	cookie := ceremonyCookie(t, w, "stem_passkey_login", "/api/v1/auth/webauthn/login/finish")
	var assertion protocol.CredentialAssertion
	if err := json.Unmarshal(w.Body.Bytes(), &assertion); err != nil {
		t.Fatal(err)
	}
	if assertion.Response.UserVerification != protocol.VerificationRequired {
		t.Errorf("userVerification = %q, want required", assertion.Response.UserVerification)
	}
	if len(assertion.Response.AllowedCredentials) != 0 {
		t.Error("sign-in is username-free and must name no credentials")
	}
	return assertion.Response.Challenge.String(), cookie
}

func TestPasskeyRegisterThenSignIn(t *testing.T) {
	f := newPasskeyFixture(t)
	authenticator := enrolPasskey(t, f)
	if got := f.server.AuthManagerForTest().WebAuthnCredentialCount(); got != 1 {
		t.Fatalf("WebAuthnCredentialCount() = %d, want 1", got)
	}

	challenge, cookie := beginPasskeyLogin(t, f)
	body := authenticator.assert(t, challenge, authenticator.handle)
	w := f.post(t, "/api/v1/auth/webauthn/login/finish", body, false, cookie)
	requireStatus(t, w, http.StatusOK)
	var login api.AuthLoginResponse
	if err := json.Unmarshal(w.Body.Bytes(), &login); err != nil {
		t.Fatal(err)
	}
	claims, err := f.server.AuthManagerForTest().ValidateToken(t.Context(), login.Token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims.Username != authTestUsername {
		t.Errorf("token username = %q, want %q", claims.Username, authTestUsername)
	}
	stored := f.server.AuthManagerForTest().PasskeyAccount().WebAuthnCredentials()
	if stored[0].Authenticator.SignCount != 1 {
		t.Errorf("stored sign count = %d, want the assertion's 1", stored[0].Authenticator.SignCount)
	}

	// Replaying the signed assertion against the spent ceremony fails.
	requireStatus(t, f.post(t, "/api/v1/auth/webauthn/login/finish", body, false, cookie), http.StatusBadRequest)
}

func TestPasskeySignInRejectsForeignCeremonyAndHandle(t *testing.T) {
	f := newPasskeyFixture(t)
	authenticator := enrolPasskey(t, f)

	t.Run("browser secret from another browser", func(t *testing.T) {
		challenge, cookie := beginPasskeyLogin(t, f)
		forged := *cookie
		forged.Value = cookie.Value[:len(cookie.Value)-1] + "x"
		body := authenticator.assert(t, challenge, authenticator.handle)
		requireStatus(t, f.post(t, "/api/v1/auth/webauthn/login/finish", body, false, &forged), http.StatusBadRequest)
	})

	t.Run("no ceremony cookie", func(t *testing.T) {
		challenge, _ := beginPasskeyLogin(t, f)
		body := authenticator.assert(t, challenge, authenticator.handle)
		requireStatus(t, f.post(t, "/api/v1/auth/webauthn/login/finish", body, false, nil), http.StatusBadRequest)
	})

	t.Run("user handle of another account", func(t *testing.T) {
		challenge, cookie := beginPasskeyLogin(t, f)
		body := authenticator.assert(t, challenge, bytes.Repeat([]byte{7}, len(authenticator.handle)))
		requireStatus(t, f.post(t, "/api/v1/auth/webauthn/login/finish", body, false, cookie), http.StatusUnauthorized)
	})

	t.Run("registration ceremony cannot finish a sign-in", func(t *testing.T) {
		options, cookie := beginPasskeyRegistration(t, f)
		login := *cookie
		login.Name = "stem_passkey_login"
		body := authenticator.assert(t, options.Challenge.String(), authenticator.handle)
		requireStatus(t, f.post(t, "/api/v1/auth/webauthn/login/finish", body, false, &login), http.StatusBadRequest)
	})
}

func TestPasskeySignInNeedsAnEnrolledPasskey(t *testing.T) {
	f := newPasskeyFixture(t)
	requireStatus(t, f.post(t, "/api/v1/auth/webauthn/login/begin", []byte(`{}`), false, nil), http.StatusBadRequest)
}

func TestPasskeyRoutesDisabledWithoutRelyingParty(t *testing.T) {
	// WebAuthn refuses an IP address as a relying party.
	t.Setenv(api.WebAuthnOriginEnv, "https://127.0.0.1:8444")
	f := newPasskeyFixture(t)
	w := f.post(t, "/api/v1/auth/webauthn/login/begin", []byte(`{}`), false, nil)
	requireStatus(t, w, http.StatusServiceUnavailable)
	requireStatus(t, f.post(t, "/api/v1/auth/webauthn/register/begin", nil, true, nil), http.StatusServiceUnavailable)
}
