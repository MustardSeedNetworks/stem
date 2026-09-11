// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// licenseHome points the license manager at a temp HOME so a case decides what
// is on disk, and returns the config directory it will read.
func licenseHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "stem")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	return dir
}

// parseTestFlagsOrFail parses args the way `stem test` does and fails the test
// if the flags the case supplies do not parse at all.
func parseTestFlagsOrFail(t *testing.T, args ...string) *testCmdFlags {
	t.Helper()

	flags, err := parseTestFlags(args)
	if err != nil {
		t.Fatalf("parseTestFlags(%v) error: %v", args, err)
	}
	return flags
}

// TestParseTestFlagsLongSpellings covers the flags an operator types in full.
func TestParseTestFlagsLongSpellings(t *testing.T) {
	f := parseTestFlagsOrFail(t,
		"--interface", "eth0", "--type", "rfc2544_throughput",
		"--duration", "30", "--frame-sizes", "64,1518")

	if f.iface != "eth0" || f.testTypes != "rfc2544_throughput" {
		t.Errorf("iface/type = %q/%q, want eth0/rfc2544_throughput", f.iface, f.testTypes)
	}
	if f.duration != 30 || f.frameSizes != "64,1518" {
		t.Errorf("duration/frameSizes = %d/%q, want 30/64,1518", f.duration, f.frameSizes)
	}
}

// TestParseTestFlagsShortSpellingsReachTheSameFields: -i/-t/-d are declared as
// separate flags bound to the same variables, which is where an alias silently
// stops working.
func TestParseTestFlagsShortSpellingsReachTheSameFields(t *testing.T) {
	f := parseTestFlagsOrFail(t, "-i", "eth1", "-t", "y1564", "-d", "5")

	if f.iface != "eth1" || f.testTypes != "y1564" || f.duration != 5 {
		t.Errorf("got %q/%q/%d, want eth1/y1564/5", f.iface, f.testTypes, f.duration)
	}
}

func TestParseTestFlagsDefaults(t *testing.T) {
	f := parseTestFlagsOrFail(t, "-i", "eth0")

	if f.testTypes != "rfc2544_throughput" {
		t.Errorf("testTypes = %q, want rfc2544_throughput", f.testTypes)
	}
	if f.frameSizes != "64,128,256,512,1024,1280,1518" {
		t.Errorf("default frame sizes = %q, want the RFC 2544 standard sizes", f.frameSizes)
	}
	if f.duration != defaultTestDuration || f.warmup != defaultWarmup {
		t.Errorf("duration/warmup = %d/%d, want %d/%d",
			f.duration, f.warmup, defaultTestDuration, defaultWarmup)
	}
	if f.resolution != defaultResolution || f.maxLoss != defaultMaxLoss {
		t.Errorf("resolution/maxLoss = %v/%v, want %v/%v",
			f.resolution, f.maxLoss, defaultResolution, defaultMaxLoss)
	}
	if f.jsonOutput || f.csvOutput {
		t.Error("json/csv output should default to off")
	}
}

// TestParseTestFlagsY1564Defaults: the three Y.1564 verdict thresholds decide
// pass or fail, so their defaults are part of the product's contract.
func TestParseTestFlagsY1564Defaults(t *testing.T) {
	f := parseTestFlagsOrFail(t, "-i", "eth0")

	if f.fdThreshold != defaultFDThreshold {
		t.Errorf("fdThreshold = %v, want %v", f.fdThreshold, defaultFDThreshold)
	}
	if f.fdvThreshold != defaultFDVThreshold {
		t.Errorf("fdvThreshold = %v, want %v", f.fdvThreshold, defaultFDVThreshold)
	}
	if f.flrThreshold != defaultFLRThreshold {
		t.Errorf("flrThreshold = %v, want %v", f.flrThreshold, defaultFLRThreshold)
	}
	if f.cir != 0 || f.eir != 0 {
		t.Errorf("cir/eir = %v/%v, want 0/0 until the operator supplies them", f.cir, f.eir)
	}
}

