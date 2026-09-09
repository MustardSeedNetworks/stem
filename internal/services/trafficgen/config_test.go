// SPDX-License-Identifier: BUSL-1.1

package trafficgen_test

import (
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
	"github.com/MustardSeedNetworks/stem/internal/services/orchestrator/dataplane"
	"github.com/MustardSeedNetworks/stem/internal/services/trafficgen"
)

// Behavioural tests for the parameter defaulting that Execute applies before it
// hands a configuration to the dataplane. Execute itself cannot be driven past
// that point in a unit test: the only context a test can build owns no C
// resources, and RunCustomStreamTest passes it straight through (issue #1096).

// defaultConfig is what buildTrafficGenConfig produces for a 1518-byte, 10-second
// run with no parameters set. The literals are the WebUI defaults; asserting them
// against the package's own constants would pin nothing.
func defaultConfig() dataplane.TrafficGenConfig {
	return dataplane.TrafficGenConfig{
		FrameSize:       1518,
		RatePct:         100.0,
		DurationSec:     10,
		WarmupSec:       2,
		StreamID:        1,
		BurstMode:       false,
		BurstSize:       100,
		InterBurstGapUs: 1000,
		SrcMac:          "",
		DstMac:          "",
		VlanID:          0,
		VlanPriority:    0,
	}
}

func buildWith(params map[string]any) *dataplane.TrafficGenConfig {
	return trafficgen.BuildTrafficGenConfig(&modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 1518,
		Duration:  10,
		Params:    params,
	})
}

// TestBuildTrafficGenConfigDefaults pins every value used when the caller sets
// no parameters at all.
func TestBuildTrafficGenConfigDefaults(t *testing.T) {
	for _, tt := range []struct {
		name   string
		params map[string]any
	}{
		{name: "nil params", params: nil},
		{name: "empty params", params: map[string]any{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			want := defaultConfig()
			if got := *buildWith(tt.params); got != want {
				t.Errorf("BuildTrafficGenConfig() = %+v, want %+v", got, want)
			}
		})
	}
}

// TestBuildTrafficGenConfigReadsEveryParameter pins the key each field is read
// from, and that reading it disturbs nothing else. A renamed key silently
// reverts its field to the default, which is what #1129 does to the whole map.
func TestBuildTrafficGenConfigReadsEveryParameter(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value any
		apply func(*dataplane.TrafficGenConfig)
	}{
		{"rate_pct", "rate_pct", 42.5, func(c *dataplane.TrafficGenConfig) { c.RatePct = 42.5 }},
		{"warmup_sec", "warmup_sec", float64(7), func(c *dataplane.TrafficGenConfig) { c.WarmupSec = 7 }},
		{"stream_id", "stream_id", float64(9), func(c *dataplane.TrafficGenConfig) { c.StreamID = 9 }},
		{"burst_mode", "burst_mode", true, func(c *dataplane.TrafficGenConfig) { c.BurstMode = true }},
		{"burst_size", "burst_size", float64(500), func(c *dataplane.TrafficGenConfig) { c.BurstSize = 500 }},
		{
			"inter_burst_gap_us", "inter_burst_gap_us", float64(250),
			func(c *dataplane.TrafficGenConfig) { c.InterBurstGapUs = 250 },
		},
		{
			"src_mac", "src_mac", "00:11:22:33:44:55",
			func(c *dataplane.TrafficGenConfig) { c.SrcMac = "00:11:22:33:44:55" },
		},
		{
			"dst_mac", "dst_mac", "66:77:88:99:aa:bb",
			func(c *dataplane.TrafficGenConfig) { c.DstMac = "66:77:88:99:aa:bb" },
		},
		{"vlan_id", "vlan_id", float64(4094), func(c *dataplane.TrafficGenConfig) { c.VlanID = 4094 }},
		{"vlan_priority", "vlan_priority", float64(7), func(c *dataplane.TrafficGenConfig) { c.VlanPriority = 7 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := defaultConfig()
			tt.apply(&want)
			if got := *buildWith(map[string]any{tt.key: tt.value}); got != want {
				t.Errorf("BuildTrafficGenConfig(%s=%v) = %+v, want %+v", tt.key, tt.value, got, want)
			}
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
	want := defaultConfig()
	got := *buildWith(map[string]any{"vlan_id": float64(100000), "vlan_priority": float64(300)})
	if got != want {
		t.Errorf("BuildTrafficGenConfig() = %+v, want %+v (out-of-range values fall back to defaults)", got, want)
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

// TestExecutorLifecycle covers Close and Cancel against both context states.
// Both dataplane implementations guard the absent C context, so neither call
// reaches it; a nil dereference here is the #1096 shape.
func TestExecutorLifecycle(t *testing.T) {
	tests := []struct {
		name string
		run  func(*trafficgen.Executor)
	}{
		{name: "cancel", run: func(e *trafficgen.Executor) { e.Cancel() }},
		{name: "close", run: func(e *trafficgen.Executor) { e.Close() }},
		{name: "close twice", run: func(e *trafficgen.Executor) { e.Close(); e.Close() }},
		{name: "close then cancel", run: func(e *trafficgen.Executor) { e.Close(); e.Cancel() }},
	}

	executors := map[string]func() *trafficgen.Executor{
		"nil context":  trafficgen.NewMockExecutor,
		"test context": trafficgen.NewExecutorWithTestContext,
	}

	for ctxName, newExecutor := range executors {
		for _, tt := range tests {
			t.Run(ctxName+"/"+tt.name, func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("%s panicked: %v", tt.name, r)
					}
				}()
				tt.run(newExecutor())
			})
		}
	}
}
