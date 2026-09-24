//go:build cgo && linux

// SPDX-License-Identifier: BUSL-1.1

package dataplane_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/reflector/config"
	"github.com/MustardSeedNetworks/stem/internal/reflector/dataplane"
)

func newDataplane(t *testing.T) *dataplane.Dataplane {
	t.Helper()
	cfg := &config.Config{
		Interface:       "lo",
		SignatureFilter: "all",
		Filtering:       config.FilterConfig{Port: 3842, OUI: "00:c0:17"},
		Reflection:      config.ReflectConfig{Mode: "all"},
	}
	dp, err := dataplane.New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(dp.Close)
	return dp
}

func TestCloseSerializesConcurrentStatsReads(t *testing.T) {
	dp := &dataplane.Dataplane{}
	var readers sync.WaitGroup
	for range 16 {
		readers.Go(func() {
			for range 100 {
				_ = dp.GetStats()
			}
		})
	}
	dp.Close()
	readers.Wait()
	dp.Close()
	dp.ResetStats()

	if stats := dp.GetStats(); stats != (dataplane.Stats{}) {
		t.Fatalf("GetStats() after close = %+v, want zero stats", stats)
	}
	if err := dp.Start(); err == nil {
		t.Fatal("Start() after close error = nil")
	}
}

func TestUpdateConfigRejectsBeforeMutation(t *testing.T) {
	dp := newDataplane(t)
	want := *dp.Config()
	port, invalid := uint16(9999), "invalid"

	err := dp.UpdateConfig(&dataplane.ConfigUpdate{Port: &port, SignatureFilter: &invalid})
	if err == nil {
		t.Fatal("UpdateConfig() error = nil; want invalid signature filter error")
	}
	if got := *dp.Config(); got != want {
		t.Fatalf("Config() = %+v, want unchanged %+v", got, want)
	}
}

func TestUpdateConfigRejectsInvalidModeBeforeMutation(t *testing.T) {
	dp := newDataplane(t)
	want := *dp.Config()
	port, invalid := uint16(9999), "invalid"

	err := dp.UpdateConfig(&dataplane.ConfigUpdate{Port: &port, Mode: &invalid})
	if err == nil {
		t.Fatal("UpdateConfig() error = nil; want invalid mode error")
	}
	if got := *dp.Config(); got != want {
		t.Fatalf("Config() = %+v, want unchanged %+v", got, want)
	}
}

func TestResetStatsAfterStopIsSafe(t *testing.T) {
	dp := &dataplane.Dataplane{}
	dp.Stop()
	dp.ResetStats()

	if stats := dp.GetStats(); stats != (dataplane.Stats{}) {
		t.Fatalf("GetStats() after reset = %+v, want zero stats", stats)
	}
}

// uidLine returns the real, effective, saved and filesystem UIDs a
// /proc/.../status file reports, or "" when the task has exited.
func uidLine(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for line := range strings.Lines(string(data)) {
		if uid, ok := strings.CutPrefix(line, "Uid:"); ok {
			return strings.Join(strings.Fields(uid), " ")
		}
	}
	return ""
}

// The reflector runs inside the daemon, so starting it must not change the
// daemon's credentials: a setuid() there reaches every thread, and the next
// reflector start or test run then fails EPERM (stem#1231). Needs root, as the
// daemon has in a container; the shipped unit gets CAP_NET_RAW instead.
func TestStartKeepsTheDaemonsCredentials(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("needs root to open an AF_PACKET socket")
	}
	want := uidLine("/proc/self/status")

	dp := newDataplane(t)
	for run := range 2 {
		if err := dp.Start(); err != nil {
			t.Fatalf("start %d: %v", run+1, err)
		}
		dp.Stop()
	}

	threads, err := filepath.Glob("/proc/self/task/*/status")
	if err != nil || len(threads) == 0 {
		t.Fatalf("list threads: %v (%d found)", err, len(threads))
	}
	for _, thread := range threads {
		if uid := uidLine(thread); uid != "" && uid != want {
			t.Errorf("%s: Uid %q after two reflector runs, want %q", thread, uid, want)
		}
	}
}

// Without CAP_NET_RAW the start fails, and the error says why.
func TestStartWithoutRawSocketAccessReportsPermission(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can open an AF_PACKET socket")
	}
	err := newDataplane(t).Start()
	if !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("Start() error = %v, want one matching fs.ErrPermission", err)
	}
}
