// SPDX-License-Identifier: BUSL-1.1

package license

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// stateKeyFor derives the key foundation encrypts activation state with:
// sha256(device fingerprint hash + product salt). Both halves are Stem's, so a
// test can rewrite a state file foundation wrote without a seam foundation
// does not offer.
func stateKeyFor(t *testing.T, mgr *Manager) []byte {
	t.Helper()
	sum := sha256.Sum256([]byte(mgr.GetFingerprint().Hash() + encryptionSalt))
	return sum[:]
}

// rewriteState decrypts the state file in dir, hands it to mutate, and writes
// it back in the same format. Used to age a trial past its end date, which no
// exported call can do.
func rewriteState(t *testing.T, dir string, mgr *Manager, mutate func(*ActivationState)) {
	t.Helper()
	path := filepath.Join(dir, licenseFileName)

	raw, readErr := os.ReadFile(filepath.Clean(path))
	if readErr != nil {
		t.Fatalf("read state: %v", readErr)
	}
	sealed, decodeErr := base64.StdEncoding.DecodeString(string(raw))
	if decodeErr != nil {
		t.Fatalf("decode state: %v", decodeErr)
	}

	gcm := stateGCM(t, mgr)
	if len(sealed) < gcm.NonceSize() {
		t.Fatal("state file shorter than a nonce")
	}
	plain, openErr := gcm.Open(nil, sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():], nil)
	if openErr != nil {
		t.Fatalf("decrypt state: %v", openErr)
	}

	state := &ActivationState{}
	if unmarshalErr := json.Unmarshal(plain, state); unmarshalErr != nil {
		t.Fatalf("parse state: %v", unmarshalErr)
	}
	mutate(state)

	updated, marshalErr := json.Marshal(state)
	if marshalErr != nil {
		t.Fatalf("marshal state: %v", marshalErr)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, nonceErr := io.ReadFull(rand.Reader, nonce); nonceErr != nil {
		t.Fatalf("nonce: %v", nonceErr)
	}
	sealedOut := base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, updated, nil))
	if writeErr := os.WriteFile(path, []byte(sealedOut), 0o600); writeErr != nil {
		t.Fatalf("write state: %v", writeErr)
	}
}

func stateGCM(t *testing.T, mgr *Manager) cipher.AEAD {
	t.Helper()
	block, blockErr := aes.NewCipher(stateKeyFor(t, mgr))
	if blockErr != nil {
		t.Fatalf("cipher: %v", blockErr)
	}
	gcm, gcmErr := cipher.NewGCM(block)
	if gcmErr != nil {
		t.Fatalf("gcm: %v", gcmErr)
	}
	return gcm
}

// startedTrial returns a directory holding a freshly started Professional
// trial, so a case can damage or age a state file that really was ours.
func startedTrial(t *testing.T) (string, *Manager) {
	t.Helper()
	dir := t.TempDir()
	mgr, _, err := LoadFromDir(dir)
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}
	if result := mgr.StartTrial(); !result.Success {
		t.Fatalf("StartTrial: %s", result.Message)
	}
	return dir, mgr
}

