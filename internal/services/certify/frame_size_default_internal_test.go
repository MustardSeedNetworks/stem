// SPDX-License-Identifier: BUSL-1.1

package certify

import "testing"

// minExtendedFrameSize mirrors api.MinExtendedFrameSize, which mirrors
// EXTENDED_MIN_FRAME_SIZE in include/rfc2544.h. Importing internal/api here
// would be an import cycle, so the value is restated and both copies are
// pinned to the header by TestExtendedFrameMinimumMatchesTheHeader.
const minExtendedFrameSize = 70

// A bare POST with no TSN config must not hand the dataplane a frame it
// rejects with -22: the default has to be legal on its own (stem#1250).
func TestDefaultTSNFrameSizeIsLegal(t *testing.T) {
	if defaultTSNFrameSize < minExtendedFrameSize {
		t.Errorf("defaultTSNFrameSize = %d, below the %d-byte extended-format minimum",
			defaultTSNFrameSize, minExtendedFrameSize)
	}
}
