// SPDX-License-Identifier: BUSL-1.1

package main

import "testing"

func TestNetAllyProfilePreservesUDPPorts(t *testing.T) {
	parsed := &reflectCmdArgs{iface: "eth0", profile: "netally"}

	cfg := buildReflectorConfig(parsed, getSignatureFilter(parsed.profile))

	if cfg.Reflection.Mode != "mac-ip" {
		t.Fatalf("NetAlly reflection mode = %q, want mac-ip", cfg.Reflection.Mode)
	}
	if cfg.Filtering.Port != 3842 {
		t.Fatalf("NetAlly reflector port = %d, want 3842", cfg.Filtering.Port)
	}
}
