// SPDX-License-Identifier: BUSL-1.1

package update_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/update"
)

// An install that never chose is off: the timer must find nothing to do.
func TestLoadSettingsDefaultsToOff(t *testing.T) {
	got, err := update.LoadSettings(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got.Automatic || got.SourceSpec() != update.DefaultSource {
		t.Fatalf("LoadSettings with no file = %+v (source %q); want off from %q",
			got, got.SourceSpec(), update.DefaultSource)
	}
}

func TestSaveSettingsRoundTrips(t *testing.T) {
	dir := t.TempDir()
	for _, want := range []update.Settings{
		{Automatic: true, Source: "/srv/stem-releases"},
		{Automatic: false, Source: "https://mirror.example/stem"},
		{Automatic: true},
	} {
		if err := update.SaveSettings(dir, want); err != nil {
			t.Fatal(err)
		}
		got, err := update.LoadSettings(dir)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("LoadSettings after SaveSettings(%+v) = %+v", want, got)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, update.SettingsName+".tmp")); !os.IsNotExist(err) {
		t.Fatalf("the temporary file is left behind: %v", err)
	}
}

// The file is writable by the service account and read by root, so anything
// that is not exactly the settings is refused rather than guessed at.
func TestLoadSettingsRefusesWhatItCannotTrust(t *testing.T) {
	for _, tc := range []struct {
		name    string
		prepare func(t *testing.T, path string)
		wantErr string
	}{
		{
			name: "unknown field",
			prepare: func(t *testing.T, path string) {
				writeFile(t, path, `{"automatic":true,"skipVerification":true}`)
			},
			wantErr: "unknown field",
		},
		{
			name: "not JSON",
			prepare: func(t *testing.T, path string) {
				writeFile(t, path, "automatic=true")
			},
			wantErr: "not valid",
		},
		{
			name: "oversized",
			prepare: func(t *testing.T, path string) {
				writeFile(t, path, `{"source":"`+strings.Repeat("a", 5000)+`"}`)
			},
			wantErr: "larger than",
		},
		{
			name: "symlink out of the data directory",
			prepare: func(t *testing.T, path string) {
				outside := filepath.Join(t.TempDir(), "secret")
				writeFile(t, outside, `{"automatic":true}`)
				if err := os.Symlink(outside, path); err != nil {
					t.Skipf("symlinks unavailable: %v", err)
				}
			},
			wantErr: "read the update settings",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			tc.prepare(t, filepath.Join(dir, update.SettingsName))

			got, err := update.LoadSettings(dir)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("LoadSettings = %+v, %v; want an error naming %q", got, err, tc.wantErr)
			}
			if got.Automatic {
				t.Fatalf("LoadSettings returned automatic updates on alongside %v", err)
			}
		})
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
