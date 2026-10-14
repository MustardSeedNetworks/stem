// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/services/orchestrator/dataplane"
	"github.com/MustardSeedNetworks/stem/internal/services/servicetest"
)

// mefRecorder passes every MEF test and records the configuration each one
// reached the dataplane with. The embedded interface is nil, so a Y.1564
// runner panics if called.
type mefRecorder struct {
	servicetest.ServiceDataplane

	mu      sync.Mutex
	configs []dataplane.MEFConfig
}

func (*mefRecorder) Configure(*dataplane.Config) error { return nil }
func (*mefRecorder) Close()                            {}

func (r *mefRecorder) record(cfg *dataplane.MEFConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.configs = append(r.configs, *cfg)
}

func (r *mefRecorder) RunMEFConfigTest(cfg *dataplane.MEFConfig) (*dataplane.MEFConfigResult, error) {
	r.record(cfg)
	return &dataplane.MEFConfigResult{OverallPassed: true}, nil
}

func (r *mefRecorder) RunMEFPerfTest(cfg *dataplane.MEFConfig) (*dataplane.MEFPerfResult, error) {
	r.record(cfg)
	return &dataplane.MEFPerfResult{OverallPassed: true}, nil
}

func (r *mefRecorder) recorded() []dataplane.MEFConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]dataplane.MEFConfig(nil), r.configs...)
}

// MEF 48/49 service activation is configured by the Y.1564 block. Before
// stem#1556 no MEF test type read any block, so every API and UI MEF run used
// the module defaults: 100 Mbps CIR, one 1518-byte frame, 10/5 ms thresholds.
func TestRunPlanMEFHonoursItsServiceConfig(t *testing.T) {
	const y1564 = `"config":{"y1564":{"cir":250,"eir":50,"frameSizes":[512,1024],` +
		`"configStepDuration":5,"perfTestDuration":30,` +
		`"flrThreshold":0.5,"fdThreshold":7,"fdvThreshold":3}}`
	tests := []struct {
		testType string
		want     []uint32
	}{
		{testType: "mef_config", want: []uint32{512, 1024}},
		{testType: "mef_perf", want: []uint32{512, 1024}},
	}
	for _, tc := range tests {
		t.Run(tc.testType, func(t *testing.T) {
			recorder := &mefRecorder{}
			s, token, ifaceName := startPlanServer(t, func(string) (api.TestExecutor, error) {
				return servicetest.NewExecutorWithDataplane(recorder), nil
			})
			runPlanUntil(t, s, token, ifaceName, `{"testType":"`+tc.testType+`",`+y1564+`}`, "completed")

			configs := recorder.recorded()
			sizes := make([]uint32, 0, len(configs))
			for _, cfg := range configs {
				sizes = append(sizes, cfg.FrameSizes...)
				assertMEFServiceConfig(t, cfg)
			}
			if !reflect.DeepEqual(sizes, tc.want) {
				t.Errorf("measured frame sizes = %v, want %v", sizes, tc.want)
			}
		})
	}
}

func assertMEFServiceConfig(t *testing.T, cfg dataplane.MEFConfig) {
	t.Helper()
	got := [...]float64{
		cfg.CIRMbps, cfg.EIRMbps, cfg.FLRThresholdPct, cfg.FDThresholdUs, cfg.FDVThresholdUs,
		float64(cfg.ConfigDurationSec), float64(cfg.PerfDurationSec),
	}
	want := [...]float64{250, 50, 0.5, 7000, 3000, 5, 30}
	if got != want {
		t.Errorf("CIR, EIR, FLR %%, FD µs, FDV µs, step s, perf s = %v, want %v", got, want)
	}
}
