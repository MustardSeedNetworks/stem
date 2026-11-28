// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/pquerna/otp/totp"

	"github.com/MustardSeedNetworks/stem/internal/auth"
)

const (
	adminStatePassword  = "kettle-orbit-granite-velvet-42"
	adminStateRecovered = "lantern-quartz-meadow-falcon-77"
)

func adminStateEnv(t *testing.T, username, plaintext string) string {
	t.Helper()
	dataDir := t.TempDir()
	t.Setenv("STEM_DATA_DIR", dataDir)
	t.Setenv("STEM_AUTH_USERNAME", username)
	t.Setenv("STEM_AUTH_PASSWORD", plaintext)
	return dataDir
}

func enrolFactors(t *testing.T, m *auth.Manager) string {
	t.Helper()
	setup, err := auth.GenerateTOTPSecret("operator", auth.DefaultTOTPIssuer)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.SetTOTPSecret(setup.Secret); err != nil {
		t.Fatal(err)
	}
	code, err := totp.GenerateCode(setup.Secret, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err = m.VerifyAndEnableTOTP(code); err != nil {
		t.Fatal(err)
	}
	if err = m.SavePasskey(webauthn.Credential{ID: []byte{1}, PublicKey: []byte{9}}); err != nil {
		t.Fatal(err)
	}
	return code
}

// The environment is a one-time seed: the account it creates is the store's,
// so the second factors enrolled on it survive a restart, and a changed or
// removed environment neither reverts nor loses it.
func TestEnvironmentSeedsTheStoreOnce(t *testing.T) {
	dataDir := adminStateEnv(t, "operator", adminStatePassword)
	first, err := newAuthManager(dataDir)
	if err != nil {
		t.Fatalf("newAuthManager: %v", err)
	}
	t.Cleanup(first.Stop)
	code := enrolFactors(t, first)

	t.Setenv("STEM_AUTH_USERNAME", "")
	t.Setenv("STEM_AUTH_PASSWORD", "")
	second, err := newAuthManager(dataDir)
	if err != nil {
		t.Fatalf("newAuthManager after a restart: %v", err)
	}
	t.Cleanup(second.Stop)

	if _, err = second.Authenticate("operator", adminStatePassword); err != nil {
		t.Errorf("seeded password after restart: %v", err)
	}
	if !second.TOTPEnabled() || second.VerifyTOTPCode(code) != nil {
		t.Error("the TOTP secret enrolled on an environment-seeded account was lost on restart")
	}
	if !second.HasWebAuthnCredentials() {
		t.Error("the passkey enrolled on an environment-seeded account was lost on restart")
	}
}

// An old credentials.json carries no second factor and is not read: the daemon
// comes up unclaimed, and the file goes once the store holds an account.
func TestLegacyCredentialFileIsNotReadAndIsRemoved(t *testing.T) {
	dataDir := adminStateEnv(t, "", "")
	legacy := filepath.Join(dataDir, legacyCredentialFile)
	hash := "$argon2id$v=19$m=8192,t=1,p=1$YWJjZGVmZ2hpamtsbW5vcA$aGFzaC1zdHViLWZvci10ZXN0aW5nLW9ubHkAAAAAAA"
	if err := os.WriteFile(legacy, []byte(`{"username":"admin","passwordHash":"`+hash+`"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	m, err := newAuthManager(dataDir)
	if err != nil {
		t.Fatalf("newAuthManager: %v", err)
	}
	t.Cleanup(m.Stop)
	if m.GetPasswordHash() != "" {
		t.Fatal("the old credentials.json was read; an install that has one must be set up again")
	}
	if _, statErr := os.Stat(legacy); statErr != nil {
		t.Fatalf("the old file went before the store held an account: %v", statErr)
	}

	t.Setenv("STEM_AUTH_USERNAME", "operator")
	t.Setenv("STEM_AUTH_PASSWORD", adminStatePassword)
	seeded, err := newAuthManager(dataDir)
	if err != nil {
		t.Fatalf("newAuthManager with the environment set: %v", err)
	}
	t.Cleanup(seeded.Stop)
	if _, statErr := os.Stat(legacy); !os.IsNotExist(statErr) {
		t.Errorf("the old credentials.json survives once the store holds an account: %v", statErr)
	}
}

// Recovery is the owner's decision of 2026-11-28: the new password and the
// loss of both second factors land together, and stay lost after a restart.
func TestRecoveryClearsSecondFactorsAcrossARestart(t *testing.T) {
	dataDir := adminStateEnv(t, "operator", adminStatePassword)
	s, err := NewServer(ListenAddr{Port: 8444})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	enrolFactors(t, s.authManager)

	if err = os.WriteFile(filepath.Join(dataDir, ".recovery"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	serve := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		return w
	}
	if w := serve(http.MethodGet, "/api/v1/recovery/status", ""); w.Code != http.StatusOK {
		t.Fatalf("recovery status = %d: %s", w.Code, w.Body.String())
	}
	token, err := os.ReadFile(filepath.Join(dataDir, ".recovery-token"))
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(
		RecoveryCompleteRequest{Token: strings.TrimSpace(string(token)), Password: adminStateRecovered},
	)
	if err != nil {
		t.Fatal(err)
	}
	if w := serve(http.MethodPost, "/api/v1/recovery/complete", string(body)); w.Code != http.StatusOK {
		t.Fatalf("recovery complete = %d: %s", w.Code, w.Body.String())
	}
	if s.authManager.TOTPEnabled() || s.authManager.HasWebAuthnCredentials() {
		t.Error("second factors survive recovery in the running daemon")
	}
	_ = s.Shutdown()

	t.Setenv("STEM_AUTH_USERNAME", "")
	t.Setenv("STEM_AUTH_PASSWORD", "")
	restarted, err := newAuthManager(dataDir)
	if err != nil {
		t.Fatalf("newAuthManager after recovery: %v", err)
	}
	t.Cleanup(restarted.Stop)
	if restarted.TOTPEnabled() || restarted.HasWebAuthnCredentials() {
		t.Error("second factors came back after a restart")
	}
	if _, err = restarted.Authenticate("operator", adminStateRecovered); err != nil {
		t.Errorf("recovered password after restart: %v", err)
	}
}
