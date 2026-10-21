// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/auth"
)

// firstRunPassword clears the layered password policy (length, zxcvbn).
const firstRunPassword = "kettle-orbit-granite-velvet-42"

// unconfiguredEnv is a fresh install (#1282): no STEM_AUTH_* in the
// environment, nothing in the data directory.
func unconfiguredEnv(t *testing.T) string {
	t.Helper()
	dataDir := t.TempDir()
	t.Setenv("STEM_DISABLE_HIBP", "1")
	t.Setenv("STEM_DATA_DIR", dataDir)
	t.Setenv("STEM_AUTH_USERNAME", "")
	t.Setenv("STEM_AUTH_PASSWORD", "")
	return dataDir
}

func newFirstRunServer(t *testing.T) *api.Server {
	t.Helper()
	s, err := api.NewServer(api.ListenAddr{Port: 8444})
	if err != nil {
		t.Fatalf("NewServer() on a fresh install: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

func serve(s *api.Server, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	return w
}

func loginStatus(s *api.Server, username, password string) int {
	return serve(s, http.MethodPost, "/api/v1/auth/login",
		fmt.Sprintf(`{"username":%q,"password":%q}`, username, password)).Code
}

func setupStatus(t *testing.T, s *api.Server) api.SetupStatusResponse {
	t.Helper()
	w := serve(s, http.MethodGet, "/api/v1/setup/status", "")
	if w.Code != http.StatusOK {
		t.Fatalf("setup status = %d: %s", w.Code, w.Body.String())
	}
	var resp api.SetupStatusResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode setup status: %v", err)
	}
	return resp
}

// readSetupToken reads the token the way an operator does: from the file
// setup/status names.
func readSetupToken(t *testing.T, s *api.Server) string {
	t.Helper()
	path := setupStatus(t, s).SetupTokenFile
	if path == "" {
		t.Fatal("setup status names no setup token file")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read setup token: %v", err)
	}
	return strings.TrimSpace(string(data))
}

func completeSetup(t *testing.T, s *api.Server) {
	t.Helper()
	body, err := json.Marshal(api.SetupCompleteRequest{Password: firstRunPassword, SetupToken: readSetupToken(t, s)})
	if err != nil {
		t.Fatalf("encode setup request: %v", err)
	}
	w := serve(s, http.MethodPost, "/api/v1/setup/complete", string(body))
	if w.Code != http.StatusOK {
		t.Fatalf("setup complete = %d: %s", w.Code, w.Body.String())
	}
}

// TestFirstRunSetupSurvivesRestart is #1282's contract: a daemon with no
// credentials anywhere starts, serves the setup page, refuses every login
// until an operator claims it, and the claimed credential is still the one
// that works after a restart.
func TestFirstRunSetupSurvivesRestart(t *testing.T) {
	dataDir := unconfiguredEnv(t)
	first := newFirstRunServer(t)

	status := setupStatus(t, first)
	if !status.NeedsSetup {
		t.Fatal("fresh install: needsSetup = false, want true")
	}
	if status.Username != auth.FirstRunUsername {
		t.Errorf("fresh install: username = %q, want %q", status.Username, auth.FirstRunUsername)
	}
	if code := loginStatus(first, auth.FirstRunUsername, firstRunPassword); code != http.StatusUnauthorized {
		t.Fatalf("login before setup = %d, want 401: an unclaimed daemon has no usable credential", code)
	}

	completeSetup(t, first)

	info, err := os.Stat(filepath.Join(dataDir, "credentials.json"))
	if err != nil {
		t.Fatalf("credential store after setup: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("credential store mode = %o, want 600", mode)
	}
	if code := loginStatus(first, auth.FirstRunUsername, firstRunPassword); code != http.StatusOK {
		t.Fatalf("login after setup = %d, want 200", code)
	}
	if _, statErr := os.Stat(filepath.Join(dataDir, ".setup-token")); !os.IsNotExist(statErr) {
		t.Errorf("setup token file after setup: stat err = %v, want not-exist", statErr)
	}
	_ = first.Shutdown()

	restarted := newFirstRunServer(t)
	if setupStatus(t, restarted).NeedsSetup {
		t.Fatal("after restart: needsSetup = true, the completed setup was lost")
	}
	if code := loginStatus(restarted, auth.FirstRunUsername, firstRunPassword); code != http.StatusOK {
		t.Fatalf("login after restart = %d, want 200", code)
	}
	wrong := loginStatus(restarted, auth.FirstRunUsername, "wrong-"+firstRunPassword)
	if wrong != http.StatusUnauthorized {
		t.Fatalf("wrong password after restart = %d, want 401", wrong)
	}
	body, err := json.Marshal(api.SetupCompleteRequest{Password: "a-second-claim-" + firstRunPassword})
	if err != nil {
		t.Fatalf("encode setup request: %v", err)
	}
	if w := serve(restarted, http.MethodPost, "/api/v1/setup/complete", string(body)); w.Code != http.StatusForbidden {
		t.Fatalf("second setup after restart = %d, want 403", w.Code)
	}
}

// TestSetupTokenNeverServed: setup/status answers any network peer, so it
// names the token file and never carries the token. Only a caller that can
// read the file completes setup.
func TestSetupTokenNeverServed(t *testing.T) {
	dataDir := unconfiguredEnv(t)
	s := newFirstRunServer(t)

	tokenFile := filepath.Join(dataDir, ".setup-token")
	info, err := os.Stat(tokenFile)
	if err != nil {
		t.Fatalf("a fresh install wrote no setup token file: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("setup token file mode = %o, want 600", mode)
	}
	token := readSetupToken(t, s)

	for range 2 {
		w := serve(s, http.MethodGet, "/api/v1/setup/status", "")
		if strings.Contains(w.Body.String(), token) {
			t.Fatalf("setup status leaks the setup token: %s", w.Body.String())
		}
		var fields map[string]any
		if decodeErr := json.Unmarshal(w.Body.Bytes(), &fields); decodeErr != nil {
			t.Fatalf("decode setup status: %v", decodeErr)
		}
		if _, ok := fields["setupToken"]; ok {
			t.Fatalf("setup status has a setupToken field: %s", w.Body.String())
		}
		if fields["setupTokenFile"] != tokenFile {
			t.Errorf("setupTokenFile = %v, want %s", fields["setupTokenFile"], tokenFile)
		}
	}

	for _, guess := range []string{"", "not-the-token"} {
		guessBody, encodeErr := json.Marshal(api.SetupCompleteRequest{Password: firstRunPassword, SetupToken: guess})
		if encodeErr != nil {
			t.Fatal(encodeErr)
		}
		if w := serve(s, http.MethodPost, "/api/v1/setup/complete", string(guessBody)); w.Code != http.StatusForbidden {
			t.Fatalf("setup with token %q = %d, want 403", guess, w.Code)
		}
	}
	if !setupStatus(t, s).NeedsSetup {
		t.Fatal("a refused setup claimed the daemon")
	}

	// A password the policy rejects does not burn the token.
	body, err := json.Marshal(api.SetupCompleteRequest{Password: "short", SetupToken: token})
	if err != nil {
		t.Fatal(err)
	}
	if w := serve(s, http.MethodPost, "/api/v1/setup/complete", string(body)); w.Code == http.StatusOK {
		t.Fatal("setup accepted a password the policy rejects")
	}
	completeSetup(t, s)
}

// TestRecoveredPasswordSurvivesRestart: a password reset through the
// recovery flow is written to the same store, so a restart does not revert
// to the forgotten one.
func TestRecoveredPasswordSurvivesRestart(t *testing.T) {
	dataDir := unconfiguredEnv(t)
	first := newFirstRunServer(t)
	completeSetup(t, first)

	if err := os.WriteFile(filepath.Join(dataDir, ".recovery"), nil, 0o600); err != nil {
		t.Fatalf("create recovery trigger: %v", err)
	}
	if w := serve(first, http.MethodGet, "/api/v1/recovery/status", ""); w.Code != http.StatusOK {
		t.Fatalf("recovery status = %d: %s", w.Code, w.Body.String())
	}
	token, err := os.ReadFile(filepath.Join(dataDir, ".recovery-token"))
	if err != nil {
		t.Fatalf("read recovery token: %v", err)
	}
	const recovered = "lantern-quartz-meadow-falcon-77"
	body, err := json.Marshal(api.RecoveryCompleteRequest{Token: strings.TrimSpace(string(token)), Password: recovered})
	if err != nil {
		t.Fatalf("encode recovery request: %v", err)
	}
	if w := serve(first, http.MethodPost, "/api/v1/recovery/complete", string(body)); w.Code != http.StatusOK {
		t.Fatalf("recovery complete = %d: %s", w.Code, w.Body.String())
	}
	_ = first.Shutdown()

	restarted := newFirstRunServer(t)
	if code := loginStatus(restarted, auth.FirstRunUsername, recovered); code != http.StatusOK {
		t.Fatalf("recovered password after restart = %d, want 200", code)
	}
	if code := loginStatus(restarted, auth.FirstRunUsername, firstRunPassword); code != http.StatusUnauthorized {
		t.Fatalf("forgotten password after restart = %d, want 401", code)
	}
}

// TestCredentialSourceRefusals covers the starts that must fail rather than
// fall back to an unclaimed daemon anyone on the network could take.
func TestCredentialSourceRefusals(t *testing.T) {
	t.Run("damaged store", func(t *testing.T) {
		dataDir := unconfiguredEnv(t)
		if err := os.WriteFile(filepath.Join(dataDir, "credentials.json"), []byte("{not json"), 0o600); err != nil {
			t.Fatalf("write damaged store: %v", err)
		}
		if _, err := api.NewServer(api.ListenAddr{Port: 8444}); err == nil {
			t.Fatal("NewServer() over a damaged credential store succeeded; want a refusal, not first-run setup")
		}
	})

	t.Run("store with no password", func(t *testing.T) {
		dataDir := unconfiguredEnv(t)
		if err := os.WriteFile(filepath.Join(dataDir, "credentials.json"),
			[]byte(`{"username":"admin","passwordHash":""}`), 0o600); err != nil {
			t.Fatalf("write empty store: %v", err)
		}
		if _, err := api.NewServer(api.ListenAddr{Port: 8444}); err == nil {
			t.Fatal("NewServer() over a store with no password succeeded; want a refusal, not first-run setup")
		}
	})

	t.Run("half-set environment", func(t *testing.T) {
		unconfiguredEnv(t)
		t.Setenv("STEM_AUTH_USERNAME", "operator")
		_, err := api.NewServer(api.ListenAddr{Port: 8444})
		if !errors.Is(err, auth.ErrMissingCredentials) {
			t.Fatalf("NewServer() with only STEM_AUTH_USERNAME = %v, want ErrMissingCredentials", err)
		}
	})
}

// TestStoredCredentialWinsOverEnvironment: once an operator has claimed or
// reset the credential in the store, a stale environment cannot silently
// revert it.
func TestStoredCredentialWinsOverEnvironment(t *testing.T) {
	unconfiguredEnv(t)
	first := newFirstRunServer(t)
	completeSetup(t, first)
	_ = first.Shutdown()

	t.Setenv("STEM_AUTH_USERNAME", "envuser")
	t.Setenv("STEM_AUTH_PASSWORD", "envpass-"+firstRunPassword)
	restarted := newFirstRunServer(t)
	if code := loginStatus(restarted, auth.FirstRunUsername, firstRunPassword); code != http.StatusOK {
		t.Fatalf("stored credential with env also set = %d, want 200", code)
	}
	if code := loginStatus(restarted, "envuser", "envpass-"+firstRunPassword); code != http.StatusUnauthorized {
		t.Fatalf("env credential with a store present = %d, want 401", code)
	}
}
