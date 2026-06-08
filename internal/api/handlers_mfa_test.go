// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"

	"github.com/MustardSeedNetworks/stem/internal/api"
)

// TestTOTPSetup_RequiresAuth asserts that POST /api/v1/auth/totp/setup
// rejects unauthenticated requests with 401.
func TestTOTPSetup_RequiresAuth(t *testing.T) {
	s := setupAuthTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/totp/setup", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d", w.Code)
	}
}

// TestTOTPSetup_CSRFRequired confirms the Wave-1 CSRF middleware
// enforces a token on POST /api/v1/auth/totp/setup. A naked
// Authorization header (no X-Csrf-Token) must 403.
func TestTOTPSetup_CSRFRequired(t *testing.T) {
	s := setupAuthTestServer(t)
	token, _ := getAuthToken(t, s)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/totp/setup", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	// Deliberately no X-Csrf-Token.
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 without CSRF token, got %d: %s", w.Code, w.Body.String())
	}
}

// TestTOTPSetup_Success exercises the happy path: setup returns a
// secret + provisioning URI + base64 QR PNG.
func TestTOTPSetup_Success(t *testing.T) {
	s := setupAuthTestServer(t)
	token, _ := getAuthToken(t, s)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/totp/setup", nil)
	authorizeWithCSRF(t, s, req, token)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["secret"] == "" {
		t.Error("expected non-empty secret")
	}
	if resp["provisioningUri"] == "" {
		t.Error("expected non-empty provisioningUri")
	}
	if resp["qrCodePngBase64"] == "" {
		t.Error("expected non-empty qrCodePngBase64")
	}
}

// TestTOTPVerify_EnablesMFA exercises the setup → verify happy path.
func TestTOTPVerify_EnablesMFA(t *testing.T) {
	s := setupAuthTestServer(t)
	token, _ := getAuthToken(t, s)

	setupReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/totp/setup", nil)
	authorizeWithCSRF(t, s, setupReq, token)
	setupW := httptest.NewRecorder()
	s.ServeHTTP(setupW, setupReq)
	if setupW.Code != http.StatusOK {
		t.Fatalf("setup failed: %d", setupW.Code)
	}
	var setupResp map[string]string
	if err := json.Unmarshal(setupW.Body.Bytes(), &setupResp); err != nil {
		t.Fatalf("decode setup: %v", err)
	}
	code, err := totp.GenerateCode(setupResp["secret"], time.Now().UTC())
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}

	body := bytes.NewBufferString(`{"code":"` + code + `"}`)
	verifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/totp/verify", body)
	authorizeWithCSRF(t, s, verifyReq, token)
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyW := httptest.NewRecorder()
	s.ServeHTTP(verifyW, verifyReq)

	if verifyW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", verifyW.Code, verifyW.Body.String())
	}
}

// TestTOTPVerify_WrongCode rejects an incorrect TOTP code with 400.
func TestTOTPVerify_WrongCode(t *testing.T) {
	s := setupAuthTestServer(t)
	token, _ := getAuthToken(t, s)

	setupReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/totp/setup", nil)
	authorizeWithCSRF(t, s, setupReq, token)
	setupW := httptest.NewRecorder()
	s.ServeHTTP(setupW, setupReq)
	if setupW.Code != http.StatusOK {
		t.Fatalf("setup failed: %d", setupW.Code)
	}

	body := bytes.NewBufferString(`{"code":"000000"}`)
	verifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/totp/verify", body)
	authorizeWithCSRF(t, s, verifyReq, token)
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyW := httptest.NewRecorder()
	s.ServeHTTP(verifyW, verifyReq)

	if verifyW.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for wrong code, got %d: %s", verifyW.Code, verifyW.Body.String())
	}
}

