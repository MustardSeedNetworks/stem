// SPDX-License-Identifier: BUSL-1.1

package authstate_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/authstate"
)

func TestDecodeRejectsInvalidState(t *testing.T) {
	t.Parallel()
	encoded, err := validState().Encode()
	if err != nil {
		t.Fatal(err)
	}
	valid := string(encoded)
	for name, record := range map[string]string{
		"empty": "", "null": "null", "incomplete": `{}`, "trailing": valid + `{}`,
		"unknown":           strings.TrimSuffix(valid, "}") + `,"unknown":true}`,
		"epoch":             strings.Replace(valid, `"epoch":1`, `"epoch":0`, 1),
		"username":          strings.Replace(valid, `"username":"operator"`, `"username":""`, 1),
		"hash":              strings.Replace(valid, `"passwordHash":"opaque-not-format-validated-here"`, `"passwordHash":""`, 1),
		"name":              strings.Replace(valid, `"name":"Primary key"`, `"name":""`, 1),
		"flags":             strings.Replace(valid, `"flags":255`, `"flags":null`, 1),
		"unknownCredential": strings.Replace(valid, `"name":"Primary key"`, `"unknown":1,"name":"Primary key"`, 1),
		"oversized":         strings.Repeat(" ", authstate.MaxRecordBytes+1),
		"invalidUTF8":       "\xff",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			state, decodeErr := authstate.Decode([]byte(record))
			if !errors.Is(decodeErr, authstate.ErrInvalid) || state.Username != "" {
				t.Fatalf("invalid state exposed: %v", decodeErr)
			}
		})
	}
}

func TestPasswordOnlyState(t *testing.T) {
	t.Parallel()
	state := validState()
	state.Credentials = nil
	encoded, err := state.Encode()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := authstate.Decode(encoded)
	if err != nil || decoded.PasswordHash != state.PasswordHash || len(decoded.Credentials) != 0 {
		t.Fatalf("password-only state lost: %v", err)
	}
	if err = state.Save(t.TempDir() + "/missing"); err == nil {
		t.Fatal("save ignored publication failure")
	}
}

func TestStateEncodedSizeLimit(t *testing.T) {
	t.Parallel()
	state := validState()
	state.PasswordHash = strings.Repeat("\x00", authstate.MaxRecordBytes/2)
	if _, err := state.Encode(); !errors.Is(err, authstate.ErrInvalid) {
		t.Fatalf("accepted escaping expansion: %v", err)
	}
	state = validState()
	state.Credentials[0].Name = strings.Repeat("x", authstate.MaxRecordBytes)
	if _, err := state.Encode(); !errors.Is(err, authstate.ErrInvalid) {
		t.Fatalf("accepted combined oversized record: %v", err)
	}
}
