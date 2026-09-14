// SPDX-License-Identifier: BUSL-1.1

package benchmark_test

import (
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/services/benchmark"
	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
)

// TestBuildDataplaneConfigCarriesThroughputSearchBounds pins the boundary that
// #1217 went through: the throughput binary search is bounded by
// initial_rate_pct and max_iterations, and Configure overwrites the C defaults
// with whatever Go sends. Sending zeros made the search exit before its first
// trial, so a five-second run finished in a millisecond having transmitted
// nothing and still reported success. No test could see this, because the stub
// Configure discards its argument and the cgo one does not build on macOS.
func TestBuildDataplaneConfigCarriesThroughputSearchBounds(t *testing.T) {
	tests := []struct {
		name           string
		params         map[string]any
		wantResolution float64
		wantLoss       float64
	}{
		{
			name:           "no params uses defaults",
			params:         nil,
			wantResolution: benchmark.DefaultResolutionForTest(),
			wantLoss:       benchmark.DefaultAcceptableLossForTest(),
		},
		{
			name:           "operator parameters are carried",
			params:         map[string]any{"resolution": 0.5, "max_loss": 0.1},
			wantResolution: 0.5,
			wantLoss:       0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dpCfg := benchmark.BuildDataplaneConfigForTest(&modtypes.TestConfig{
				Interface: "lo",
				Peer:      "192.0.2.1",
				PeerPort:  3842,
				FrameSize: 64,
				Duration:  5,
				Params:    tt.params,
			})

			// The two that were dropped. Zero for either means run_trial is
			// never called.
			if dpCfg.InitialRatePct != benchmark.DefaultInitialRatePctForTest() {
				t.Errorf("InitialRatePct = %v, want %v (zero stops the search on entry)",
					dpCfg.InitialRatePct, benchmark.DefaultInitialRatePctForTest())
			}
			if dpCfg.MaxIterations != benchmark.DefaultMaxIterationsForTest() {
				t.Errorf("MaxIterations = %v, want %v (zero stops the search on entry)",
					dpCfg.MaxIterations, benchmark.DefaultMaxIterationsForTest())
			}

			if dpCfg.ResolutionPct != tt.wantResolution {
				t.Errorf("ResolutionPct = %v, want %v", dpCfg.ResolutionPct, tt.wantResolution)
			}
			if dpCfg.AcceptableLoss != tt.wantLoss {
				t.Errorf("AcceptableLoss = %v, want %v", dpCfg.AcceptableLoss, tt.wantLoss)
			}
			if dpCfg.TrialDuration != 5*time.Second {
				t.Errorf("TrialDuration = %v, want 5s", dpCfg.TrialDuration)
			}
			if dpCfg.Interface != "lo" || dpCfg.Peer != "192.0.2.1" || dpCfg.PeerPort != 3842 {
				t.Errorf("target not carried: %q %q %d", dpCfg.Interface, dpCfg.Peer, dpCfg.PeerPort)
			}
		})
	}
}

// TestBuildDataplaneConfigCarriesWarmup covers the one remaining parameter the
// builder reads from Params, so the whole boundary is asserted in one place.
func TestBuildDataplaneConfigCarriesWarmup(t *testing.T) {
	dpCfg := benchmark.BuildDataplaneConfigForTest(&modtypes.TestConfig{
		Interface: "lo",
		Duration:  5,
		Params:    map[string]any{"warmup": 2},
	})

	if dpCfg.WarmupPeriod != 2*time.Second {
		t.Errorf("WarmupPeriod = %v, want 2s", dpCfg.WarmupPeriod)
	}
}
