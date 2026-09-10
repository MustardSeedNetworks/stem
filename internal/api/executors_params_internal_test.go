// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"reflect"
	"testing"
)

func TestConvertToModuleConfigUsesExecutorParameterNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testType string
		config   *TestConfig
		want     map[string]any
	}{
		{
			name:     "RFC 2544",
			testType: "rfc2544_throughput",
			config: &TestConfig{RFC2544: &RFC2544TestConfig{
				Duration: 11, FrameSizes: []uint32{64, 512}, Resolution: 0.5,
				MaxLoss: 0.25, Warmup: 3, Trials: 4, StepSize: 2.5, Bidirectional: true,
			}},
			want: map[string]any{
				"duration": 11, "frame_sizes": []uint32{64, 512}, "resolution": 0.5,
				"max_loss": 0.25, "warmup": 3, "trials": 4, "step_size": 2.5,
				"bidirectional": true,
			},
		},
		{
			name:     "RFC 2889",
			testType: "rfc2889_forwarding",
			config: &TestConfig{RFC2889: &RFC2889TestConfig{
				FrameSize: 128, Duration: 12, Warmup: 2, AddressCount: 20,
				AcceptableLoss: 0.1, PortCount: 8, Pattern: 1,
			}},
			want: map[string]any{
				"frame_size": uint32(128), "duration_sec": uint32(12),
				"warmup_sec": uint32(2), "address_count": uint32(20),
				"acceptable_loss_pct": 0.1, "port_count": uint32(8), "pattern": uint32(1),
			},
		},
		{
			name:     "RFC 6349",
			testType: "rfc6349_throughput",
			config: &TestConfig{RFC6349: &RFC6349TestConfig{
				TargetRateMbps: 900, MinRTTMs: 2, MaxRTTMs: 40, RWNDSize: 65535,
				Duration: 13, ParallelStreams: 3, MSS: 1460, Mode: 2,
			}},
			want: map[string]any{
				"target_rate_mbps": float64(900), "min_rtt_ms": float64(2),
				"max_rtt_ms": float64(40), "rwnd_size": uint32(65535),
				"duration_sec": uint32(13), "parallel_streams": uint32(3),
				"mss": uint32(1460), "mode": uint32(2),
			},
		},
		{
			name:     "Y.1564",
			testType: "y1564_config",
			config: &TestConfig{Y1564: &Y1564TestConfig{
				CIR: 100, EIR: 50, CBS: 32, EBS: 16, FrameSizes: []uint32{256},
				ConfigStepDuration: 14, PerfTestDuration: 60, VlanID: 200, PCP: 5,
				ColorAware: true, FLRThreshold: 0.01, FDThreshold: 5, FDVThreshold: 2,
			}},
			want: map[string]any{
				"cir": float64(100), "eir": float64(50), "cbs": uint32(32),
				"ebs": uint32(16), "frame_sizes": []uint32{256},
				"config_duration_sec": uint32(14), "perf_duration_sec": uint32(60),
				"vlan_id": uint16(200), "cos": uint8(5), "color_aware": true,
				"flr_threshold_pct": 0.01, "fd_threshold_ms": float64(5),
				"fdv_threshold_ms": float64(2),
			},
		},
		{
			name:     "Y.1731",
			testType: "y1731_delay",
			config: &TestConfig{Y1731: &Y1731TestConfig{
				MepID: 10, MegLevel: 3, MegID: "metro", CCMInterval: 100,
				Priority: 6, Duration: 15, IntervalMs: 20, Count: 30,
				FrameSize: 512, PriorityTagged: true,
			}},
			want: map[string]any{
				"mep_id": uint32(10), "meg_level": uint32(3), "meg_id": "metro",
				"ccm_interval": uint32(100), "priority": uint8(6), "duration": uint32(15),
				"interval_ms": uint32(20), "count": uint32(30), "frame_size": uint32(512),
				"priority_tagged": true,
			},
		},
		{
			name:     "TSN",
			testType: "tsn_full",
			config: &TestConfig{TSN: &TSNTestConfig{
				Duration: 16, Warmup: 2, FrameSize: 1024, MaxLatencyNs: 1000,
				MaxJitterNs: 200, RequirePTPSync: true, MaxSyncOffsetNs: 50,
				PTPEnabled: true, PreemptionEnabled: true, NumTrafficClasses: 4,
				BaseTimeNs: 10000, CycleTimeNs: 20000, TrafficClass: 3,
			}},
			want: map[string]any{
				"duration_sec": uint32(16), "warmup_sec": uint32(2),
				"frame_size": uint32(1024), "max_latency_ns": uint32(1000),
				"max_jitter_ns": uint32(200), "require_ptp_sync": true,
				"max_sync_offset_ns": uint32(50), "ptp_enabled": true,
				"preemption_enabled": true, "num_traffic_classes": uint32(4),
				"base_time_ns": uint64(10000), "cycle_time_ns": uint32(20000),
				"traffic_class": uint32(3),
			},
		},
		{
			name:     "TrafficGen",
			testType: "custom_stream",
			config: &TestConfig{TrafficGen: &TrafficGenTestConfig{
				FrameSize: 1518, RatePct: 42.5, Duration: 17, Warmup: 4,
				StreamID: 9, BurstMode: true, BurstSize: 500, InterBurstGapUs: 250,
				SrcMac: "00:11:22:33:44:55", DstMac: "66:77:88:99:aa:bb",
				VlanID: 100, VlanPriority: 7,
			}},
			want: map[string]any{
				"frame_size": uint32(1518), "rate_pct": 42.5, "duration_sec": uint32(17),
				"warmup_sec": uint32(4), "stream_id": uint32(9), "burst_mode": true,
				"burst_size": uint32(500), "inter_burst_gap_us": uint32(250),
				"src_mac": "00:11:22:33:44:55", "dst_mac": "66:77:88:99:aa:bb",
				"vlan_id": uint16(100), "vlan_priority": uint8(7),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := convertToModuleConfig("eth0", tt.testType, tt.config)
			if !reflect.DeepEqual(got.Params, tt.want) {
				t.Fatalf("Params = %#v, want %#v", got.Params, tt.want)
			}
		})
	}
}
