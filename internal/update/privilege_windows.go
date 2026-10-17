// SPDX-License-Identifier: BUSL-1.1

package update

import "golang.org/x/sys/windows"

// Privileged reports whether this process may replace the installed binary:
// an elevated Administrator token on Windows.
func Privileged() bool { return windows.GetCurrentProcessToken().IsElevated() }
