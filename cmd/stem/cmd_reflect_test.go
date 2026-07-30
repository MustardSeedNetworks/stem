// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"testing"

	reflectorConfig "github.com/MustardSeedNetworks/stem/internal/reflector/config"
)

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

func TestBuildTUIReflectorConfigSuppliesAFPacketGuardPort(t *testing.T) {
	cfg := buildTUIReflectorConfig("eth0")
	if cfg.Filtering.Port != reflectorConfig.NetAllyPort {
		t.Fatalf("TUI reflector port = %d, want %d", cfg.Filtering.Port, reflectorConfig.NetAllyPort)
	}
}
