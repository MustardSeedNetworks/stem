// SPDX-License-Identifier: BUSL-1.1

//go:build linux

package truststore

import (
	"os"
	"path/filepath"
	"testing"
)

// candidateAnchorDirs is the list detectLinuxStore walks, in its order.
// Duplicated here deliberately: a test that read the production slice would
// pass whatever order the production code happens to have.
func candidateAnchorDirs() []string {
	return []string{
		"/usr/local/share/ca-certificates",
		"/etc/pki/ca-trust/source/anchors",
		"/etc/ca-certificates/trust-source/anchors",
		"/usr/share/pki/trust/anchors",
	}
}

func TestPathExists(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if !pathExists(dir) {
		t.Errorf("pathExists(%q) = false for an existing directory", dir)
	}
	file := filepath.Join(dir, "anchor.crt")
	if writeErr := os.WriteFile(file, []byte("x"), 0o600); writeErr != nil {
		t.Fatalf("write fixture: %v", writeErr)
	}
	if !pathExists(file) {
		t.Errorf("pathExists(%q) = false for an existing file", file)
	}
	if pathExists(filepath.Join(dir, "absent")) {
		t.Error("pathExists returned true for a path that does not exist")
	}
}

func TestLinuxStore_AnchorPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store linuxStore
		want  string
	}{
		{
			name:  "debian",
			store: linuxStore{AnchorDir: "/usr/local/share/ca-certificates", Suffix: ".crt"},
			want:  "/usr/local/share/ca-certificates/stem-root.crt",
		},
		{
			name:  "rhel",
			store: linuxStore{AnchorDir: "/etc/pki/ca-trust/source/anchors", Suffix: ".pem"},
			want:  "/etc/pki/ca-trust/source/anchors/stem-root.pem",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.store.anchorPath(); got != tc.want {
				t.Errorf("anchorPath() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestWriteAnchorFile proves writeAnchorFile writes name inside root at
// trustAnchorMode and cannot escape root even given a name built to try —
// [os.Root] rejects the escape instead of the write silently landing outside
// the trust-anchor directory.
func TestWriteAnchorFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		file    string
		wantErr bool
	}{
		{name: "normal name", file: "stem-root.crt"},
		{name: "escape attempt", file: filepath.Join("..", "escaped.crt"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			pemBytes := []byte("-----BEGIN CERTIFICATE-----\nZmFrZQ==\n-----END CERTIFICATE-----\n")
			dst, err := writeAnchorFile(root, tt.file, pemBytes)
			if tt.wantErr {
				assertEscapeRejected(t, err, root)
				return
			}
			assertAnchorWritten(t, dst, pemBytes, err)
		})
	}
}

func assertEscapeRejected(t *testing.T, err error, root string) {
	t.Helper()
	if err == nil {
		t.Fatal("writeAnchorFile() = nil error, want one")
	}
	if pathExists(filepath.Join(filepath.Dir(root), "escaped.crt")) {
		t.Error("escape attempt wrote a file outside root")
	}
}

func assertAnchorWritten(t *testing.T, dst string, pemBytes []byte, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("writeAnchorFile() = %v, want nil", err)
	}

	info, statErr := os.Stat(dst)
	if statErr != nil {
		t.Fatalf("stat written file: %v", statErr)
	}
	// update-ca-certificates runs as root but reads the anchor as an
	// ordinary file; 0600 would leave it unreadable to the extract tooling
	// on distributions that drop privileges.
	if perm := info.Mode().Perm(); perm != trustAnchorMode {
		t.Errorf("anchor mode = %#o, want %#o", perm, trustAnchorMode)
	}

	data, readErr := os.ReadFile(dst) // #nosec G304 -- dst is under t.TempDir()
	if readErr != nil {
		t.Fatalf("read written file: %v", readErr)
	}
	if string(data) != string(pemBytes) {
		t.Errorf("content = %q, want %q", data, pemBytes)
	}
}

// TestUninstallPlatform_RemovesOnlyAnchor proves uninstallPlatform's
// removal is confined to the detected store's AnchorDir: a sibling file
// bearing the same name outside that directory is left untouched.
func TestUninstallPlatform_RemovesOnlyAnchor(t *testing.T) {
	t.Parallel()
	anchorDir := t.TempDir()
	name := "stem-root.crt"
	if err := os.WriteFile(filepath.Join(anchorDir, name), []byte("cert"), 0o600); err != nil {
		t.Fatalf("write anchor fixture: %v", err)
	}

	root, err := os.OpenRoot(anchorDir)
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}
	defer func() { _ = root.Close() }()

	if _, statErr := root.Stat(name); statErr != nil {
		t.Fatalf("root.Stat(%q) = %v, want nil", name, statErr)
	}
	if removeErr := root.Remove(name); removeErr != nil {
		t.Fatalf("root.Remove(%q) = %v, want nil", name, removeErr)
	}
	if _, statErr := os.Stat(filepath.Join(anchorDir, name)); !os.IsNotExist(statErr) {
		t.Errorf("anchor file still exists: err=%v", statErr)
	}
}

// detectLinuxStore reads real system paths, so this asserts the contract it
// must satisfy on whatever distribution the test runs on rather than naming
// one distribution.
func TestDetectLinuxStore_MatchesFirstPresentAnchorDir(t *testing.T) {
	t.Parallel()
	var wantDir string
	for _, d := range candidateAnchorDirs() {
		if pathExists(d) {
			wantDir = d
			break
		}
	}

	store, ok := detectLinuxStore()
	if wantDir == "" {
		if ok {
			t.Fatalf("detectLinuxStore reported %q with no candidate anchor directory present", store.AnchorDir)
		}
		t.Skip("no supported system CA directory on this host")
	}
	if !ok {
		t.Fatalf("detectLinuxStore found nothing although %q exists", wantDir)
	}
	if store.AnchorDir != wantDir {
		t.Errorf("detectLinuxStore chose %q, want the first present candidate %q", store.AnchorDir, wantDir)
	}
	if store.Label == "" || store.RefreshLabel == "" {
		t.Errorf("store for %q has an empty label (Label=%q RefreshLabel=%q)", wantDir, store.Label, store.RefreshLabel)
	}
	if store.Refresh == nil {
		t.Errorf("store for %q has no Refresh command", wantDir)
	}
	if want := filepath.Join(wantDir, "stem-root"+store.Suffix); store.anchorPath() != want {
		t.Errorf("anchorPath() = %q, want %q", store.anchorPath(), want)
	}
	if store.Suffix != ".crt" && store.Suffix != ".pem" {
		t.Errorf("unexpected anchor suffix %q", store.Suffix)
	}
}