// dirWithNoLicence is a fresh install: nothing has ever been activated.
func dirWithNoLicence(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// dirWithUnreadableLicence holds a real state file the process cannot open.
func dirWithUnreadableLicence(t *testing.T) string {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("root reads a 0000 file; the case cannot be built")
	}
	dir, _ := startedTrial(t)
	if err := os.Chmod(filepath.Join(dir, licenseFileName), 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	return dir
}

// dirWithMalformedLicence holds a file in the right place that is not one.
func dirWithMalformedLicence(t *testing.T) string {
	t.Helper()
	dir, _ := startedTrial(t)
	path := filepath.Join(dir, licenseFileName)
	if err := os.WriteFile(path, []byte("not a licence"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	return dir
}

// dirWithExpiredLicence holds a state file that loads and parses cleanly and
// is past its end date, which no exported call can produce.
func dirWithExpiredLicence(t *testing.T) string {
	t.Helper()
	dir, mgr := startedTrial(t)
	rewriteState(t, dir, mgr, func(state *ActivationState) {
		state.TrialStartedAt = time.Now().AddDate(0, 0, -(TrialDays + 1))
		state.ExpiresAt = time.Now().AddDate(0, 0, -1)
	})
	return dir
}

// TestLoadFromDirFailsClosedToFree is the entitlement half of #1068: whatever
// is on disk, a state Stem cannot stand behind grants no paid standard. The
// cases are the four ways an install arrives without a usable licence.
func TestLoadFromDirFailsClosedToFree(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		wantStatus Status
	}{
		{name: "missing", setup: dirWithNoLicence, wantStatus: StatusMissing},
		{name: "unreadable", setup: dirWithUnreadableLicence, wantStatus: StatusUnreadable},
		{name: "malformed", setup: dirWithMalformedLicence, wantStatus: StatusMalformed},
		{name: "expired", setup: dirWithExpiredLicence, wantStatus: StatusLoaded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, status, err := LoadFromDir(tt.setup(t))
			if err != nil {
				t.Fatalf("LoadFromDir: %v", err)
			}
			if status != tt.wantStatus {
				t.Errorf("status = %v, want %v", status, tt.wantStatus)
			}
			if mgr.IsActivated() {
				t.Error("IsActivated() = true; an unusable state must not activate")
			}
			for _, feature := range ProFeatures() {
				if mgr.HasFeature(feature) {
					t.Errorf("HasFeature(%q) = true; want the Free grant only", feature)
				}
			}
		})
	}
}

// TestLoadFromDirReportsAUsableLicence is the control: the same call on a
// state Stem can use reports StatusLoaded and grants the Pro catalog, so the
// cases above fail closed because the state is unusable and not because
// LoadFromDir denies everything.
func TestLoadFromDirReportsAUsableLicence(t *testing.T) {
	dir, _ := startedTrial(t)

	mgr, status, err := LoadFromDir(dir)
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}
	if status != StatusLoaded {
		t.Fatalf("status = %v, want %v", status, StatusLoaded)
	}
	for _, feature := range ProFeatures() {
		if !mgr.HasFeature(feature) {
			t.Errorf("HasFeature(%q) = false on an active trial", feature)
		}
	}
}

// TestStatusNamesItself pins the strings an operator message and the startup
// log print, and the Usable predicate that decides whether a trial may start.
func TestStatusNamesItself(t *testing.T) {
	tests := []struct {
		status     Status
		wantName   string
		wantUsable bool
	}{
		{StatusLoaded, "loaded", true},
		{StatusMissing, "missing", true},
		{StatusUnreadable, "unreadable", false},
		{StatusMalformed, "malformed", false},
		{Status(99), "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.wantName, func(t *testing.T) {
			if got := tt.status.String(); got != tt.wantName {
				t.Errorf("String() = %q, want %q", got, tt.wantName)
			}
			if got := tt.status.Usable(); got != tt.wantUsable {
				t.Errorf("Usable() = %v, want %v", got, tt.wantUsable)
			}
		})
	}
}

// TestLoadUsesTheHomeConfigDirectory asserts the default path Load reads and
// DefaultLicensePath names are the same file, so the message an operator is
// told to fix points at the file the daemon actually opened.
func TestLoadUsesTheHomeConfigDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	want := filepath.Join(home, ".config", configSubdir, licenseFileName)
	if got := DefaultLicensePath(); got != want {
		t.Fatalf("DefaultLicensePath() = %q, want %q", got, want)
	}

	mgr, status, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if status != StatusMissing {
		t.Errorf("status = %v on an empty home, want %v", status, StatusMissing)
	}
	if result := mgr.StartTrial(); !result.Success {
		t.Fatalf("StartTrial: %s", result.Message)
	}
	if _, statErr := os.Stat(want); statErr != nil {
		t.Errorf("Load did not read %s: %v", want, statErr)
	}
}
