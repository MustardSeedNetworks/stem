// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/auth"
)

func TestWriteCLITokenRoundTrips(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := auth.WriteCLIToken(dir, "sentinel-token"); err != nil {
		t.Fatalf("WriteCLIToken: %v", err)
	}

	got, err := auth.ReadCLIToken(dir)
	if err != nil {
		t.Fatalf("ReadCLIToken: %v", err)
	}
	if got != "sentinel-token" {
		t.Errorf("token = %q, want %q", got, "sentinel-token")
	}
}

// The token is a bearer credential at rest: anything the daemon's own user
// cannot read must not be able to read it either.
func TestWriteCLITokenIsOwnerOnly(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes are not enforced on Windows")
	}

	dir := t.TempDir()
	if err := auth.WriteCLIToken(dir, "sentinel-token"); err != nil {
		t.Fatalf("WriteCLIToken: %v", err)
	}

	info, err := os.Stat(auth.CLITokenPath(dir))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("mode = %#o, want %#o", perm, 0o600)
	}
}

// A file an operator has loosened is not a credential this code will use:
// reading it would hand every local account the daemon's API.
func TestReadCLITokenRejectsGroupOrWorldReadable(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes are not enforced on Windows")
	}

	for name, mode := range map[string]os.FileMode{
		"group readable": 0o640,
		"world readable": 0o604,
		"world writable": 0o622,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			if err := auth.WriteCLIToken(dir, "sentinel-token"); err != nil {
				t.Fatalf("WriteCLIToken: %v", err)
			}
			if err := os.Chmod(auth.CLITokenPath(dir), mode); err != nil {
				t.Fatalf("chmod: %v", err)
			}

			_, err := auth.ReadCLIToken(dir)
			if !errors.Is(err, auth.ErrCLITokenPermissions) {
				t.Errorf("err = %v, want ErrCLITokenPermissions", err)
			}
		})
	}
}

func TestReadCLITokenMissingFile(t *testing.T) {
	t.Parallel()

	_, err := auth.ReadCLIToken(t.TempDir())
	if !errors.Is(err, auth.ErrCLITokenNotFound) {
		t.Errorf("err = %v, want ErrCLITokenNotFound", err)
	}
}

func TestReadCLITokenRejectsEmptyFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(auth.CLITokenPath(dir), []byte("  \n"), 0o600); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	_, err := auth.ReadCLIToken(dir)
	if !errors.Is(err, auth.ErrCLITokenNotFound) {
		t.Errorf("err = %v, want ErrCLITokenNotFound", err)
	}
}

func TestWriteCLITokenRejectsEmptyToken(t *testing.T) {
	t.Parallel()

	if err := auth.WriteCLIToken(t.TempDir(), " "); err == nil {
		t.Error("WriteCLIToken(empty) = nil, want error")
	}
}

// Rewriting must not leave the reader a torn file or a stray temp file
// holding the previous credential.
func TestWriteCLITokenReplacesAtomically(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	for _, token := range []string{"first", "second", "third"} {
		if err := auth.WriteCLIToken(dir, token); err != nil {
			t.Fatalf("WriteCLIToken(%q): %v", token, err)
		}
	}

	got, err := auth.ReadCLIToken(dir)
	if err != nil {
		t.Fatalf("ReadCLIToken: %v", err)
	}
	if got != "third" {
		t.Errorf("token = %q, want %q", got, "third")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(entries) != 1 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("directory holds %v, want only the token file", names)
	}
}

// A token left behind by a stopped daemon is a credential nothing can revoke,
// and the CLI would present it to whatever answers next.
func TestRemoveCLITokenIsIdempotent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := auth.WriteCLIToken(dir, "sentinel-token"); err != nil {
		t.Fatalf("WriteCLIToken: %v", err)
	}
	for range 2 {
		if err := auth.RemoveCLIToken(dir); err != nil {
			t.Fatalf("RemoveCLIToken: %v", err)
		}
	}
	if _, err := os.Stat(auth.CLITokenPath(dir)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("stat err = %v, want ErrNotExist", err)
	}
}

func TestCLITokenPathIsUnderDataDir(t *testing.T) {
	t.Parallel()

	got := auth.CLITokenPath("/var/lib/stem")
	if want := filepath.Join("/var/lib/stem", ".cli-token"); got != want {
		t.Errorf("CLITokenPath = %q, want %q", got, want)
	}
}

// The CLI presents this token on the same Bearer path a browser session uses,
// so the one validator must accept it and name it as a CLI credential.
func TestGenerateCLITokenValidates(t *testing.T) {
	t.Parallel()

	mgr, err := auth.NewManager("test-secret-value-at-least-32-chars-long", 0, "operator", "Correct-Horse-Battery-9")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	t.Cleanup(mgr.Stop)

	token, err := mgr.GenerateCLIToken()
	if err != nil {
		t.Fatalf("GenerateCLIToken: %v", err)
	}

	claims, err := mgr.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.Username != "operator" {
		t.Errorf("username = %q, want %q", claims.Username, "operator")
	}
	if claims.TokenType != auth.TokenTypeCLI {
		t.Errorf("token type = %q, want %q", claims.TokenType, auth.TokenTypeCLI)
	}
}

// A leaked file must stop working on its own; an unexpiring bearer token on
// disk is a permanent grant.
func TestGenerateCLITokenExpires(t *testing.T) {
	t.Parallel()

	mgr, err := auth.NewManager("test-secret-value-at-least-32-chars-long", 0, "operator", "Correct-Horse-Battery-9")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	t.Cleanup(mgr.Stop)

	token, err := mgr.GenerateCLIToken()
	if err != nil {
		t.Fatalf("GenerateCLIToken: %v", err)
	}
	claims, err := mgr.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.ExpiresAt == nil {
		t.Fatal("CLI token has no expiry")
	}
	// Pinned to the literal, not to CLITokenLifetime: comparing the constant
	// against itself would accept any value the constant is changed to.
	if got := claims.ExpiresAt.Sub(claims.IssuedAt.Time); got != 24*time.Hour {
		t.Errorf("lifetime = %v, want %v", got, 24*time.Hour)
	}
}
