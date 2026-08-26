// SPDX-License-Identifier: BUSL-1.1

//go:build !windows

package api

import (
	"errors"
	"syscall"
)

// isAddrInUse reports whether err indicates the address-in-use condition,
// i.e. whether bindWithFallback should walk to the next port rather than
// treat the bind failure as fatal.
func isAddrInUse(err error) bool {
	return errors.Is(err, syscall.EADDRINUSE)
}
