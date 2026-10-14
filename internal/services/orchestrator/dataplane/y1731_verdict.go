// SPDX-License-Identifier: BUSL-1.1

package dataplane

import "fmt"

// y1731Verdict fails a Y.1731 test that put nothing on the wire or got
// nothing back. Each would otherwise pass with its 100 % loss, 0 %
// availability or zero delay reported as a measurement, when the peer is not
// answering at all (#1731). Loss short of that is reported as measured:
// Y.1731 takes no loss threshold.
func y1731Verdict(test string, sent, received uint64) error {
	switch {
	case sent == 0:
		return fmt.Errorf("Y.1731 %s test %w", test, ErrNoTrialsRun)
	case received == 0:
		return fmt.Errorf("Y.1731 %s test %w", test, ErrNoFramesReturned)
	default:
		return nil
	}
}
