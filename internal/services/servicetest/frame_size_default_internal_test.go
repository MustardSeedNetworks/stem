// SPDX-License-Identifier: BUSL-1.1

package servicetest

import "testing"

// minExtendedFrameSize mirrors api.MinExtendedFrameSize, which mirrors
// EXTENDED_MIN_FRAME_SIZE in include/rfc2544.h. Importing internal/api here
// would be an import cycle, so the value is restated; both copies are pinned
// to the header by TestExtendedFrameMinimumMatchesTheHeader.
const minExtendedFrameSize = 70

// A bare POST with no Y.1564 config never sets FrameSize, so this default is
// what reaches the dataplane. The Y.1564 payload does not fit a 64-byte frame
// and the trial would fail with -22 (stem#1250).
func TestDefaultFrameSizeIsLegal(t *testing.T) {
	if defaultFrameSize < minExtendedFrameSize {
		t.Errorf("defaultFrameSize = %d, below the %d-byte extended-format minimum",
			defaultFrameSize, minExtendedFrameSize)
	}
}
