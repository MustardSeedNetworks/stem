// SPDX-License-Identifier: BUSL-1.1

//go:build unix

package update

import "os"

// Privileged reports whether this process may replace the installed binary
// and restart the service: root on Unix.
func Privileged() bool { return os.Geteuid() == 0 }
