// SPDX-License-Identifier: BUSL-1.1

package dataplane

import "fmt"

// customStreamVerdict fails a traffic-generation stream that put nothing on
// the wire or got nothing back. Either would otherwise pass with its 100 %
// loss reported as a measurement, when the peer is not reflecting at all
// (#1563). Loss short of that is reported as measured: a raw stream has no
// loss threshold.
func customStreamVerdict(r TrafficGenResult) error {
	switch {
	case r.PacketsSent == 0:
		return fmt.Errorf("custom stream %w", ErrNoTrialsRun)
	case r.PacketsRecv == 0:
		return fmt.Errorf("custom stream %w", ErrNoFramesReturned)
	default:
		return nil
	}
}