func TestParseTestFlagsY1564ServiceParameters(t *testing.T) {
	f := parseTestFlagsOrFail(t,
		"-i", "eth0", "--cir", "100", "--eir", "50",
		"--fd-threshold", "8", "--fdv-threshold", "3", "--flr-threshold", "0.05")

	if f.cir != 100 || f.eir != 50 {
		t.Errorf("cir/eir = %v/%v, want 100/50", f.cir, f.eir)
	}
	if f.fdThreshold != 8 || f.fdvThreshold != 3 || f.flrThreshold != 0.05 {
		t.Errorf("thresholds = %v/%v/%v, want 8/3/0.05",
			f.fdThreshold, f.fdvThreshold, f.flrThreshold)
	}
}

func TestParseTestFlagsOutputFormatsAndTuning(t *testing.T) {
	f := parseTestFlagsOrFail(t, "-i", "eth0", "--json", "--csv",
		"--resolution", "0.5", "--max-loss", "1.5", "--warmup", "7")

	if !f.jsonOutput || !f.csvOutput {
		t.Error("--json and --csv should both be set")
	}
	if f.resolution != 0.5 || f.maxLoss != 1.5 || f.warmup != 7 {
		t.Errorf("got %v/%v/%d, want 0.5/1.5/7", f.resolution, f.maxLoss, f.warmup)
	}
}

func TestParseTestFlagsRejectsUnknownFlags(t *testing.T) {
	_, err := parseTestFlags([]string{"--not-a-flag"})
	if err == nil {
		t.Fatal("parseTestFlags(--not-a-flag) = nil error, want a parse error")
	}
	if !strings.Contains(err.Error(), "failed to parse test flags") {
		t.Errorf("error = %v, want it wrapped as a test-flag parse failure", err)
	}
}

func TestParseTestFlagsHelpIsNotAnError(t *testing.T) {
	_, err := parseTestFlags([]string{"-h"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Errorf("parseTestFlags(-h) error = %v, want flag.ErrHelp", err)
	}
}

// TestValidateTestTypesList accepts exactly what the module registry
// registers; anything else is refused before the dataplane is opened.
func TestValidateTestTypesList(t *testing.T) {
	tests := []struct {
		name  string
		tests []string
		want  bool
	}{
		{"single registered type", []string{"rfc2544_throughput"}, true},
		{"a full RFC 2544 suite", []string{
			"rfc2544_throughput", "rfc2544_latency",
			"rfc2544_frame_loss", "rfc2544_back_to_back",
		}, true},
		{"y1564 service test", []string{"y1564"}, true},
		{"empty list is vacuously valid", nil, true},
		{"unregistered short name", []string{"throughput"}, false},
		{"one bad name spoils the list", []string{"rfc2544_throughput", "nonsense"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			out := captureStdout(t, func() { got = validateTestTypesList(tt.tests) })
			if got != tt.want {
				t.Errorf("validateTestTypesList(%v) = %v, want %v", tt.tests, got, tt.want)
			}
			if !tt.want && !strings.Contains(out, "stem list-tests") {
				t.Errorf("a refusal must point at 'stem list-tests'; got %q", out)
			}
		})
	}
}

// TestPrintTestConfigurationEchoesWhatWillRun is the banner an operator
// screenshots into a report, so every parameter has to appear.
func TestPrintTestConfigurationEchoesWhatWillRun(t *testing.T) {
	out := captureStdout(t, func() {
		printTestConfiguration("eth0", "rfc2544_throughput", "64,1518", 30, 0.1, 0.5, 2)
	})

	for _, want := range []string{
		"Interface:    eth0",
		"Tests:        rfc2544_throughput",
		"Duration:     30 seconds",
		"Frame sizes:  64,1518",
		"Resolution:   0.10%",
		"Max loss:     0.50%",
		"Warmup:       2 seconds",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("configuration banner is missing %q:\n%s", want, out)
		}
	}
}

// TestTestCmdRequiresAnInterface: the dataplane cannot be opened without one,
// so the CLI refuses before it tries.
func TestTestCmdRequiresAnInterface(t *testing.T) {
	var err error
	out := captureStdout(t, func() { err = testCmd([]string{"-t", "rfc2544_throughput"}) })

	if err == nil {
		t.Fatal("testCmd without --interface = nil error")
	}
	if !strings.Contains(out, "--interface is required") {
		t.Errorf("testCmd printed %q, which does not say the interface is required", out)
	}
}

