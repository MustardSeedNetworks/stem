// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"errors"
	"strings"
	"testing"
)

func TestServiceCmdArguments(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "no action", args: nil, wantErr: "expected one action"},
		{name: "two actions", args: []string{"start", "stop"}, wantErr: "expected one action"},
		{name: "unknown action", args: []string{"restart"}, wantErr: `unknown action "restart"`},
		{name: "unknown flag", args: []string{"--force", "start"}, wantErr: "flag provided but not defined"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			captureStderr(t, func() { err = serviceCmd(tt.args) })
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("serviceCmd(%v) error = %v, want one containing %q", tt.args, err, tt.wantErr)
			}
			if errors.Is(err, errServiceWindowsOnly) {
				t.Errorf("serviceCmd(%v) reached the platform layer with a bad argument", tt.args)
			}
		})
	}
}

func TestServiceCmdHelpSucceeds(t *testing.T) {
	var err error
	stderr := captureStderr(t, func() { err = serviceCmd([]string{"--help"}) })
	if err != nil {
		t.Fatalf("serviceCmd(--help) = %v, want nil", err)
	}
	if !strings.Contains(stderr, "stem service install|start|stop|status|uninstall") {
		t.Errorf("help output = %q, want the usage line", stderr)
	}
}
