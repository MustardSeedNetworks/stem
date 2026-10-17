// SPDX-License-Identifier: BUSL-1.1

package update_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/foundation/pkg/release"

	"github.com/MustardSeedNetworks/stem/internal/update"
)

// fakeSource offers one latest tag; the fake fetcher never opens it.
type fakeSource struct{ latest string }

func (s fakeSource) Latest(context.Context) (string, error) { return s.latest, nil }

func (fakeSource) Open(context.Context, string, string) (io.ReadCloser, error) {
	return nil, errors.New("not used")
}

// fakeFetcher writes the asset as if it had verified, unless the tag is
// listed as failing verification.
type fakeFetcher struct{ refuse map[string]bool }

func (f fakeFetcher) Fetch(_ context.Context, _ release.Source, tag, asset, dir string) (string, error) {
	if f.refuse[tag] {
		return "", release.ErrVerification
	}
	path := filepath.Join(dir, asset)
	return path, os.WriteFile(path, []byte(tag), 0o600)
}

// fakeMethod records what the updater did. Apply and Reapply write the
// artifact's tag into the data directory, standing in for a migration that
// the rollback has to undo.
type fakeMethod struct {
	dataDir  string
	calls    []string
	restarts int
	// crashOn makes the service restart itself once after starting this tag.
	crashOn string
}

func (*fakeMethod) Name() string { return "fake" }

func (*fakeMethod) Asset(version, goarch string) (string, error) {
	return "stem_" + version + "_" + goarch + ".deb", nil
}

func (m *fakeMethod) install(verb, path string) error {
	tag, err := os.ReadFile(path) // #nosec G304 -- test fixture
	if err != nil {
		return err
	}
	m.calls = append(m.calls, verb+" "+string(tag))
	return os.WriteFile(filepath.Join(m.dataDir, "installed"), tag, 0o600)
}

func (m *fakeMethod) Apply(_ context.Context, path string) error { return m.install("apply", path) }

func (m *fakeMethod) Reapply(_ context.Context, path string) error { return m.install("reapply", path) }

func (m *fakeMethod) Stop(context.Context) error {
	m.calls = append(m.calls, "stop")
	return nil
}

func (m *fakeMethod) Start(context.Context) error {
	m.calls = append(m.calls, "start")
	return nil
}

func (m *fakeMethod) Restarts(context.Context) (int, error) {
	n := m.restarts
	if m.crashOn != "" && strings.HasSuffix(m.calls[len(m.calls)-2], m.crashOn) {
		m.restarts++
	}
	return n, nil
}

type fixture struct {
	updater *update.Updater
	method  *fakeMethod
	dataDir string
	healthy func(version string) bool
}

func newFixture(t *testing.T, running, latest string) *fixture {
	t.Helper()
	dataDir := filepath.Join(t.TempDir(), "stem")
	if err := os.Mkdir(dataDir, 0o750); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"installed": "v" + running, "stem.db": "results"} {
		if err := os.WriteFile(filepath.Join(dataDir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink("stem.db", filepath.Join(dataDir, "current.db")); err != nil {
		t.Fatal(err)
	}
	f := &fixture{
		method:  &fakeMethod{dataDir: dataDir},
		dataDir: dataDir,
		healthy: func(string) bool { return true },
	}
	f.updater = &update.Updater{
		Verifier: fakeFetcher{},
		Source:   fakeSource{latest: latest},
		Method:   f.method,
		DataDir:  dataDir,
		Running:  running,
		GOARCH:   "amd64",
		Health: func(_ context.Context, version string) error {
			installed, err := os.ReadFile(filepath.Join(dataDir, "installed")) // #nosec G304 -- test fixture
			if err != nil {
				return err
			}
			if string(installed) != "v"+version || !f.healthy(version) {
				return errors.New("not healthy")
			}
			return nil
		},
		HealthTimeout: 50 * time.Millisecond,
		Log:           slog.New(slog.DiscardHandler),
	}
	return f
}

func (f *fixture) read(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(f.dataDir, name)) // #nosec G304 -- test fixture
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// leftovers lists what the update left beside the data directory.
func (f *fixture) leftovers(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Dir(f.dataDir))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		if e.Name() != "stem" {
			names = append(names, e.Name())
		}
	}
	return names
}

func (f *fixture) assertUntouched(t *testing.T) {
	t.Helper()
	if len(f.method.calls) != 0 {
		t.Fatalf("calls = %v; want none", f.method.calls)
	}
	if left := f.leftovers(t); len(left) != 0 {
		t.Fatalf("left behind %v", left)
	}
}

func TestUpdateInstallsTheNewerRelease(t *testing.T) {
	f := newFixture(t, "0.28.0", "v0.29.0")

	got, err := f.updater.Run(t.Context())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got != "0.29.0" {
		t.Fatalf("Run = %q; want 0.29.0", got)
	}
	if want := []string{"stop", "apply v0.29.0", "start"}; !slices.Equal(f.method.calls, want) {
		t.Fatalf("calls = %v; want %v", f.method.calls, want)
	}
	if left := f.leftovers(t); len(left) != 0 {
		t.Fatalf("left behind %v; want the staging directory removed", left)
	}
}

