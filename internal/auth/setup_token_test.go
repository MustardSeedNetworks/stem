// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/auth"
)

func readSetupToken(t *testing.T, m *auth.SetupTokenManager) string {
	t.Helper()
	data, err := os.ReadFile(m.FilePath())
	if err != nil {
		t.Fatalf("read setup token file: %v", err)
	}
	return strings.TrimSpace(string(data))
}

func TestSetupTokenIssueWritesOwnerOnlyFile(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "data")
	m := auth.NewSetupTokenManager(dir)

	if m.FilePath() != filepath.Join(dir, ".setup-token") {
		t.Fatalf("FilePath() = %q, want .setup-token in the data directory", m.FilePath())
	}
	// A pre-existing world-readable file must not keep its mode.
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(m.FilePath(), []byte("stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := m.Issue(); err != nil {
		t.Fatalf("Issue() = %v", err)
	}
	info, err := os.Stat(m.FilePath())
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("token file mode = %o, want 600", mode)
	}
	token := readSetupToken(t, m)
	if token == "stale" || len(token) < 40 {
		t.Errorf("token = %q, want a fresh 32-byte token", token)
	}
}

func TestSetupTokenFilePathIsAbsolute(t *testing.T) {
	t.Chdir(t.TempDir())
	m := auth.NewSetupTokenManager(".")
	if !filepath.IsAbs(m.FilePath()) {
		t.Errorf("FilePath() = %q for data directory \".\", want an absolute path", m.FilePath())
	}
}

func TestSetupTokenValidate(t *testing.T) {
	t.Parallel()
	m := auth.NewSetupTokenManager(t.TempDir())

	if m.Validate("anything") {
		t.Error("Validate() accepted a token before one was issued")
	}
	if err := m.Issue(); err != nil {
		t.Fatal(err)
	}
	token := readSetupToken(t, m)

	tests := []struct {
		name     string
		provided string
		want     bool
	}{
		{"empty", "", false},
		{"wrong", "wrong-token", false},
		{"prefix", token[:len(token)-1], false},
		{"issued", token, true},
		{"issued again: a rejected password can be retried", token, true},
	}
	for _, tt := range tests {
		if got := m.Validate(tt.provided); got != tt.want {
			t.Errorf("%s: Validate() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestSetupTokenReissueReplacesOldToken(t *testing.T) {
	t.Parallel()
	m := auth.NewSetupTokenManager(t.TempDir())
	if err := m.Issue(); err != nil {
		t.Fatal(err)
	}
	first := readSetupToken(t, m)
	if err := m.Issue(); err != nil {
		t.Fatal(err)
	}
	second := readSetupToken(t, m)

	if first == second {
		t.Fatal("two issues produced the same token")
	}
	if m.Validate(first) {
		t.Error("the replaced token is still accepted")
	}
	if !m.Validate(second) {
		t.Error("the current token is refused")
	}
}

func TestSetupTokenConsume(t *testing.T) {
	t.Parallel()
	m := auth.NewSetupTokenManager(t.TempDir())
	if err := m.Issue(); err != nil {
		t.Fatal(err)
	}
	token := readSetupToken(t, m)

	if err := m.Consume(); err != nil {
		t.Fatalf("Consume() = %v", err)
	}
	if m.Validate(token) {
		t.Error("a consumed token is still accepted")
	}
	if _, err := os.Stat(m.FilePath()); !os.IsNotExist(err) {
		t.Errorf("token file after Consume: stat err = %v, want not-exist", err)
	}
	if err := m.Consume(); err != nil {
		t.Errorf("second Consume() = %v, want nil", err)
	}
}

func TestSetupTokenConcurrentAccess(t *testing.T) {
	t.Parallel()
	m := auth.NewSetupTokenManager(t.TempDir())
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() { _ = m.Issue() })
		wg.Go(func() { _ = m.Validate("sometoken") })
	}
	wg.Wait()
}
