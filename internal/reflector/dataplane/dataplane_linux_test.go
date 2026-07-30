//go:build cgo && linux

// SPDX-License-Identifier: BUSL-1.1

package dataplane_test

import (
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
