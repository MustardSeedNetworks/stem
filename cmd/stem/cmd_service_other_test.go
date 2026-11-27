// SPDX-License-Identifier: BUSL-1.1

//go:build !windows

package main

import (
	"errors"
	"strings"
	"testing"
)

func TestServiceCmdRefusesOffWindows(t *testing.T) {
	for _, action := range []string{"install", "start", "stop", "status", "uninstall", "run"} {
		t.Run(action, func(t *testing.T) {
			var err error
			stderr := captureStderr(t, func() { err = serviceCmd([]string{action}) })
			if !errors.Is(err, errServiceWindowsOnly) {
				t.Fatalf("serviceCmd(%q) error = %v, want errServiceWindowsOnly", action, err)
			}
			if !strings.Contains(stderr, "Windows only") {
				t.Errorf("stderr = %q, want it to say the command is Windows only", stderr)
			}
		})
	}
}
