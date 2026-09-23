// SPDX-License-Identifier: BUSL-1.1

package authstate_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/MustardSeedNetworks/stem/internal/authstate"
)

func validState() authstate.State {
	return authstate.State{
		Username: "operator", PasswordHash: "opaque-not-format-validated-here",
		UserHandle: bytes.Repeat([]byte{3}, 32), Epoch: 1,
		Credentials: []authstate.Credential{{
			Name: "Primary key", CreatedAt: time.Unix(1234, 0).UTC(),
			Value: webauthn.Credential{
				ID: []byte{1}, PublicKey: []byte{2}, Flags: webauthn.NewCredentialFlags(255),
				Transport:     []protocol.AuthenticatorTransport{protocol.Internal},
				Authenticator: webauthn.Authenticator{SignCount: 10, CloneWarning: true, AAGUID: []byte{4}},
				Attestation: webauthn.CredentialAttestation{
					ClientDataJSON: []byte{5}, ClientDataHash: []byte{6},
					AuthenticatorData: []byte{7}, PublicKeyAlgorithm: -7, Object: []byte{8},
				},
				AttestationType: "basic_full", AttestationFormat: "packed",
				Extensions: webauthn.CredentialExtensions{RK: new(false), PRFEnabled: new(true)},
			},
		}},
	}
}

func TestStateSaveRoundTrip(t *testing.T) {
	t.Parallel()
	state := validState()
	directory := t.TempDir()
	if err := state.Save(directory); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(directory, "auth.json"))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := authstate.Decode(data)
	if err != nil || !reflect.DeepEqual(state, decoded) {
		t.Fatalf("state did not round trip: %v", err)
	}
}

func TestInvalidStateDoesNotReplaceRecord(t *testing.T) {
	t.Parallel()
	for name, mutate := range map[string]func(*authstate.State){
		"username":     func(s *authstate.State) { s.Username = " " },
		"usernameUTF8": func(s *authstate.State) { s.Username = "\xff" },
		"hash":         func(s *authstate.State) { s.PasswordHash = "" },
		"handleLength": func(s *authstate.State) { s.UserHandle = []byte{1} },
		"handleZero":   func(s *authstate.State) { s.UserHandle = make([]byte, 32) },
		"epoch":        func(s *authstate.State) { s.Epoch = 0 },
		"name":         func(s *authstate.State) { s.Credentials[0].Name = "" },
		"created":      func(s *authstate.State) { s.Credentials[0].CreatedAt = time.Time{} },
		"id":           func(s *authstate.State) { s.Credentials[0].Value.ID = nil },
		"key":          func(s *authstate.State) { s.Credentials[0].Value.PublicKey = nil },
		"duplicate":    func(s *authstate.State) { s.Credentials = append(s.Credentials, s.Credentials[0]) },
		"large":        func(s *authstate.State) { s.PasswordHash = strings.Repeat("x", authstate.MaxRecordBytes+1) },
		"flags":        func(s *authstate.State) { s.Credentials[0].Value.Flags.UserVerified = false },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			state := validState()
			if err := state.Save(directory); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(filepath.Join(directory, "auth.json"))
			if err != nil {
				t.Fatal(err)
			}
			mutate(&state)
			if err = state.Save(directory); !errors.Is(err, authstate.ErrInvalid) {
				t.Fatalf("invalid state accepted: %v", err)
			}
			after, err := os.ReadFile(filepath.Join(directory, "auth.json"))
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("invalid save changed existing state")
			}
		})
	}
}