// TestTestCmdRefusesUnknownTestTypesBeforeOpeningTheDataplane: validation runs
// first, so this case never needs a real NIC.
func TestTestCmdRefusesUnknownTestTypes(t *testing.T) {
	var err error
	out := captureStdout(t, func() { err = testCmd([]string{"-i", "eth0", "-t", "nonsense"}) })

	if err == nil {
		t.Fatal("testCmd with an unknown test type = nil error")
	}
	if !strings.Contains(out, "Unknown test type 'nonsense'") {
		t.Errorf("testCmd printed %q, which does not name the rejected type", out)
	}
}

// buildStartRequest is where a mistyped mapping silently changes what the
// operator measured — the run plan the daemon executes comes from here.
func TestBuildStartRequestCarriesTheFlagsItIsGiven(t *testing.T) {
	req := buildStartRequest(&testCmdFlags{
		iface:        "eth3",
		peer:         "192.0.2.10",
		peerPort:     4842,
		duration:     45,
		warmup:       9,
		resolution:   0.25,
		maxLoss:      2.5,
		cir:          100,
		eir:          20,
		fdThreshold:  3.5,
		fdvThreshold: 1.25,
		flrThreshold: 0.01,
	}, []string{"rfc2544_throughput"}, []uint32{64, 1518}, 45)

	if req.Interface != "eth3" {
		t.Errorf("Interface = %q, want eth3", req.Interface)
	}
	if req.Peer != "192.0.2.10" || req.PeerPort != 4842 {
		t.Errorf("Peer = %s:%d, want 192.0.2.10:4842", req.Peer, req.PeerPort)
	}
	if len(req.Tests) != 1 || req.Tests[0].TestType != "rfc2544_throughput" {
		t.Fatalf("Tests = %+v, want one rfc2544_throughput step", req.Tests)
	}

	rfc := req.Tests[0].Config.RFC2544
	if rfc.Duration != 45 || rfc.Warmup != 9 {
		t.Errorf("RFC2544 duration/warmup = %d/%d, want 45/9", rfc.Duration, rfc.Warmup)
	}
	if rfc.Resolution != 0.25 || rfc.MaxLoss != 2.5 {
		t.Errorf("RFC2544 resolution/maxLoss = %v/%v, want 0.25/2.5", rfc.Resolution, rfc.MaxLoss)
	}
	if len(rfc.FrameSizes) != 2 || rfc.FrameSizes[0] != 64 || rfc.FrameSizes[1] != 1518 {
		t.Errorf("RFC2544 frame sizes = %v, want [64 1518]", rfc.FrameSizes)
	}

	y := req.Tests[0].Config.Y1564
	if y.CIR != 100 || y.EIR != 20 {
		t.Errorf("Y.1564 CIR/EIR = %v/%v, want 100/20", y.CIR, y.EIR)
	}
	if y.FDThreshold != 3.5 || y.FDVThreshold != 1.25 || y.FLRThreshold != 0.01 {
		t.Errorf("Y.1564 thresholds = %v/%v/%v, want 3.5/1.25/0.01", y.FDThreshold, y.FDVThreshold, y.FLRThreshold)
	}
	if y.ConfigStepDuration != 45 || y.PerfTestDuration != 45 {
		t.Errorf("Y.1564 durations = %d/%d, want 45/45", y.ConfigStepDuration, y.PerfTestDuration)
	}
}

// A duration the wire cannot carry must be refused, not wrapped into a
// plausible-looking one: the operator would be handed numbers for a test
// that never ran for the time they asked for.
func TestValidateDurationBounds(t *testing.T) {
	for name, tc := range map[string]struct {
		duration int
		want     uint32
		wantErr  bool
	}{
		"one second": {duration: 1, want: 1},
		"typical":    {duration: 60, want: 60},
		"one day":    {duration: 86400, want: 86400},
		"zero":       {duration: 0, wantErr: true},
		"negative":   {duration: -1, wantErr: true},
		"over a day": {duration: 86401, wantErr: true},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := validateDuration(tc.duration)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("validateDuration(%d) = %d, nil; want an error", tc.duration, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("validateDuration(%d): %v", tc.duration, err)
			}
			if got != tc.want {
				t.Errorf("validateDuration(%d) = %d, want %d", tc.duration, got, tc.want)
			}
		})
	}
}
