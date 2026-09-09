// SPDX-License-Identifier: BUSL-1.1

// Behavioural tests for the parameter defaulting that Execute applies before it
// hands a configuration to the dataplane. Execute itself cannot be driven past
// that point in a unit test: the only context a test can build owns no C
// resources, and RunCustomStreamTest passes it straight through (issue #1096).
package trafficgen_test

import (
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
	"github.com/MustardSeedNetworks/stem/internal/services/orchestrator/dataplane"
	"github.com/MustardSeedNetworks/stem/internal/services/trafficgen"
)

// TestBuildTrafficGenConfigDefaults pins the value used for every parameter the
// caller leaves unset. The literals are the WebUI defaults; asserting them
// against the package's own constants would pin nothing.
func TestBuildTrafficGenConfigDefaults(t *testing.T) {
	for _, params := range []struct {
		name string
		in   map[string]any
	}{
		{name: "nil params", in: nil},
		{name: "empty params", in: map[string]any{}},
	} {
		t.Run(params.name, func(t *testing.T) {
			got := trafficgen.BuildTrafficGenConfig(&modtypes.TestConfig{
				Interface: "eth0",
				FrameSize: 1518,
				Duration:  10,
				Params:    params.in,
			})

			if got.FrameSize != 1518 {
				t.Errorf("FrameSize = %d, want 1518", got.FrameSize)
			}
			if got.RatePct != 100.0 {
				t.Errorf("RatePct = %v, want 100", got.RatePct)
			}
			if got.WarmupSec != 2 {
				t.Errorf("WarmupSec = %d, want 2", got.WarmupSec)
			}
			if got.StreamID != 1 {
				t.Errorf("StreamID = %d, want 1", got.StreamID)
			}
			if got.BurstMode {
				t.Error("BurstMode = true, want false")
			}
			if got.BurstSize != 100 {
				t.Errorf("BurstSize = %d, want 100", got.BurstSize)
			}
			if got.InterBurstGapUs != 1000 {
				t.Errorf("InterBurstGapUs = %d, want 1000", got.InterBurstGapUs)
			}
			if got.SrcMac != "" {
				t.Errorf("SrcMac = %q, want empty", got.SrcMac)
			}
			if got.DstMac != "" {
				t.Errorf("DstMac = %q, want empty", got.DstMac)
			}
			if got.VlanID != 0 {
				t.Errorf("VlanID = %d, want 0", got.VlanID)
			}
			if got.VlanPriority != 0 {
				t.Errorf("VlanPriority = %d, want 0", got.VlanPriority)
			}
		})
	}
}

// TestBuildTrafficGenConfigReadsEveryParameter pins the key each field is read
// from. A renamed key silently reverts that field to its default, which is what
// #1129 does to the whole map, so each key is asserted on its own.
func TestBuildTrafficGenConfigReadsEveryParameter(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value any
		check func(t *testing.T, got *dataplane.TrafficGenConfig)
	}{
		{
			name:  "rate_pct",
			key:   "rate_pct",
			value: 42.5,
			check: func(t *testing.T, got *dataplane.TrafficGenConfig) {
				t.Helper()
				if got.RatePct != 42.5 {
					t.Errorf("RatePct = %v, want 42.5", got.RatePct)
				}
			},
		},
		{
			name:  "warmup_sec",
			key:   "warmup_sec",
			value: float64(7),
			check: func(t *testing.T, got *dataplane.TrafficGenConfig) {
				t.Helper()
				if got.WarmupSec != 7 {
					t.Errorf("WarmupSec = %d, want 7", got.WarmupSec)
				}
			},
		},
		{
			name:  "stream_id",
			key:   "stream_id",
			value: float64(9),
			check: func(t *testing.T, got *dataplane.TrafficGenConfig) {
				t.Helper()
				if got.StreamID != 9 {
					t.Errorf("StreamID = %d, want 9", got.StreamID)
				}
			},
		},
		{
			name:  "burst_mode",
			key:   "burst_mode",
			value: true,
			check: func(t *testing.T, got *dataplane.TrafficGenConfig) {
				t.Helper()
				if !got.BurstMode {
					t.Error("BurstMode = false, want true")
				}
			},
		},
		{
			name:  "burst_size",
			key:   "burst_size",
			value: float64(500),
			check: func(t *testing.T, got *dataplane.TrafficGenConfig) {
				t.Helper()
				if got.BurstSize != 500 {
					t.Errorf("BurstSize = %d, want 500", got.BurstSize)
				}
			},
		},
		{
			name:  "inter_burst_gap_us",
			key:   "inter_burst_gap_us",
			value: float64(250),
			check: func(t *testing.T, got *dataplane.TrafficGenConfig) {
				t.Helper()
				if got.InterBurstGapUs != 250 {
					t.Errorf("InterBurstGapUs = %d, want 250", got.InterBurstGapUs)
				}
			},
		},
		{
			name:  "src_mac",
			key:   "src_mac",
			value: "00:11:22:33:44:55",
			check: func(t *testing.T, got *dataplane.TrafficGenConfig) {
				t.Helper()
				if got.SrcMac != "00:11:22:33:44:55" {
					t.Errorf("SrcMac = %q, want 00:11:22:33:44:55", got.SrcMac)
				}
			},
		},
		{
			name:  "dst_mac",
			key:   "dst_mac",
			value: "66:77:88:99:aa:bb",
			check: func(t *testing.T, got *dataplane.TrafficGenConfig) {
				t.Helper()
				if got.DstMac != "66:77:88:99:aa:bb" {
					t.Errorf("DstMac = %q, want 66:77:88:99:aa:bb", got.DstMac)
				}
			},
		},
		{
			name:  "vlan_id",
			key:   "vlan_id",
			value: float64(4094),
			check: func(t *testing.T, got *dataplane.TrafficGenConfig) {
				t.Helper()
				if got.VlanID != 4094 {
					t.Errorf("VlanID = %d, want 4094", got.VlanID)
				}
			},
		},
		{
			name:  "vlan_priority",
			key:   "vlan_priority",
			value: float64(7),
			check: func(t *testing.T, got *dataplane.TrafficGenConfig) {
				t.Helper()
				if got.VlanPriority != 7 {
					t.Errorf("VlanPriority = %d, want 7", got.VlanPriority)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trafficgen.BuildTrafficGenConfig(&modtypes.TestConfig{
				FrameSize: 1518,
				Duration:  10,
				Params:    map[string]any{tt.key: tt.value},
			})
			tt.check(t, got)
		})
	}
}

