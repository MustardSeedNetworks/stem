// SPDX-License-Identifier: BUSL-1.1

package authstate

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/MustardSeedNetworks/foundation/pkg/passkey"
	"github.com/go-webauthn/webauthn/webauthn"
)

// MaxRecordBytes matches the ceremony input budget for one local principal.
const MaxRecordBytes = 1 << 20

const userHandleBytes = 32

// ErrInvalid never means missing state or permission to bootstrap.
var ErrInvalid = errors.New("invalid auth state")

// State is authoritative once created; environment settings must not replace it.
// PasswordHash is opaque here: the password factory/verifier owns PHC validation.
type State struct {
	Username     string       `json:"username"`
	PasswordHash string       `json:"passwordHash"`
	UserHandle   []byte       `json:"userHandle"`
	Epoch        uint64       `json:"epoch"`
	Credentials  []Credential `json:"credentials"`
}

// Credential retains product metadata alongside the complete library record.
type Credential struct {
	Name      string              `json:"name"`
	CreatedAt time.Time           `json:"createdAt"`
	Value     webauthn.Credential `json:"value"`
}

type (
	stateFields      State
	credentialFields Credential
)

type storedCredential struct {
	credentialFields

	Value json.RawMessage `json:"value"`
}

type storedState struct {
	stateFields

	Credentials []storedCredential `json:"credentials"`
}

// Save validates before protected publication. Callers serialize mutations and
// publish in-memory state only after success; ErrUncertain requires reload.
func (state State) Save(directory string) error {
	record, err := state.Encode()
	if err != nil {
		return err
	}
	return Publish(directory, record)
}

// Encode supplies the current schema, without legacy migration or hash fallback.
func (state State) Encode() ([]byte, error) {
	if !state.valid() {
		return nil, ErrInvalid
	}
	stored := storedState{stateFields: stateFields(state)}
	remaining := MaxRecordBytes - len(state.Username) - len(state.PasswordHash)
	for _, credential := range state.Credentials {
		value, err := passkey.MarshalCredential(credential.Value)
		if err != nil || len(value) > remaining || len(credential.Name) > remaining-len(value) {
			return nil, ErrInvalid
		}
		remaining -= len(value) + len(credential.Name)
		stored.Credentials = append(stored.Credentials, storedCredential{credentialFields(credential), value})
	}
	record, err := json.Marshal(stored)
	if err != nil || len(record) > MaxRecordBytes {
		return nil, ErrInvalid
	}
	return record, nil
}

// Decode validates bytes only; the secure filesystem loader remains separate.
func Decode(record []byte) (State, error) {
	if len(record) > MaxRecordBytes || !utf8.Valid(record) {
		return State{}, ErrInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(record))
	decoder.DisallowUnknownFields()
	var stored storedState
	if err := decoder.Decode(&stored); err != nil {
		return State{}, ErrInvalid
	}
	if err := decoder.Decode(new(json.RawMessage)); !errors.Is(err, io.EOF) {
		return State{}, ErrInvalid
	}
	state := State(stored.stateFields)
	for _, credential := range stored.Credentials {
		value, err := passkey.UnmarshalCredential(credential.Value)
		if err != nil {
			return State{}, ErrInvalid
		}
		state.Credentials = append(state.Credentials, Credential{credential.Name, credential.CreatedAt, value})
	}
	if !state.valid() {
		return State{}, ErrInvalid
	}
	return state, nil
}

func (state State) valid() bool {
	if !validText(state.Username) || !validText(state.PasswordHash) || state.Epoch == 0 ||
		len(state.UserHandle) != userHandleBytes || bytes.Equal(state.UserHandle, make([]byte, userHandleBytes)) ||
		len(state.Username) > MaxRecordBytes || len(state.PasswordHash) > MaxRecordBytes-len(state.Username) ||
		len(state.Credentials) > MaxRecordBytes/len(`{"name":"","createdAt":"","value":{}}`) {
		return false
	}
	seen := make(map[string]bool)
	remaining := MaxRecordBytes
	for _, credential := range state.Credentials {
		if !validText(credential.Name) || credential.CreatedAt.IsZero() ||
			len(credential.Value.ID) == 0 || len(credential.Value.ID) > remaining ||
			len(credential.Value.PublicKey) == 0 || seen[string(credential.Value.ID)] {
			return false
		}
		remaining -= len(credential.Value.ID)
		seen[string(credential.Value.ID)] = true
	}
	return true
}

func validText(value string) bool {
	return value != "" && strings.TrimSpace(value) == value && utf8.ValidString(value)
}
