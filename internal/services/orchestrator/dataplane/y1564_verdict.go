// SPDX-License-Identifier: BUSL-1.1

package dataplane

import "errors"

// ErrY1564NothingTransmitted reports a Y.1564 step that put no frame on the
// wire, so the service was never measured (#1482). The dataplane fails such a
// step; this names why, since its zero loss and delay read like a clean link.
var ErrY1564NothingTransmitted = errors.New(
	"a Y.1564 step transmitted no frames, so the service was not measured")

// NothingTransmitted reports whether any step of the test sent no frame.
func (r *Y1564ConfigResult) NothingTransmitted() bool {
	for _, step := range r.Steps {
		if step.FramesTx == 0 {
			return true
		}
	}
	return false
}

// NothingTransmitted reports whether the test sent no frame.
func (r *Y1564PerfResult) NothingTransmitted() bool {
	return r.FramesTx == 0
}
