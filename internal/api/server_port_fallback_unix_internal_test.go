// SPDX-License-Identifier: BUSL-1.1

//go:build !windows

package api

// server_port_fallback_unix_internal_test.go exercises the unix build of
// isAddrInUse directly. The filename uses the `_internal_test.go` suffix so
// the testpackage linter accepts the internal-package test (matches the
// configured skip-regex `(export|internal)_test\.go`); the `unix` component
// is decorative, since only the `//go:build` line above actually gates it.

import (
	"net"
	"syscall"
	"testing"
)

// TestIsAddrInUse_RecognisesSyscall confirms isAddrInUse matches a wrapped
// EADDRINUSE via [errors.Is], the path every unix target relies on.
func TestIsAddrInUse_RecognisesSyscall(t *testing.T) {
	wrapped := &net.OpError{Op: "listen", Err: syscall.EADDRINUSE}
	if !isAddrInUse(wrapped) {
		t.Fatalf("expected isAddrInUse to match EADDRINUSE")
	}
}
