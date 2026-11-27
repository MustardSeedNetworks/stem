// SPDX-License-Identifier: BUSL-1.1

//go:build !windows

package update

// newServiceController is nil off Windows: the other install methods own
// their service through systemd or launchd.
func newServiceController() serviceController { return nil }
