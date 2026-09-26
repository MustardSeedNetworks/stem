// SPDX-License-Identifier: BUSL-1.1

package dataplane

import "errors"

var (
	// ErrThroughputNoIterations reports a search that never called run_trial,
	// so no frame was transmitted (#1217).
	ErrThroughputNoIterations = errors.New(
		"throughput test ran no iterations: no frames were transmitted")
	// ErrThroughputNoFramesReturned reports a search whose trials transmitted
	// and got nothing back: the peer is not reflecting test traffic (#1466).
	ErrThroughputNoFramesReturned = errors.New(
		"throughput test received no frames back from the peer")
	// ErrThroughputNoPassingRate reports a search whose peer answered but lost
	// more than the acceptable loss at every rate it tried (#1466).
	ErrThroughputNoPassingRate = errors.New(
		"throughput test found no rate within the acceptable loss")
)

// throughputVerdict decides whether a completed search measured a throughput.
// A zero result is worse than failing, because an operator cannot tell it from
// a link that genuinely carries nothing. The search only moves OfferedRatePct
// off zero at a trial that met the loss criterion, so zero after at least one
// trial means none did.
func throughputVerdict(r ThroughputResult) error {
	switch {
	case r.Iterations == 0:
		return ErrThroughputNoIterations
	case r.OfferedRatePct > 0:
		return nil
	case r.FramesReceived == 0:
		return ErrThroughputNoFramesReturned
	default:
		return ErrThroughputNoPassingRate
	}
}