// TestLogin_MFARequiredAfterEnrollment asserts the regular login
// returns mfa_required=true once TOTP is enabled, and that the
// follow-up /api/v1/auth/login/totp endpoint completes the flow.
func TestLogin_MFARequiredAfterEnrollment(t *testing.T) {
	s := setupAuthTestServer(t)
	token, _ := getAuthToken(t, s)

	// Enroll TOTP.
	setupReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/totp/setup", nil)
	authorizeWithCSRF(t, s, setupReq, token)
	setupW := httptest.NewRecorder()
	s.ServeHTTP(setupW, setupReq)
	var setupResp map[string]string
	_ = json.Unmarshal(setupW.Body.Bytes(), &setupResp)

	code, err := totp.GenerateCode(setupResp["secret"], time.Now().UTC())
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}
	verifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/totp/verify",
		bytes.NewBufferString(`{"code":"`+code+`"}`))
	authorizeWithCSRF(t, s, verifyReq, token)
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyW := httptest.NewRecorder()
	s.ServeHTTP(verifyW, verifyReq)
	if verifyW.Code != http.StatusOK {
		t.Fatalf("enroll failed: %d", verifyW.Code)
	}

	// Now log in again — should return mfa_required.
	loginBody := bytes.NewBufferString(
		`{"username":"` + authTestUsername + `","password":"` + authTestPassword + `"}`,
	)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", loginBody)
	loginW := httptest.NewRecorder()
	s.ServeHTTP(loginW, loginReq)

	if loginW.Code != http.StatusOK {
		t.Fatalf("login expected 200, got %d", loginW.Code)
	}
	var loginResp map[string]any
	if loginDecodeErr := json.Unmarshal(loginW.Body.Bytes(), &loginResp); loginDecodeErr != nil {
		t.Fatalf("decode login: %v", loginDecodeErr)
	}
	mfaRequired, _ := loginResp["mfaRequired"].(bool)
	if !mfaRequired {
		t.Errorf("expected mfaRequired=true, got %v", loginResp)
	}
	mfaToken, _ := loginResp["mfaToken"].(string)
	if mfaToken == "" {
		t.Fatal("expected non-empty mfaToken")
	}

	// Finish the MFA login.
	freshCode, err := totp.GenerateCode(setupResp["secret"], time.Now().UTC())
	if err != nil {
		t.Fatalf("generate finish code: %v", err)
	}
	finishBody := bytes.NewBufferString(
		`{"mfaToken":"` + mfaToken + `","code":"` + freshCode + `"}`,
	)
	finishReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/totp", finishBody)
	finishReq.Header.Set("Content-Type", "application/json")
	finishW := httptest.NewRecorder()
	s.ServeHTTP(finishW, finishReq)

	if finishW.Code != http.StatusOK {
		t.Fatalf("finish expected 200, got %d: %s", finishW.Code, finishW.Body.String())
	}
	var finishResp map[string]any
	if finishDecodeErr := json.Unmarshal(
		finishW.Body.Bytes(), &finishResp,
	); finishDecodeErr != nil {
		t.Fatalf("decode finish: %v", finishDecodeErr)
	}
	if _, ok := finishResp["token"].(string); !ok {
		t.Error("expected token in finish response")
	}
}

// TestLoginTOTP_WrongCode rejects with 400.
func TestLoginTOTP_WrongCode(t *testing.T) {
	s := setupAuthTestServer(t)
	token, _ := getAuthToken(t, s)

	// Enroll TOTP.
	setupReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/totp/setup", nil)
	authorizeWithCSRF(t, s, setupReq, token)
	setupW := httptest.NewRecorder()
	s.ServeHTTP(setupW, setupReq)
	var setupResp map[string]string
	_ = json.Unmarshal(setupW.Body.Bytes(), &setupResp)
	code, _ := totp.GenerateCode(setupResp["secret"], time.Now().UTC())
	verifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/totp/verify",
		bytes.NewBufferString(`{"code":"`+code+`"}`))
	authorizeWithCSRF(t, s, verifyReq, token)
	verifyReq.Header.Set("Content-Type", "application/json")
	s.ServeHTTP(httptest.NewRecorder(), verifyReq)

	// Re-login to get mfaToken.
	loginBody := bytes.NewBufferString(
		`{"username":"` + authTestUsername + `","password":"` + authTestPassword + `"}`,
	)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", loginBody)
	loginW := httptest.NewRecorder()
	s.ServeHTTP(loginW, loginReq)
	var loginResp map[string]any
	_ = json.Unmarshal(loginW.Body.Bytes(), &loginResp)
	mfaToken, _ := loginResp["mfaToken"].(string)

	// Wrong code.
	finishBody := bytes.NewBufferString(
		`{"mfaToken":"` + mfaToken + `","code":"000000"}`,
	)
	finishReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/totp", finishBody)
	finishReq.Header.Set("Content-Type", "application/json")
	finishW := httptest.NewRecorder()
	s.ServeHTTP(finishW, finishReq)
	if finishW.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for wrong code, got %d", finishW.Code)
	}
}

