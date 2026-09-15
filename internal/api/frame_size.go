// SPDX-License-Identifier: BUSL-1.1

package api

import "fmt"

// MinExtendedFrameSize is the smallest frame the extended payload formats
// (TSN, TrafficGen and Y.1564) can carry. They use a 24-byte payload where the
// RFC 2544 measurement format uses 18, so the payload runs to offset 66 and a
// 70-byte frame is the first that holds one once the NIC's four-byte FCS is
// counted (ADR 0008).
//
// The single definition is EXTENDED_MIN_FRAME_SIZE in include/rfc2544.h;
// TestExtendedFrameMinimumMatchesTheHeader fails if the two drift apart.
const MinExtendedFrameSize = 70

// validateFrameSizes rejects a step whose configured frame size is below the
// minimum its payload format needs. Without this the dataplane returns a bare
// -22 from deep inside the trial (stem#1250).
func validateFrameSizes(step RunPlanStep) error {
	if step.Config == nil {
		return nil
	}
	if tsn := step.Config.TSN; tsn != nil {
		if err := checkFrameSize("TSN", tsn.FrameSize); err != nil {
			return err
		}
	}
	if gen := step.Config.TrafficGen; gen != nil {
		if err := checkFrameSize("TrafficGen", gen.FrameSize); err != nil {
			return err
		}
	}
	if y1564 := step.Config.Y1564; y1564 != nil {
		for _, size := range y1564.FrameSizes {
			if err := checkFrameSize("Y.1564", size); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkFrameSize treats 0 as "unset": the module applies its own default,
// which the defaults tests hold at or above the minimum.
func checkFrameSize(format string, size uint32) error {
	if size == 0 || size >= MinExtendedFrameSize {
		return nil
	}
	return fmt.Errorf(
		"%s frame size %d is below the %d-byte minimum for its payload format",
		format, size, MinExtendedFrameSize)
}
