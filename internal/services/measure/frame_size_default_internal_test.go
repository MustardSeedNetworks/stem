// SPDX-License-Identifier: BUSL-1.1

package measure

import "testing"

// minExtendedFrameSize restates api.MinExtendedFrameSize, which
// TestExtendedFrameMinimumMatchesTheHeader pins to EXTENDED_MIN_FRAME_SIZE in
// include/rfc2544.h. internal/api depends on this package, so importing it
// here would be a cycle.
const minExtendedFrameSize = 70

// A Y.1731 request with no frame size gets this default, and every Y.1731
// trial sends it (stem#1555). The payload does not fit a 64-byte frame, so a
// smaller default would fail every bare run with -22.
func TestDefaultFrameSizeIsLegal(t *testing.T) {
	if defaultFrameSize < minExtendedFrameSize {
		t.Errorf("defaultFrameSize = %d, below the %d-byte extended-format minimum",
			defaultFrameSize, minExtendedFrameSize)
	}
}
