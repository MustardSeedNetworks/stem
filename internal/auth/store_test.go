// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/pquerna/otp/totp"

	"github.com/MustardSeedNetworks/foundation/pkg/credstore"
	"github.com/MustardSeedNetworks/foundation/pkg/password"

	"github.com/MustardSeedNetworks/stem/internal/auth"
)

const storeTestPassword = "TestPassw0rd!123"

func openStore(t *testing.T, dir string) *credstore.Store {
	t.Helper()
	trust, err := credstore.ProcessTrust()
	if err != nil {
		t.Fatalf("ProcessTrust: %v", err)
	}
	store, err := credstore.Open(dir, credstore.Options{Trust: trust})
	if err != nil {
		t.Fatalf("credstore.Open: %v", err)
	}
	return store
}

// reopenManager is a restart: a new manager over whatever the directory holds.
func reopenManager(t *testing.T, dir string) *auth.Manager {
	t.Helper()
	mgr, err := auth.NewManagerFromStore("", time.Hour, openStore(t, dir), auth.FirstRunUsername)
	if err != nil {
		t.Fatalf("NewManagerFromStore: %v", err)
	}
	t.Cleanup(mgr.Stop)
	return mgr
}

// newStoreManager builds a claimed manager over a credential store in dir.
func newStoreManager(t *testing.T, dir, username, plaintext string) *auth.Manager {
	t.Helper()
	mgr, err := auth.NewManagerFromStore("", time.Hour, openStore(t, dir), username)
	if err != nil {
		t.Fatalf("NewManagerFromStore: %v", err)
	}
	t.Cleanup(mgr.Stop)
	hash, err := password.Hash(plaintext)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if err = mgr.ReplaceCredential(hash, false); err != nil {
		t.Fatalf("claim: %v", err)
	}
	return mgr
}

func savePasskey(t *testing.T, mgr *auth.Manager, credential webauthn.Credential) {
	t.Helper()
	if err := mgr.SavePasskey(credential); err != nil {
		t.Fatalf("SavePasskey: %v", err)
	}
}

func enrolTOTP(t *testing.T, mgr *auth.Manager) string {
	t.Helper()
	setup, err := auth.GenerateTOTPSecret("admin", auth.DefaultTOTPIssuer)
	if err != nil {
		t.Fatal(err)
	}
	if err = mgr.SetTOTPSecret(setup.Secret); err != nil {
		t.Fatal(err)
	}
	code, err := totp.GenerateCode(setup.Secret, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err = mgr.VerifyAndEnableTOTP(code); err != nil {
		t.Fatalf("VerifyAndEnableTOTP: %v", err)
	}
	return code
}

func TestStore_TOTPAndPasskeysSurviveARestart(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "auth")
	mgr := newStoreManager(t, dir, "operator", storeTestPassword)
	code := enrolTOTP(t, mgr)
	savePasskey(t, mgr, webauthn.Credential{ID: []byte{1}, PublicKey: []byte{9}})
	savePasskey(t, mgr, webauthn.Credential{ID: []byte{2}, PublicKey: []byte{9}})
	savePasskey(t, mgr, webauthn.Credential{
		ID: []byte{1}, PublicKey: []byte{9}, Authenticator: webauthn.Authenticator{SignCount: 7},
	})
	handle := mgr.PasskeyAccount().WebAuthnID()

	restarted := reopenManager(t, dir)

	if got := restarted.GetUsername(); got != "operator" {
		t.Errorf("username after restart = %q, want operator", got)
	}
	if _, err := restarted.Authenticate("operator", storeTestPassword); err != nil {
		t.Errorf("password login after restart: %v", err)
	}
	if !restarted.TOTPEnabled() {
		t.Fatal("TOTP is off after a restart")
	}
	if err := restarted.VerifyTOTPCode(code); err != nil {
		t.Errorf("enrolled secret rejected after restart: %v", err)
	}
	if got := restarted.WebAuthnCredentialCount(); got != 2 {
		t.Fatalf("passkeys after restart = %d, want 2", got)
	}
	if !bytes.Equal(restarted.PasskeyAccount().WebAuthnID(), handle) {
		t.Error("the passkey user handle changed across a restart; stored passkeys could never match an assertion")
	}
	if _, err := restarted.ResolvePasskey([]byte{1}, handle); err != nil {
		t.Errorf("ResolvePasskey after restart: %v", err)
	}
	if got := restarted.PasskeyAccount().WebAuthnCredentials()[0].Authenticator.SignCount; got != 7 {
		t.Errorf("sign count after restart = %d, want 7: the counter is the clone defence", got)
	}
}

