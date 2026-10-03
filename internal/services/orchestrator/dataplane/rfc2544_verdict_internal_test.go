// SPDX-License-Identifier: BUSL-1.1

package dataplane

import (
	"errors"
	"testing"
)

func TestThroughputVerdict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result ThroughputResult
		want   error
	}{
		{
			name:   "no trial ran",
			result: ThroughputResult{},
			want:   ErrNoTrialsRun,
		},
		{
			// The ST-2 bench with the reflector stopped (#1466): ten trials,
			// every one over the loss criterion.
			name:   "every trial lost every frame",
			result: ThroughputResult{FrameSize: 64, Iterations: 10, FramesTested: 412074},
			want:   ErrNoFramesReturned,
		},
		{
			name: "the peer answered but lost frames at every rate",
			result: ThroughputResult{
				FrameSize: 64, Iterations: 10, FramesTested: 412074, FramesReceived: 400000,
			},
			want: ErrThroughputNoPassingRate,
		},
		{
			name: "a rate met the loss criterion",
			result: ThroughputResult{
				FrameSize: 64, Iterations: 10, OfferedRatePct: 7.2,
				MaxRatePct: 7.2, MaxRatePps: 107000,
			},
			want: nil,
		},
		{
			// #1233: a trial the generator could not drive ends the search at
			// the rate it carried, which met the loss criterion.
			name: "generator limited at a passing trial",
			result: ThroughputResult{
				FrameSize: 64, Iterations: 1, OfferedRatePct: 50,
				MaxRatePct: 0.1, MaxRatePps: 7000, GeneratorLimited: true,
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := throughputVerdict(tt.result); !errors.Is(got, tt.want) {
				t.Fatalf("throughputVerdict() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLatencyVerdict(t *testing.T) {
	t.Parallel()

	answered := LatencyStats{Count: 9800, MinNs: 41000, AvgNs: 52000, MaxNs: 97000}
	tests := []struct {
		name    string
		results []LatencyResultCLI
		want    error
	}{
		{
			// #1275: every load level ran and timed nothing.
			name: "no level timed a returned frame",
			results: []LatencyResultCLI{
				{FrameSize: 64, LoadPct: 10}, {FrameSize: 64, LoadPct: 50}, {FrameSize: 64, LoadPct: 90},
			},
			want: ErrNoFramesReturned,
		},
		{
			name: "only the highest load went unanswered",
			results: []LatencyResultCLI{
				{FrameSize: 64, LoadPct: 10, Latency: answered}, {FrameSize: 64, LoadPct: 90},
			},
			want: nil,
		},
		{
			name:    "every level answered",
			results: []LatencyResultCLI{{FrameSize: 64, LoadPct: 50, Latency: answered}},
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := latencyVerdict(tt.results); !errors.Is(got, tt.want) {
				t.Fatalf("latencyVerdict() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFrameLossVerdict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		points []FrameLossPoint
		want   error
	}{
		{
			name: "no trial ran",
			want: ErrNoTrialsRun,
		},
		{
			// #1275: 100 % loss at every rate is a silent peer.
			name: "nothing came back at any rate",
			points: []FrameLossPoint{
				{OfferedRatePct: 100, FramesSent: 148809, LossPct: 100},
				{OfferedRatePct: 90, FramesSent: 133928, LossPct: 100},
			},
			want: ErrNoFramesReturned,
		},
		{
			// A lossy but answering peer is the measurement this test exists for.
			name: "the peer answered with loss",
			points: []FrameLossPoint{
				{OfferedRatePct: 100, FramesSent: 148809, LossPct: 100},
				{OfferedRatePct: 90, FramesSent: 133928, FramesRecv: 120000, LossPct: 10.4},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := frameLossVerdict(tt.points); !errors.Is(got, tt.want) {
				t.Fatalf("frameLossVerdict() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackToBackVerdict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result BurstResult
		want   error
	}{
		{
			// #1275: the first burst lost every frame, so the search reported
			// a maximum burst of zero frames.
			name:   "nothing came back from any burst",
			result: BurstResult{FrameSize: 64},
			want:   ErrNoFramesReturned,
		},
		{
			// A peer that drops part of the first burst measures a zero burst.
			name:   "the peer answered but dropped part of the first burst",
			result: BurstResult{FrameSize: 64, FramesRecv: 900},
			want:   nil,
		},
		{
			name:   "bursts came back whole",
			result: BurstResult{FrameSize: 64, MaxBurst: 4096, Trials: 3, FramesRecv: 30000},
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := backToBackVerdict(tt.result); !errors.Is(got, tt.want) {
				t.Fatalf("backToBackVerdict() = %v, want %v", got, tt.want)
			}
		})
	}
}
