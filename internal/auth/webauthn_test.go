// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/MustardSeedNetworks/stem/internal/auth"
)

func TestPasskeyAccount_OpaqueHandle(t *testing.T) {
	mgr := mustNewManagerForTOTPTest(t)
	account := mgr.PasskeyAccount()
	handle := account.WebAuthnID()
	if len(handle) != 32 {
		t.Fatalf("user handle is %d bytes, want 32", len(handle))
	}
	if bytes.Contains(handle, []byte(account.WebAuthnName())) {
		t.Error("user handle must not carry the username")
	}
	if !bytes.Equal(mgr.PasskeyAccount().WebAuthnID(), handle) {
		t.Error("user handle must be stable for the manager's lifetime")
	}
	if bytes.Equal(mustNewManagerForTOTPTest(t).PasskeyAccount().WebAuthnID(), handle) {
		t.Error("each manager must draw its own random handle")
	}
}

func TestSavePasskey_ReplacesByID(t *testing.T) {
	mgr := mustNewManagerForTOTPTest(t)
	if mgr.HasWebAuthnCredentials() {
		t.Fatal("expected no credentials on a fresh manager")
	}
	mgr.SavePasskey(webauthn.Credential{ID: []byte{1}, PublicKey: []byte{9}})
	mgr.SavePasskey(webauthn.Credential{ID: []byte{2}, PublicKey: []byte{9}})
	mgr.SavePasskey(webauthn.Credential{
		ID: []byte{1}, PublicKey: []byte{9}, Authenticator: webauthn.Authenticator{SignCount: 7},
	})
	if got := mgr.WebAuthnCredentialCount(); got != 2 {
		t.Fatalf("WebAuthnCredentialCount() = %d, want 2", got)
	}
	stored := mgr.PasskeyAccount().WebAuthnCredentials()
	if stored[0].Authenticator.SignCount != 7 {
		t.Errorf("sign count = %d, want the replaced record's 7", stored[0].Authenticator.SignCount)
	}
}

func TestResolvePasskey(t *testing.T) {
	mgr := mustNewManagerForTOTPTest(t)
	mgr.SavePasskey(webauthn.Credential{ID: []byte{1}, PublicKey: []byte{9}})
	handle := mgr.PasskeyAccount().WebAuthnID()

	tests := []struct {
		name   string
		rawID  []byte
		handle []byte
		ok     bool
	}{
		{"own credential and handle", []byte{1}, handle, true},
		{"unknown credential", []byte{2}, handle, false},
		{"foreign handle", []byte{1}, bytes.Repeat([]byte{7}, len(handle)), false},
		{"no handle", []byte{1}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := mgr.ResolvePasskey(tt.rawID, tt.handle)
			if !tt.ok {
				if !errors.Is(err, auth.ErrPasskeyUnknown) {
					t.Fatalf("ResolvePasskey() error = %v, want ErrPasskeyUnknown", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolvePasskey() error = %v", err)
			}
			if !bytes.Equal(user.WebAuthnID(), handle) || len(user.WebAuthnCredentials()) != 1 {
				t.Errorf("resolved %x with %d credentials", user.WebAuthnID(), len(user.WebAuthnCredentials()))
			}
		})
	}
}