func TestStore_RecoveryClearsTheSecondFactors(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "auth")
	mgr := newStoreManager(t, dir, "admin", storeTestPassword)
	enrolTOTP(t, mgr)
	savePasskey(t, mgr, webauthn.Credential{ID: []byte{1}, PublicKey: []byte{9}})
	if err := mgr.SetTOTPSecret("STAGED"); err == nil {
		t.Fatal("staging over an enabled factor must be refused")
	}

	recoveredHash, err := password.Hash("RecoveredPassw0rd!456")
	if err != nil {
		t.Fatal(err)
	}
	if err = mgr.ReplaceCredential(recoveredHash, true); err != nil {
		t.Fatalf("recovery: %v", err)
	}

	for name, m := range map[string]*auth.Manager{"in memory": mgr, "after a restart": reopenManager(t, dir)} {
		if m.TOTPEnabled() || m.HasWebAuthnCredentials() {
			t.Errorf("%s: second factors survive recovery (totp %v, passkeys %d)",
				name, m.TOTPEnabled(), m.WebAuthnCredentialCount())
		}
		if _, authErr := m.Authenticate("admin", "RecoveredPassw0rd!456"); authErr != nil {
			t.Errorf("%s: the recovered password does not log in: %v", name, authErr)
		}
		if _, authErr := m.Authenticate("admin", storeTestPassword); authErr == nil {
			t.Errorf("%s: the old password still logs in", name)
		}
	}
}

func TestStore_SettingAPasswordKeepsTheSecondFactors(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "auth")
	mgr := newStoreManager(t, dir, "admin", storeTestPassword)
	enrolTOTP(t, mgr)
	savePasskey(t, mgr, webauthn.Credential{ID: []byte{1}, PublicKey: []byte{9}})
	hash, err := password.Hash("AnotherPassw0rd!789")
	if err != nil {
		t.Fatal(err)
	}
	if err = mgr.ReplaceCredential(hash, false); err != nil {
		t.Fatal(err)
	}
	restarted := reopenManager(t, dir)
	if !restarted.TOTPEnabled() || restarted.WebAuthnCredentialCount() != 1 {
		t.Errorf("a password change dropped a factor (totp %v, passkeys %d)",
			restarted.TOTPEnabled(), restarted.WebAuthnCredentialCount())
	}
}

func TestStore_AStagedSecretIsNotPersisted(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "auth")
	mgr := newStoreManager(t, dir, "admin", storeTestPassword)
	if err := mgr.SetTOTPSecret("JBSWY3DPEHPK3PXP"); err != nil {
		t.Fatal(err)
	}
	// A passkey saved mid-enrolment refreshes the working copy from the store;
	// the staged secret must survive that, or the operator's code fails.
	savePasskey(t, mgr, webauthn.Credential{ID: []byte{1}, PublicKey: []byte{9}})
	if secret, ok := mgr.GetTOTPSecret(); !ok || secret != "JBSWY3DPEHPK3PXP" {
		t.Errorf("staged secret after a passkey save = %q, %v", secret, ok)
	}

	restarted := reopenManager(t, dir)
	if _, ok := restarted.GetTOTPSecret(); ok || restarted.TOTPEnabled() {
		t.Error("an unverified secret reached the store")
	}
}

func TestStore_DisablingTOTPIsDurable(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "auth")
	mgr := newStoreManager(t, dir, "admin", storeTestPassword)
	enrolTOTP(t, mgr)
	if err := mgr.DisableTOTP(); err != nil {
		t.Fatal(err)
	}
	if reopenManager(t, dir).TOTPEnabled() {
		t.Error("TOTP came back after a restart although it was disabled")
	}
}

func TestStore_SetupClaimsAnUnclaimedDaemonAcrossARestart(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "auth")
	mgr, err := auth.NewManagerFromStore("", time.Hour, openStore(t, dir), auth.FirstRunUsername)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mgr.Stop)
	if mgr.GetPasswordHash() != "" {
		t.Fatal("a fresh store must be unclaimed")
	}
	hash, err := password.Hash(storeTestPassword)
	if err != nil {
		t.Fatal(err)
	}
	if err = mgr.ReplaceCredential(hash, false); err != nil {
		t.Fatal(err)
	}
	restarted := reopenManager(t, dir)
	if restarted.GetPasswordHash() != hash || restarted.GetUsername() != auth.FirstRunUsername {
		t.Errorf("claim lost across a restart: user %q, hash kept %v",
			restarted.GetUsername(), restarted.GetPasswordHash() == hash)
	}
}

func TestStore_ManagerWithoutAStoreRefusesToChangeTheAccount(t *testing.T) {
	t.Parallel()
	mgr := mustNewMemoryManager(t)
	setup, err := auth.GenerateTOTPSecret("admin", auth.DefaultTOTPIssuer)
	if err != nil {
		t.Fatal(err)
	}
	if err = mgr.SetTOTPSecret(setup.Secret); err != nil {
		t.Fatal(err)
	}
	code, err := totp.GenerateCode(setup.Secret, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err = mgr.VerifyAndEnableTOTP(code); !errors.Is(err, auth.ErrNoCredentialStore) {
		t.Fatalf("VerifyAndEnableTOTP = %v, want ErrNoCredentialStore", err)
	}
	if mgr.TOTPEnabled() {
		t.Error("TOTP is on although it could not be stored")
	}
}

func mustNewMemoryManager(t *testing.T) *auth.Manager {
	t.Helper()
	mgr, err := auth.NewManager("", time.Hour, "admin", storeTestPassword)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mgr.Stop)
	return mgr
}
