// SPDX-License-Identifier: BUSL-1.1

package dataplane

import (
	"errors"
	"testing"
)

func TestY1731Verdict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sent     uint64
		received uint64
		want     error
	}{
		{"nothing sent", 0, 0, ErrNoTrialsRun},
		{"no reflection", 60, 0, ErrNoFramesReturned},
		{"partial loss", 60, 1, nil},
		{"full reflection", 60, 60, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := y1731Verdict("delay", tc.sent, tc.received); !errors.Is(err, tc.want) {
				t.Errorf("y1731Verdict(%d, %d) = %v, want %v", tc.sent, tc.received, err, tc.want)
			}
		})
	}
}
