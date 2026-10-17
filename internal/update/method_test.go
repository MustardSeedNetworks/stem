// SPDX-License-Identifier: BUSL-1.1

package update_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/update"
)

// owns answers dpkg -S / rpm -qf as if the named package manager owned the
// binary.
func owns(manager string) update.Runner {
	return func(_ context.Context, name string, _ ...string) ([]byte, error) {
		if name == manager {
			return []byte("stem: /usr/bin/stem\n"), nil
		}
		return nil, errors.New("not owned")
	}
}

// The asset each install takes must be the name the release publishes
// (checked against v0.28.0's asset list).
func TestDetectPicksTheInstallMethodAndItsAsset(t *testing.T) {
	for _, tc := range []struct {
		name      string
		run       update.Runner
		goos, exe string
		goarch    string
		wantName  string
		wantAsset string
	}{
		{"deb", owns("dpkg"), "linux", "/usr/bin/stem", "amd64", "deb", "stem_0.28.0_amd64.deb"},
		{"deb arm64", owns("dpkg"), "linux", "/usr/bin/stem", "arm64", "deb", "stem_0.28.0_arm64.deb"},
		{"rpm", owns("rpm"), "linux", "/usr/bin/stem", "amd64", "rpm", "stem-0.28.0-1.x86_64.rpm"},
		{"rpm arm64", owns("rpm"), "linux", "/usr/bin/stem", "arm64", "rpm", "stem-0.28.0-1.aarch64.rpm"},
		{"macos", owns(""), "darwin", "/usr/local/stem/stem", "arm64", "macos", "stem-0.28.0-darwin-arm64.tar.gz"},
		{"windows", owns(""), "windows", `C:\stem\stem.exe`, "amd64", "windows", "stem-0.28.0-windows-amd64.zip"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, err := update.Detect(t.Context(), tc.run, tc.goos, tc.exe)
			if err != nil {
				t.Fatalf("Detect: %v", err)
			}
			asset, err := m.Asset("0.28.0", tc.goarch)
			if err != nil {
				t.Fatalf("Asset: %v", err)
			}
			if m.Name() != tc.wantName || asset != tc.wantAsset {
				t.Fatalf("Detect = %s %s; want %s %s", m.Name(), asset, tc.wantName, tc.wantAsset)
			}
		})
	}
}

func TestDetectRefusesAnUnownedBinary(t *testing.T) {
	for _, tc := range []struct{ name, goos, exe string }{
		{"linux build tree", "linux", "/home/op/stem/bin/stem"},
		{"macos outside the package prefix", "darwin", "/Users/op/stem"},
		{"unsupported platform", "freebsd", "/usr/local/bin/stem"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := update.Detect(t.Context(), owns(""), tc.goos, tc.exe)
			if !errors.Is(err, update.ErrUnsupportedInstall) {
				t.Fatalf("Detect = %v; want ErrUnsupportedInstall", err)
			}
		})
	}
}

func TestRestartsReadsTheServiceManager(t *testing.T) {
	for _, tc := range []struct {
		name, goos, exe string
		output          string
		want            int
	}{
		{"systemd", "linux", "/usr/bin/stem", "3\n", 3},
		{"launchd", "darwin", "/usr/local/stem/stem", "system/com.stem = {\n\tactive count = 1\n\truns = 7\n}\n", 7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run := func(_ context.Context, name string, _ ...string) ([]byte, error) {
				if name == "dpkg" || name == "systemctl" || name == "launchctl" {
					return []byte(tc.output), nil
				}
				return nil, errors.New("not owned")
			}
			m, err := update.Detect(t.Context(), run, tc.goos, tc.exe)
			if err != nil {
				t.Fatal(err)
			}
			got, err := m.Restarts(t.Context())
			if err != nil || got != tc.want {
				t.Fatalf("Restarts = %d, %v; want %d", got, err, tc.want)
			}
		})
	}
}

// A deliberate start must not be refused by the unit's start-rate limit,
// which an update that rolls back would otherwise exceed.
func TestSystemdStartClearsTheStartLimit(t *testing.T) {
	var calls []string
	run := func(_ context.Context, name string, args ...string) ([]byte, error) {
		calls = append(calls, name+" "+strings.Join(args, " "))
		return nil, nil
	}
	m, err := update.Detect(t.Context(), run, "linux", "/usr/bin/stem")
	if err != nil {
		t.Fatal(err)
	}
	calls = nil
	if err = m.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	want := []string{"systemctl reset-failed stem", "systemctl restart stem"}
	if !slices.Equal(calls, want) {
		t.Fatalf("Start ran %v; want %v", calls, want)
	}
}