// TestLoginTOTP_InvalidMFAToken rejects garbage mfa_token with 401.
func TestLoginTOTP_InvalidMFAToken(t *testing.T) {
	s := setupAuthTestServer(t)

	finishBody := bytes.NewBufferString(`{"mfaToken":"garbage","code":"123456"}`)
	finishReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/totp", finishBody)
	finishReq.Header.Set("Content-Type", "application/json")
	finishW := httptest.NewRecorder()
	s.ServeHTTP(finishW, finishReq)
	if finishW.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for garbage mfa_token, got %d", finishW.Code)
	}
}

// TestMFAStatus_ReportsDisabledByDefault verifies the status endpoint
// returns totpEnabled=false on a fresh server.
func TestMFAStatus_ReportsDisabledByDefault(t *testing.T) {
	s := setupAuthTestServer(t)
	token, _ := getAuthToken(t, s)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/mfa/status", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got, _ := resp["totpEnabled"].(bool); got {
		t.Error("expected totpEnabled=false on fresh server")
	}
	if got, _ := resp["webauthnRegistered"].(bool); got {
		t.Error("expected webauthnRegistered=false on fresh server")
	}
}

// TestWebAuthn_RegisterBegin_CSRFRequired confirms the WebAuthn
// register-begin endpoint is behind the Wave-1 CSRF middleware.
func TestWebAuthn_RegisterBegin_CSRFRequired(t *testing.T) {
	s := setupAuthTestServer(t)
	token, _ := getAuthToken(t, s)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/webauthn/register/begin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 without CSRF, got %d", w.Code)
	}
}

// TestWebAuthn_RegisterBegin_Success returns a creation options blob
// with a non-empty challenge.
func TestWebAuthn_RegisterBegin_Success(t *testing.T) {
	t.Setenv("STEM_WEBAUTHN_RPID", "localhost")
	t.Setenv("STEM_WEBAUTHN_ORIGINS", "http://localhost:8080")
	s := setupAuthTestServer(t)
	token, _ := getAuthToken(t, s)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/webauthn/register/begin", nil)
	authorizeWithCSRF(t, s, req, token)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := resp["publicKey"]; !ok {
		t.Errorf("expected publicKey field in creation options, got %v", resp)
	}
}

// TestAuthRoutesRegistered_Wave3 verifies the new MFA endpoints are
// reachable (not 404).
func TestAuthRoutesRegistered_Wave3(t *testing.T) {
	s := setupAuthTestServer(t)

	routes := []struct {
		path   string
		method string
	}{
		{"/api/v1/auth/totp/setup", http.MethodPost},
		{"/api/v1/auth/totp/verify", http.MethodPost},
		{"/api/v1/auth/totp/disable", http.MethodPost},
		{"/api/v1/auth/login/totp", http.MethodPost},
		{"/api/v1/auth/mfa/status", http.MethodGet},
		{"/api/v1/auth/webauthn/register/begin", http.MethodPost},
		{"/api/v1/auth/webauthn/register/finish", http.MethodPost},
		{"/api/v1/auth/webauthn/login/begin", http.MethodPost},
		{"/api/v1/auth/webauthn/login/finish", http.MethodPost},
	}
	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()
			s.ServeHTTP(w, req)
			if w.Code == http.StatusNotFound {
				t.Errorf("route %s %s returned 404", route.method, route.path)
			}
		})
	}
}

// Static compile-time assertion that we only import api.Server.
var _ = &api.Server{}
