// SPDX-License-Identifier: BUSL-1.1

package dataplane

import "testing"

func TestY1564NothingTransmitted(t *testing.T) {
	sentAll := Y1564ConfigResult{}
	for i := range sentAll.Steps {
		sentAll.Steps[i].FramesTx = 100
	}
	lastSilent := sentAll
	lastSilent.Steps[3].FramesTx = 0

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{"config, every step sent", sentAll.NothingTransmitted(), false},
		{"config, one step sent nothing", lastSilent.NothingTransmitted(), true},
		{"perf sent", (&Y1564PerfResult{FramesTx: 1}).NothingTransmitted(), false},
		{"perf sent nothing", (&Y1564PerfResult{}).NothingTransmitted(), true},
	}
	for _, tc := range tests {
		if tc.got != tc.want {
			t.Errorf("%s: NothingTransmitted() = %v, want %v", tc.name, tc.got, tc.want)
		}
	}
}
