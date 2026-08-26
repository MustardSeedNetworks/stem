// SPDX-License-Identifier: BUSL-1.1

//go:build windows

package api

// server_port_fallback_windows_internal_test.go exercises the Windows build
// of isAddrInUse directly. The filename uses the `_internal_test.go` suffix
// so the testpackage linter accepts the internal-package test (matches the
// configured skip-regex `(export|internal)_test\.go`); the `windows`
// component is decorative, since only the `//go:build` line above actually
// gates it.

import (
	"net"
	"testing"

	"golang.org/x/sys/windows"
)

// TestIsAddrInUse_RecognisesWinsock confirms isAddrInUse matches a wrapped
// WSAEADDRINUSE, the errno Winsock actually reports — [syscall.EADDRINUSE]
// never appears on the wire here, so a test built on it would pass while
// the walk it's meant to prove stayed dead.
func TestIsAddrInUse_RecognisesWinsock(t *testing.T) {
	wrapped := &net.OpError{Op: "listen", Err: windows.WSAEADDRINUSE}
	if !isAddrInUse(wrapped) {
		t.Fatalf("expected isAddrInUse to match WSAEADDRINUSE")
	}
}
