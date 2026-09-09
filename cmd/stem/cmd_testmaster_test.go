// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

// TestCheckTestLicenseRefusesAMalformedFile is the CLI half of #1068: a file
// the manager cannot parse used to be indistinguishable from a fresh install,
// so `stem test` started a Professional trial and overwrote the operator's
// license. It must refuse instead.
func TestCheckTestLicenseRefusesAMalformedFile(t *testing.T) {
	dir := licenseHome(t)
	if err := os.WriteFile(filepath.Join(dir, ".license"), []byte("not a licence"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	if checkTestLicense() {
		t.Error("checkTestLicense() = true on a malformed license file")
	}
	if _, err := os.ReadFile(filepath.Join(dir, ".license")); err != nil {
		t.Errorf("license file no longer readable after the check: %v", err)
	}
}

// TestCheckTestLicenseStartsTheTrialOnAFreshInstall is the control: refusing
// the damaged file must not have made every unlicensed host unable to test.
func TestCheckTestLicenseStartsTheTrialOnAFreshInstall(t *testing.T) {
	dir := licenseHome(t)

	if !checkTestLicense() {
		t.Fatal("checkTestLicense() = false on a fresh install; the trial should start")
	}
	if _, err := os.Stat(filepath.Join(dir, ".license")); err != nil {
		t.Errorf("no state written after starting the trial: %v", err)
	}
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

// TestCreateTestConfigMapsTheFlagsItIsGiven: the dataplane config is where a
// mistyped mapping silently changes what the operator measured.
func TestCreateTestConfigMapsTheFlagsItIsGiven(t *testing.T) {
	cfg := createTestConfig(&testCmdFlags{
		iface:      "eth3",
		duration:   45,
		warmup:     9,
		resolution: 0.25,
		maxLoss:    2.5,
	})

	if cfg.Interface != "eth3" {
		t.Errorf("Interface = %q, want eth3", cfg.Interface)
	}
	if cfg.TrialDuration != 45*time.Second {
		t.Errorf("TrialDuration = %v, want 45s", cfg.TrialDuration)
	}
	if cfg.WarmupPeriod != 9*time.Second {
		t.Errorf("WarmupPeriod = %v, want 9s", cfg.WarmupPeriod)
	}
	if cfg.ResolutionPct != 0.25 {
		t.Errorf("ResolutionPct = %v, want 0.25", cfg.ResolutionPct)
	}
	if cfg.AcceptableLoss != 2.5 {
		t.Errorf("AcceptableLoss = %v, want 2.5", cfg.AcceptableLoss)
	}
	if !cfg.AutoDetect {
		t.Error("AutoDetect should be on: the CLI never asks for a line rate")
	}
	if !cfg.MeasureLatency {
		t.Error("MeasureLatency should be on: the CLI prints latency with every throughput run")
	}
	if cfg.InitialRatePct != 100 || cfg.MaxIterations != 20 || cfg.BatchSize != 32 {
		t.Errorf("throughput defaults = %v%%/%d iterations/%d batch, want 100%%/20/32",
			cfg.InitialRatePct, cfg.MaxIterations, cfg.BatchSize)
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
