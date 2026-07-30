//go:build cgo && linux

// SPDX-License-Identifier: BUSL-1.1

package dataplane

/*
#cgo CFLAGS: -I${SRCDIR}/../../../../include
#include "rfc2544.h"
#include <stdlib.h>
*/
import "C"

import "unsafe"

// NewTestContext creates a test context for unit tests that do not execute dataplane operations.
func NewTestContext() *Context {
	return &Context{}
}

// GetLineRate returns the interface line rate in bits/sec.
func GetLineRate(iface string) uint64 {
	cIface := C.CString(iface)
	defer C.free(unsafe.Pointer(cIface))
	return uint64(C.rfc2544_get_line_rate(cIface))
}

// CalcPPS calculates packets per second for a line rate and frame size.
func CalcPPS(lineRate uint64, frameSize uint32) uint64 {
	return uint64(C.rfc2544_calc_pps(C.uint64_t(lineRate), C.uint32_t(frameSize)))
}
