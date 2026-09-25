//go:build cgo && linux

package servicetest_test

import (
	"os"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
	"github.com/MustardSeedNetworks/stem/internal/services/orchestrator/dataplane"
	"github.com/MustardSeedNetworks/stem/internal/services/servicetest"
)

// Y.1564 through the executor the daemon runs, against a real reflector
// (#1411). Every daemon Y.1564 run used to fail -22 before sending a frame, and
// no test could see it: the C tests prepared their contexts by hand.
//
// It needs a reflector to answer and CAP_NET_RAW, so it runs only when pointed
// at one:
//
//	STEM_Y1564_INTERFACE=eth0 STEM_Y1564_PEER=172.18.0.2 \
//	  go test -run Y1564Reflected ./internal/services/servicetest/
const (
	reflectorPort = 3842
	cirMbps       = 10.0
	frameSize     = 512
	// Preamble and inter-frame gap, which the dataplane's rate includes.
	wireOverhead = 20
)

func reflectedConfig(t *testing.T, params map[string]any) (string, *modtypes.TestConfig) {
	t.Helper()
	iface, peer := os.Getenv("STEM_Y1564_INTERFACE"), os.Getenv("STEM_Y1564_PEER")
	if iface == "" || peer == "" {
		t.Skip("set STEM_Y1564_INTERFACE and STEM_Y1564_PEER to run Y.1564 against a reflector")
	}
	params["cir"] = cirMbps
	return iface, &modtypes.TestConfig{
		Interface: iface,
		Peer:      peer,
		PeerPort:  reflectorPort,
		FrameSize: frameSize,
		Params:    params,
	}
}

func execute(t *testing.T, testType string, iface string, cfg *modtypes.TestConfig) any {
	t.Helper()
	exec, err := servicetest.NewExecutor(iface)
	if err != nil {
		t.Fatalf("NewExecutor(%q): %v", iface, err)
	}
	defer exec.Close()

	result, err := exec.Execute(testType, cfg)
	if err != nil {
		t.Fatalf("Execute(%s): %v", testType, err)
	}
	return result.Data
}

func TestY1564ConfigReflectedEveryStep(t *testing.T) {
	iface, cfg := reflectedConfig(t, map[string]any{"config_duration_sec": uint32(1)})
	data, ok := execute(t, "y1564_config", iface, cfg).(*dataplane.Y1564ConfigResult)
	if !ok {
		t.Fatalf("y1564_config data is not a *Y1564ConfigResult")
	}

	want := [len(data.Steps)]float64{25, 50, 75, 100}
	for i, step := range data.Steps {
		t.Logf("step %d: %.0f%% CIR, tx %d, rx %d",
			step.Step, step.OfferedRatePct, step.FramesTx, step.FramesRx)
		if step.OfferedRatePct != want[i] {
			t.Errorf("step %d ran at %.0f%% of the CIR, want %.0f%%", i+1, step.OfferedRatePct, want[i])
		}
		if step.FramesRx == 0 {
			t.Errorf("step %d: the reflector returned no frames (tx %d)", i+1, step.FramesTx)
		}
		// One second at the step's share of the CIR: a step that ran for the
		// dataplane's default 60 s, or at the wrong rate, is far outside this.
		wantTx := cirMbps * 1e6 * want[i] / 100 / ((frameSize + wireOverhead) * 8)
		if tx := float64(step.FramesTx); tx < 0.9*wantTx || tx > 1.1*wantTx {
			t.Errorf("step %d sent %d frames, want about %.0f (1 s at %.0f%% of %.0f Mbps)",
				i+1, step.FramesTx, wantTx, want[i], cirMbps)
		}
		if i > 0 && step.FramesTx <= data.Steps[i-1].FramesTx {
			t.Errorf("step %d sent %d frames, no more than step %d's %d",
				i+1, step.FramesTx, i, data.Steps[i-1].FramesTx)
		}
	}
}

func TestY1564PerfReflected(t *testing.T) {
	iface, cfg := reflectedConfig(t, map[string]any{})
	cfg.Duration = 1
	data, ok := execute(t, "y1564_perf", iface, cfg).(*dataplane.Y1564PerfResult)
	if !ok {
		t.Fatalf("y1564_perf data is not a *Y1564PerfResult")
	}
	t.Logf("perf: tx %d, rx %d, FLR %.4f%%", data.FramesTx, data.FramesRx, data.FLRPct)
	if data.FramesRx == 0 {
		t.Errorf("the reflector returned no frames (tx %d)", data.FramesTx)
	}
}
