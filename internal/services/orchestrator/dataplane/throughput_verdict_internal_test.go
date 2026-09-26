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
			want:   ErrThroughputNoIterations,
		},
		{
			// The ST-2 bench with the reflector stopped (#1466): ten trials,
			// every one over the loss criterion.
			name:   "every trial lost every frame",
			result: ThroughputResult{FrameSize: 64, Iterations: 10, FramesTested: 412074},
			want:   ErrThroughputNoFramesReturned,
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