// TestBuildTrafficGenConfigDuration covers the precedence between the Duration
// field and the "duration_sec" parameter, including the SafeIntToUint32 clamp
// that turns a negative duration into the parameter fallback.
func TestBuildTrafficGenConfigDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration int
		params   map[string]any
		want     uint32
	}{
		{name: "field wins when set", duration: 30, params: nil, want: 30},
		{name: "zero field falls back to default", duration: 0, params: nil, want: 60},
		{
			name:     "zero field falls back to parameter",
			duration: 0,
			params:   map[string]any{"duration_sec": float64(15)},
			want:     15,
		},
		{
			name:     "field takes precedence over parameter",
			duration: 30,
			params:   map[string]any{"duration_sec": float64(15)},
			want:     30,
		},
		{
			name:     "negative field clamps to zero then falls back",
			duration: -5,
			params:   map[string]any{"duration_sec": float64(15)},
			want:     15,
		},
		{name: "negative field with no parameter uses default", duration: -5, params: nil, want: 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trafficgen.BuildTrafficGenConfig(&modtypes.TestConfig{
				FrameSize: 1518,
				Duration:  tt.duration,
				Params:    tt.params,
			})
			if got.DurationSec != tt.want {
				t.Errorf("DurationSec = %d, want %d", got.DurationSec, tt.want)
			}
		})
	}
}

// TestBuildTrafficGenConfigClampsOversizedValues pins what the VLAN conversions
// do with a value that does not fit the field, which the API's JSON decoding
// does not prevent.
func TestBuildTrafficGenConfigClampsOversizedValues(t *testing.T) {
	got := trafficgen.BuildTrafficGenConfig(&modtypes.TestConfig{
		FrameSize: 1518,
		Duration:  10,
		Params: map[string]any{
			"vlan_id":       float64(100000),
			"vlan_priority": float64(300),
		},
	})

	if got.VlanID != 0 {
		t.Errorf("VlanID = %d, want 0 (out-of-range value falls back to the default)", got.VlanID)
	}
	if got.VlanPriority != 0 {
		t.Errorf("VlanPriority = %d, want 0 (out-of-range value falls back to the default)", got.VlanPriority)
	}
}

// TestExecuteRejectsEveryTypeOutsideTestTypes asserts the module runs exactly
// the types it advertises. Execute no longer carries a second custom_stream
// check, so a type added to TestTypes without an execution path fails here
// rather than running as a custom stream.
func TestExecuteRejectsEveryTypeOutsideTestTypes(t *testing.T) {
	executor := trafficgen.NewMockExecutor()

	advertised := executor.TestTypes()
	if len(advertised) != 1 || advertised[0] != "custom_stream" {
		t.Fatalf("TestTypes() = %v; Execute has an execution path for custom_stream only", advertised)
	}

	rejected := []string{"rfc2544_throughput", "y1564", "reflect", "rfc2889_forwarding", "custom", "trafficgen", ""}
	for _, testType := range rejected {
		t.Run("rejects "+testType, func(t *testing.T) {
			result, err := executor.Execute(testType, &modtypes.TestConfig{FrameSize: 1518, Duration: 60})
			if err == nil {
				t.Fatalf("Execute(%q) returned no error", testType)
			}
			if result != nil {
				t.Errorf("Execute(%q) result = %+v, want nil", testType, result)
			}
		})
	}
}

// TestExecutorCancel covers Cancel against both context states. The cgo and stub
// implementations both guard the absent C context, so neither call reaches the
// dataplane.
func TestExecutorCancel(t *testing.T) {
	t.Run("nil context", func(t *testing.T) {
		trafficgen.NewMockExecutor().Cancel()
	})
	t.Run("test context", func(t *testing.T) {
		trafficgen.NewExecutorWithTestContext().Cancel()
	})
}

// TestExecutorCloseWithTestContext covers the non-nil branch of Close.
func TestExecutorCloseWithTestContext(t *testing.T) {
	executor := trafficgen.NewExecutorWithTestContext()
	executor.Close()
	executor.Close() // Close is idempotent; the second call must not panic.
}
