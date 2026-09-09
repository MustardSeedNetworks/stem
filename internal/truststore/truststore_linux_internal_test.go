// SPDX-License-Identifier: BUSL-1.1

//go:build linux

package truststore

import (
	"os"
	"path/filepath"
	"strings"
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

func TestWriteAnchorFile_WritesWorldReadable(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dst := filepath.Join(root, "stem-root.crt")
	pemBytes := []byte("-----BEGIN CERTIFICATE-----\nZmFrZQ==\n-----END CERTIFICATE-----\n")

	got, err := writeAnchorFile(root, dst, pemBytes)
	if err != nil {
		t.Fatalf("writeAnchorFile: %v", err)
	}
	if got != dst {
		t.Errorf("writeAnchorFile returned %q, want %q", got, dst)
	}
	onDisk, readErr := os.ReadFile(dst) // #nosec G304 -- path is t.TempDir()
	if readErr != nil {
		t.Fatalf("read back: %v", readErr)
	}
	if string(onDisk) != string(pemBytes) {
		t.Errorf("contents round-tripped as %q", onDisk)
	}
	info, statErr := os.Stat(dst)
	if statErr != nil {
		t.Fatalf("stat: %v", statErr)
	}
	// update-ca-certificates runs as root but reads the anchor as an
	// ordinary file; 0600 would leave it unreadable to the extract tooling
	// on distributions that drop privileges.
	// The literal, not trustAnchorMode: comparing against the constant would
	// track any change to it instead of pinning the on-disk contract.
	const wantMode os.FileMode = 0o644
	if perm := info.Mode().Perm(); perm != wantMode {
		t.Errorf("anchor mode = %#o, want %#o", perm, wantMode)
	}
}

func TestWriteAnchorFile_RejectsEscapingPaths(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	tests := map[string]string{
		"parent":         filepath.Dir(root),
		"parent child":   filepath.Join(filepath.Dir(root), "stem-root.crt"),
		"traversal":      filepath.Join(root, "..", "stem-root.crt"),
		"deep traversal": filepath.Join(root, "sub", "..", "..", "stem-root.crt"),
	}
	for name, dst := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// The "parent" case names root's own parent directory, which
			// already exists; only a path the guard could have created is
			// evidence that it wrote outside root.
			existedBefore := pathExists(dst)
			if _, err := writeAnchorFile(root, dst, []byte("x")); err == nil {
				t.Fatalf("writeAnchorFile accepted %q, which escapes %q", dst, root)
			} else if !strings.Contains(err.Error(), "escapes") {
				t.Errorf("unexpected error for escaping path: %v", err)
			}
			if !existedBefore && pathExists(dst) {
				t.Errorf("rejected path %q was written anyway", dst)
			}
		})
	}
}

// A child whose name merely starts with ".." is inside root. Guarding with a
// bare [strings.HasPrefix] on ".." would reject it.
func TestWriteAnchorFile_AllowsChildNamedLikeTraversal(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dst := filepath.Join(root, "..stem-root.crt")
	if _, err := writeAnchorFile(root, dst, []byte("x")); err != nil {
		t.Fatalf("writeAnchorFile rejected in-root child %q: %v", dst, err)
	}
	if !pathExists(dst) {
		t.Errorf("%q was accepted but not written", dst)
	}
}

func TestWriteAnchorFile_ReportsWriteFailure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dst := filepath.Join(root, "missing-dir", "stem-root.crt")
	_, err := writeAnchorFile(root, dst, []byte("x"))
	if err == nil {
		t.Fatal("expected an error writing into a directory that does not exist")
	}
	if !strings.Contains(err.Error(), dst) {
		t.Errorf("error does not name the destination: %v", err)
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
