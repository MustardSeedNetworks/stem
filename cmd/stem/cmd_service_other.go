// SPDX-License-Identifier: BUSL-1.1

//go:build !windows

package main

import "io"

// runServiceAction refuses every action off Windows: the other platforms
// ship a systemd unit or a launchd plist instead.
func runServiceAction(string, io.Writer) error {
	return errServiceWindowsOnly
}
