// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
		{
			name: "tui flag",
			args: []string{"-i", "eth0", "--tui"},
			want: reflectCmdArgs{iface: "eth0", profile: DefaultProfile, useTUI: true},
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

// TestBuildReflectorConfigOUIFilter: the OUI filter is only switched on when
// the operator supplies one; the default OUI must not filter by itself.
func TestBuildReflectorConfigOUIFilter(t *testing.T) {
	off := buildReflectorConfig(&reflectCmdArgs{iface: "eth0", profile: DefaultProfile}, "")
	if off.Filtering.FilterOUI {
		t.Error("FilterOUI is on without --oui")
	}

	on := buildReflectorConfig(
		&reflectCmdArgs{iface: "eth0", profile: DefaultProfile, oui: "00:11:22"},
		"",
	)
	if !on.Filtering.FilterOUI {
		t.Error("FilterOUI is off with --oui set")
	}
	if on.Filtering.OUI != "00:11:22" {
		t.Errorf("OUI = %q, want 00:11:22", on.Filtering.OUI)
	}
}

func TestBuildReflectorConfigCarriesTheParsedFlags(t *testing.T) {
	cfg := buildReflectorConfig(
		&reflectCmdArgs{iface: "eth7", profile: "netally", port: 4000, useTUI: true},
		"rfc2544",
	)

	if cfg.Interface != "eth7" {
		t.Errorf("Interface = %q, want eth7", cfg.Interface)
	}
	if cfg.SignatureFilter != "rfc2544" {
		t.Errorf("SignatureFilter = %q, want rfc2544", cfg.SignatureFilter)
	}
	if cfg.Filtering.Port != 4000 {
		t.Errorf("Port = %d, want the requested 4000", cfg.Filtering.Port)
	}
	if !cfg.TUI.Enabled {
		t.Error("TUI.Enabled = false with --tui")
	}
	if cfg.WebUI.Enabled {
		t.Error("WebUI should stay off in reflect mode")
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

// TestReportReflectorLicenseNeverSpendsTheTrial is the reflector half of
// #1068: reflecting is the Free grant, so starting the reflector on an
// unlicensed host must report and continue, never start the 14 Professional
// days and never write to the license file.
func TestReportReflectorLicenseNeverSpendsTheTrial(t *testing.T) {
	dir := licenseHome(t)
	path := filepath.Join(dir, ".license")

	out := captureStdout(t, reportReflectorLicense)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("reportReflectorLicense wrote %s on a fresh install (stat err = %v)", path, err)
	}
	if strings.Contains(out, "trial") || strings.Contains(out, "Trial") {
		t.Errorf("reportReflectorLicense mentioned a trial on a fresh install: %q", out)
	}
}

// TestReportReflectorLicenseWarnsAboutADamagedFileAndContinues: a file the
// manager cannot parse is worth saying out loud, but it must not stop the
// reflector, which needs no license at all.
func TestReportReflectorLicenseWarnsAboutADamagedFileAndContinues(t *testing.T) {
	dir := licenseHome(t)
	path := filepath.Join(dir, ".license")
	if err := os.WriteFile(path, []byte("not a licence"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	out := captureStdout(t, reportReflectorLicense)

	if !strings.Contains(out, "Warning") {
		t.Errorf("a damaged license file produced no warning: %q", out)
	}
	if !strings.Contains(out, "which is free") {
		t.Errorf("the warning does not tell the operator the reflector runs anyway: %q", out)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "not a licence" {
		t.Errorf("reportReflectorLicense modified the operator's license file: %q, %v", data, err)
	}
}
