//go:build cgo && linux

// SPDX-License-Identifier: BUSL-1.1

package dataplane

import (
	"sync"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/reflector/config"
)

func TestSignatureFilterValueAcceptsITOSignatures(t *testing.T) {
	for filter, want := range map[string]int{"probeot": 6, "dataot": 7, "latency": 8} {
		value, err := signatureFilterValue(filter)
		if err != nil || value != want {
			t.Fatalf("signatureFilterValue(%q) = %d, %v; want %d, nil", filter, value, err, want)
		}
	}
}

func TestCloseSerializesConcurrentStatsReads(t *testing.T) {
	dp := &Dataplane{}
	var readers sync.WaitGroup
	for range 16 {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for range 100 {
				_ = dp.GetStats()
			}
		}()
	}
	dp.Close()
	readers.Wait()
	dp.Close()
	dp.ResetStats()

	if stats := dp.GetStats(); stats != (Stats{}) {
		t.Fatalf("GetStats() after close = %+v, want zero stats", stats)
	}
	if err := dp.Start(); err == nil {
		t.Fatal("Start() after close error = nil")
	}
}

func TestUpdateConfigRejectsBeforeMutation(t *testing.T) {
	cfg := &config.Config{SignatureFilter: "all", Filtering: config.FilterConfig{Port: 3842}}
	dp := &Dataplane{cfg: cfg}
	port, invalid := uint16(9999), "invalid"

	err := dp.UpdateConfig(&ConfigUpdate{Port: &port, SignatureFilter: &invalid})
	if err == nil {
		t.Fatal("UpdateConfig() error = nil; want invalid signature filter error")
	}
	if cfg.Filtering.Port != 3842 {
		t.Fatalf("port changed to %d after rejected update", cfg.Filtering.Port)
	}
}

func TestUpdateConfigRejectsInvalidModeBeforeMutation(t *testing.T) {
	cfg := &config.Config{
		Filtering:  config.FilterConfig{Port: 3842},
		Reflection: config.ReflectConfig{Mode: "all"},
	}
	dp := &Dataplane{cfg: cfg}
	port, invalid := uint16(9999), "invalid"

	err := dp.UpdateConfig(&ConfigUpdate{Port: &port, Mode: &invalid})
	if err == nil {
		t.Fatal("UpdateConfig() error = nil; want invalid mode error")
	}
	if cfg.Filtering.Port != 3842 || cfg.Reflection.Mode != "all" {
		t.Fatalf("config changed after rejected update: %+v", cfg)
	}
}

func TestResetStatsAfterStopIsSafe(t *testing.T) {
	dp := &Dataplane{}
	dp.Stop()
	dp.ResetStats()

	if stats := dp.GetStats(); stats != (Stats{}) {
		t.Fatalf("GetStats() after reset = %+v, want zero stats", stats)
	}
}
