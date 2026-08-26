// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"errors"

	"golang.org/x/sys/windows"
)

// isAddrInUse reports whether err indicates the address-in-use condition,
// i.e. whether bindWithFallback should walk to the next port rather than
// treat the bind failure as fatal.
//
// Winsock reports that as WSAEADDRINUSE, which is not the value
// [syscall.EADDRINUSE] carries here: on Windows that constant is a
// synthesised APPLICATION_ERROR with nothing mapping the Winsock errno back
// to it, so comparing against it is always false and the walk would never
// happen.
func isAddrInUse(err error) bool {
	return errors.Is(err, windows.WSAEADDRINUSE)
}