// Any failed check after the install restores the state copy and the
// previous version, and the error names the check.
func TestUpdateRollsBack(t *testing.T) {
	for _, tc := range []struct {
		name     string
		sabotage func(*fixture)
		wantErr  string
	}{
		{
			name:     "never healthy",
			sabotage: func(f *fixture) { f.healthy = func(v string) bool { return v != "0.29.0" } },
			wantErr:  "did not come up",
		},
		{
			name:     "restarted by the service manager",
			sabotage: func(f *fixture) { f.method.crashOn = "v0.29.0" },
			wantErr:  "the service restarted 1 times",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, "0.28.0", "v0.29.0")
			tc.sabotage(f)

			_, err := f.updater.Run(t.Context())
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) ||
				!strings.Contains(err.Error(), "0.28.0 was restored") {
				t.Fatalf("Run = %v; want a restored rollback naming %q", err, tc.wantErr)
			}
			want := []string{"stop", "apply v0.29.0", "start", "stop", "reapply v0.28.0", "start"}
			if !slices.Equal(f.method.calls, want) {
				t.Fatalf("calls = %v; want %v", f.method.calls, want)
			}
			if got := f.read(t, "installed"); got != "v0.28.0" {
				t.Fatalf("installed after rollback = %q; want v0.28.0", got)
			}
			if got := f.read(t, "current.db"); got != "results" {
				t.Fatalf("symlinked state after rollback = %q; want the original", got)
			}
			if left := f.leftovers(t); len(left) != 0 {
				t.Fatalf("left behind %v; want the staging directory removed", left)
			}
		})
	}
}

// The state copy undoes a change the new version made, not just the binary.
func TestRollbackRestoresTheStateCopy(t *testing.T) {
	f := newFixture(t, "0.28.0", "v0.29.0")
	health := f.updater.Health
	f.updater.Health = func(ctx context.Context, version string) error {
		if version == "0.29.0" {
			_ = os.WriteFile(filepath.Join(f.dataDir, "stem.db"), []byte("migrated"), 0o600)
			_ = os.WriteFile(filepath.Join(f.dataDir, "new-table"), nil, 0o600)
			return errors.New("not healthy")
		}
		return health(ctx, version)
	}

	if _, err := f.updater.Run(t.Context()); err == nil {
		t.Fatal("Run succeeded; want the health check to fail")
	}
	if got := f.read(t, "stem.db"); got != "results" {
		t.Fatalf("stem.db after rollback = %q; want the pre-update copy", got)
	}
	if _, err := os.Stat(filepath.Join(f.dataDir, "new-table")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("a file the new version created survived the rollback: %v", err)
	}
	if _, err := os.Stat(filepath.Join(f.dataDir, update.RunLockName)); err != nil {
		t.Fatalf("the run lock the update held is gone: %v", err)
	}
}

// Nothing is stopped, copied or installed when the update cannot start.
func TestUpdateRefusesBeforeTouchingAnything(t *testing.T) {
	for _, tc := range []struct {
		name    string
		running string
		latest  string
		prepare func(*testing.T, *fixture)
		wantIs  error
	}{
		{name: "same version", running: "0.28.0", latest: "v0.28.0", wantIs: update.ErrNotNewer},
		{name: "older version", running: "0.28.0", latest: "v0.27.15", wantIs: update.ErrNotNewer},
		{name: "minor beats patch", running: "0.28.10", latest: "v0.28.9", wantIs: update.ErrNotNewer},
		{name: "development build", running: "dev", latest: "v0.29.0", wantIs: update.ErrUnsupportedInstall},
		{
			name: "target refused", running: "0.28.0", latest: "v0.29.0",
			prepare: func(_ *testing.T, f *fixture) {
				f.updater.Verifier = fakeFetcher{refuse: map[string]bool{"v0.29.0": true}}
			},
			wantIs: release.ErrVerification,
		},
		{
			name: "no verified rollback", running: "0.28.0", latest: "v0.29.0",
			prepare: func(_ *testing.T, f *fixture) {
				f.updater.Verifier = fakeFetcher{refuse: map[string]bool{"v0.28.0": true}}
			},
			wantIs: release.ErrVerification,
		},
		{
			name: "a run is active", running: "0.28.0", latest: "v0.29.0",
			prepare: func(t *testing.T, f *fixture) {
				unlock, err := update.NewRunGuard(f.dataDir).Acquire()
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(unlock)
			},
			wantIs: update.ErrRunActive,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, tc.running, tc.latest)
			if tc.prepare != nil {
				tc.prepare(t, f)
			}

			_, err := f.updater.Run(t.Context())
			if !errors.Is(err, tc.wantIs) {
				t.Fatalf("Run = %v; want %v", err, tc.wantIs)
			}
			f.assertUntouched(t)
		})
	}
}

// An interrupt after the service stopped must not abandon the install
// halfway: the update still finishes.
func TestUpdateIgnoresCancellationOnceStarted(t *testing.T) {
	f := newFixture(t, "0.28.0", "v0.29.0")
	ctx, cancel := context.WithCancel(t.Context())
	health := f.updater.Health
	f.updater.Health = func(hctx context.Context, version string) error {
		cancel()
		if err := hctx.Err(); err != nil {
			return err
		}
		return health(hctx, version)
	}

	if got, err := f.updater.Run(ctx); err != nil || got != "0.29.0" {
		t.Fatalf("Run = %q, %v; want the update to finish", got, err)
	}
}
