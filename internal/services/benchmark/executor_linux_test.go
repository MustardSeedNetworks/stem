//go:build cgo && linux

package benchmark_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/services/benchmark"
	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
	"github.com/MustardSeedNetworks/stem/internal/services/orchestrator/dataplane"
)

func TestExecuteLatencyDoesNotCrash(t *testing.T) {
	if os.Getenv("STEM_LATENCY_CRASH_HELPER") == "1" {
		ctx, err := dataplane.NewContext("lo")
		if err != nil {
			return
		}
		defer ctx.Close()

		executor := benchmark.NewExecutorWithContext(ctx)
		_, _ = executor.Execute("rfc2544_latency", &modtypes.TestConfig{
			Interface: "lo",
			FrameSize: 128,
			Duration:  1,
			Params: map[string]any{
				"load_levels": []float64{10},
			},
		})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestExecuteLatencyDoesNotCrash$")
	cmd.Env = append(os.Environ(), "STEM_LATENCY_CRASH_HELPER=1")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("RFC 2544 latency execution terminated the process: %v\n%s", err, output)
	}
}
