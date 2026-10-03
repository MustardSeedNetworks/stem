// SPDX-License-Identifier: BUSL-1.1

package dataplane

import (
	"errors"
	"fmt"
)

var (
	// ErrNoTrialsRun reports a test that never called run_trial, so no frame
	// was transmitted (#1217).
	ErrNoTrialsRun = errors.New("ran no trials: no frames were transmitted")
	// ErrNoFramesReturned reports a test whose trials transmitted and got
	// nothing back across the whole test: the peer is not reflecting test
	// traffic (#1466, #1275).
	ErrNoFramesReturned = errors.New("received no frames back from the peer")
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
		return fmt.Errorf("throughput test %w", ErrNoTrialsRun)
	case r.OfferedRatePct > 0:
		return nil
	case r.FramesReceived == 0:
		return fmt.Errorf("throughput test %w", ErrNoFramesReturned)
	default:
		return ErrThroughputNoPassingRate
	}
}

// latencyVerdict fails a latency test in which no load level timed a single
// returned frame: its zero latencies would read as an impossibly fast link.
// A level the peer answered keeps its measurement even when others did not.
func latencyVerdict(results []LatencyResultCLI) error {
	for _, r := range results {
		if r.Latency.Count > 0 {
			return nil
		}
	}
	return fmt.Errorf("latency test %w", ErrNoFramesReturned)
}

// frameLossVerdict fails a frame loss test the peer never answered. 100 % loss
// at every rate is not a loss measurement of the path; it is a silent peer.
// A peer that returned anything has its loss reported as measured.
func frameLossVerdict(points []FrameLossPoint) error {
	if len(points) == 0 {
		return fmt.Errorf("frame loss test %w", ErrNoTrialsRun)
	}
	for _, p := range points {
		if p.FramesRecv > 0 {
			return nil
		}
	}
	return fmt.Errorf("frame loss test %w", ErrNoFramesReturned)
}

// backToBackVerdict fails a back-to-back test the peer never answered, which
// would otherwise report a maximum burst of zero frames as a measurement.
func backToBackVerdict(r BurstResult) error {
	if r.FramesRecv == 0 {
		return fmt.Errorf("back-to-back test %w", ErrNoFramesReturned)
	}
	return nil
}
