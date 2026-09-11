// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"flag"
	"io"
	"strings"
	"testing"
)

// CT307 reflects NetAlly traffic on UDP 3842 in mac-ip mode; that is the
// lab's acceptance condition, so it is pinned at the settings the CLI sends
// to the daemon rather than at a local dataplane config it no longer builds.
func TestNetAllyProfilePreservesUDPPorts(t *testing.T) {
	if mode := getReflectionMode("netally"); mode != "mac-ip" {
		t.Fatalf("NetAlly reflection mode = %q, want mac-ip", mode)
	}
	if port := getReflectorPort("netally", 0); port != 3842 {
		t.Fatalf("NetAlly reflector port = %d, want 3842", port)
	}
}

func TestParseReflectFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want reflectCmdArgs
	}{
		{
			name: "defaults",
			args: nil,
			want: reflectCmdArgs{profile: DefaultProfile},
		},
		{
			name: "long spellings",
			args: []string{
				"--interface",
				"eth0",
				"--profile",
				"netally",
				"--port",
				"3842",
				"--oui",
				"00:c0:17",
			},
			want: reflectCmdArgs{iface: "eth0", profile: "netally", port: 3842, oui: "00:c0:17"},
		},
		{
			name: "interface shorthand",
			args: []string{"-i", "eth1"},
			want: reflectCmdArgs{iface: "eth1", profile: DefaultProfile},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := parseReflectFlags(tt.args)
			if err != nil {
				t.Fatalf("parseReflectFlags(%v) error: %v", tt.args, err)
			}
			if *got != tt.want {
				t.Errorf("parseReflectFlags(%v) = %+v, want %+v", tt.args, *got, tt.want)
			}
		})
	}
}

// TestParseReflectFlagsRejectsAnOutOfRangePort: --port is parsed as a uint and
// narrowed to uint16, so a value above 65535 has to be refused rather than
// wrapped into a different port.
func TestParseReflectFlagsRejectsAnOutOfRangePort(t *testing.T) {
	_, _, err := parseReflectFlags([]string{"-i", "eth0", "--port", "70000"})
	if err == nil {
		t.Fatal("parseReflectFlags(--port 70000) = nil error")
	}
	if !strings.Contains(err.Error(), "70000") {
		t.Errorf("error %v does not name the rejected port", err)
	}
}

func TestRequireReflectInterface(t *testing.T) {
	fs := flag.NewFlagSet(subReflect, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	if err := requireReflectInterface("eth0", fs); err != nil {
		t.Errorf("requireReflectInterface(eth0) = %v, want nil", err)
	}

	var err error
	out := captureStdout(t, func() { err = requireReflectInterface("", fs) })
	if err == nil {
		t.Fatal("requireReflectInterface(\"\") = nil error")
	}
	if !strings.Contains(out, "--interface is required") {
		t.Errorf("output %q does not say the interface is required", out)
	}
}

// TestGetReflectorPortPrefersTheRequestedPort: --port overrides the profile's
// port, and 0 means "whatever the profile says".
func TestGetReflectorPortPrefersTheRequestedPort(t *testing.T) {
	if got := getReflectorPort("netally", 9999); got != 9999 {
		t.Errorf("getReflectorPort(netally, 9999) = %d, want 9999", got)
	}
	if got := getReflectorPort("netally", 0); got != 3842 {
		t.Errorf("getReflectorPort(netally, 0) = %d, want the profile's 3842", got)
	}
}

func TestPrintReflectorStartup(t *testing.T) {
	out := captureStdout(t, func() {
		printReflectorStartup(
			&reflectCmdArgs{iface: "eth0", profile: "netally", port: 3842, oui: "00:c0:17"},
		)
	})

	for _, want := range []string{"Interface:  eth0", "Profile:    netally", "Port:       3842", "OUI:        00:c0:17"} {
		if !strings.Contains(out, want) {
			t.Errorf("startup banner is missing %q:\n%s", want, out)
		}
	}
}

// TestPrintReflectorStartupOmitsUnsetOptionalFields: an unset port or OUI is
// left out rather than printed as a zero value.
func TestPrintReflectorStartupOmitsUnsetOptionalFields(t *testing.T) {
	out := captureStdout(t, func() {
		printReflectorStartup(&reflectCmdArgs{iface: "eth0", profile: DefaultProfile})
	})

	if strings.Contains(out, "Port:") {
		t.Errorf("startup banner printed a port that was never set:\n%s", out)
	}
	if strings.Contains(out, "OUI:") {
		t.Errorf("startup banner printed an OUI that was never set:\n%s", out)
	}
}
