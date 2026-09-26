// SPDX-License-Identifier: BUSL-1.1

package servicetest_test

import (
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
	"github.com/MustardSeedNetworks/stem/internal/services/orchestrator/dataplane"
	"github.com/MustardSeedNetworks/stem/internal/services/servicetest"
)

// verdictDataplane answers every runner with a result whose service verdict
// is pass, so a test flips one verdict and checks the executor reports it.
type verdictDataplane struct {
	y1564Config *dataplane.Y1564ConfigResult
	y1564Perf   *dataplane.Y1564PerfResult
	mefConfig   *dataplane.MEFConfigResult
	mefPerf     *dataplane.MEFPerfResult
}

var _ servicetest.ServiceDataplane = (*verdictDataplane)(nil)

func passingDataplane() *verdictDataplane {
	return &verdictDataplane{
		y1564Config: &dataplane.Y1564ConfigResult{ServiceID: 1, ServicePass: true},
		y1564Perf:   &dataplane.Y1564PerfResult{ServiceID: 1, ServicePass: true},
		mefConfig:   &dataplane.MEFConfigResult{ServiceID: "evc-1", OverallPassed: true},
		mefPerf:     &dataplane.MEFPerfResult{ServiceID: "evc-1", OverallPassed: true},
	}
}

func (d *verdictDataplane) Configure(*dataplane.Config) error { return nil }

func (d *verdictDataplane) RunY1564ConfigTest(*dataplane.Y1564Service) (*dataplane.Y1564ConfigResult, error) {
	return d.y1564Config, nil
}

func (d *verdictDataplane) RunY1564PerfTest(*dataplane.Y1564Service, uint32) (*dataplane.Y1564PerfResult, error) {
	return d.y1564Perf, nil
}

func (d *verdictDataplane) RunMEFConfigTest(*dataplane.MEFConfig) (*dataplane.MEFConfigResult, error) {
	return d.mefConfig, nil
}

func (d *verdictDataplane) RunMEFPerfTest(*dataplane.MEFConfig) (*dataplane.MEFPerfResult, error) {
	return d.mefPerf, nil
}

func (d *verdictDataplane) RunMEFFullTest(
	*dataplane.MEFConfig,
) (*dataplane.MEFConfigResult, *dataplane.MEFPerfResult, error) {
	return d.mefConfig, d.mefPerf, nil
}

func (d *verdictDataplane) Cancel() {}
func (d *verdictDataplane) Close()  {}

// A service that misses its acceptance criteria is a failed test with its
// measurements attached, never a pass (#1463).
func TestExecuteReportsServiceVerdict(t *testing.T) {
	tests := []struct {
		name     string
		testType string
		fail     func(*verdictDataplane)
	}{
		{"y1564_config", "y1564_config", func(d *verdictDataplane) { d.y1564Config.ServicePass = false }},
		{"y1564_perf", "y1564_perf", func(d *verdictDataplane) { d.y1564Perf.ServicePass = false }},
		{"y1564 config half", "y1564", func(d *verdictDataplane) { d.y1564Config.ServicePass = false }},
		{"y1564 perf half", "y1564", func(d *verdictDataplane) { d.y1564Perf.ServicePass = false }},
		{"mef_config", "mef_config", func(d *verdictDataplane) { d.mefConfig.OverallPassed = false }},
		{"mef_perf", "mef_perf", func(d *verdictDataplane) { d.mefPerf.OverallPassed = false }},
		{"mef config half", "mef", func(d *verdictDataplane) { d.mefConfig.OverallPassed = false }},
		{"mef perf half", "mef", func(d *verdictDataplane) { d.mefPerf.OverallPassed = false }},
	}
	cfg := func() *modtypes.TestConfig {
		return &modtypes.TestConfig{Interface: "eth0", Duration: 1, Params: map[string]any{}}
	}

	for _, tt := range tests {
		t.Run(tt.name+" passes", func(t *testing.T) {
			result, err := servicetest.NewExecutorWithDataplane(passingDataplane()).Execute(tt.testType, cfg())
			if err != nil {
				t.Fatalf("Execute: %v", err)
			}
			if !result.Success || result.Error != "" {
				t.Errorf("passing service: success=%v error=%q, want success and no error",
					result.Success, result.Error)
			}
		})
		t.Run(tt.name+" fails", func(t *testing.T) {
			dp := passingDataplane()
			tt.fail(dp)
			result, err := servicetest.NewExecutorWithDataplane(dp).Execute(tt.testType, cfg())
			if err != nil {
				t.Fatalf("Execute: %v (a failed verdict is a result, not an execution error)", err)
			}
			if result.Success {
				t.Error("a service that failed its acceptance criteria reported success")
			}
			if result.Error == "" {
				t.Error("a failed service carries no operator-facing error")
			}
			if result.Data == nil {
				t.Error("a failed service dropped its measurements")
			}
		})
	}
}
