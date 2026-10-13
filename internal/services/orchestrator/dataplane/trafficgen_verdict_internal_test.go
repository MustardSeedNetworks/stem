// SPDX-License-Identifier: BUSL-1.1

package dataplane

import (
	"errors"
	"testing"
)

func TestCustomStreamVerdict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result TrafficGenResult
		want   error
	}{
		{"nothing sent", TrafficGenResult{}, ErrNoTrialsRun},
		{"nothing back", TrafficGenResult{PacketsSent: 117482, LossPct: 100}, ErrNoFramesReturned},
		{"partial loss", TrafficGenResult{PacketsSent: 1000, PacketsRecv: 1, LossPct: 99.9}, nil},
		{"all back", TrafficGenResult{PacketsSent: 1000, PacketsRecv: 1000}, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := customStreamVerdict(tc.result); !errors.Is(err, tc.want) {
				t.Errorf("customStreamVerdict() = %v, want %v", err, tc.want)
			}
		})
	}
}
