// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/services"
	"github.com/MustardSeedNetworks/stem/internal/version"
)

// captureOutput captures output during function execution and returns the output.
func captureOutput(t *testing.T, fn func(w io.Writer)) string {
	t.Helper()

	var buf bytes.Buffer
	fn(&buf)
	return buf.String()
}

// captureStdout runs fn with [os.Stdout] redirected to a pipe and returns what
// it wrote. Most of this package's print helpers address [os.Stdout] directly
// rather than taking an [io.Writer], so this is how their output is asserted.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	saved := os.Stdout
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	os.Stdout = saved
	if closeErr := w.Close(); closeErr != nil {
		t.Fatalf("close pipe: %v", closeErr)
	}
	out := <-done
	if closeErr := r.Close(); closeErr != nil {
		t.Fatalf("close pipe read end: %v", closeErr)
	}
	return out
}

func TestVersion(t *testing.T) {
	if version.GetVersion() == "" {
		t.Error("Version should not be empty")
	}
	// Version is "dev" when not built with ldflags, or semver when built.
	if version.GetVersion() != "dev" && !strings.Contains(version.GetVersion(), ".") {
		t.Error("Version should be 'dev' or contain dots (semantic versioning)")
	}
}

func TestProductName(t *testing.T) {
	if ProductName != "The Stem" {
		t.Errorf("Expected ProductName 'The Stem', got '%s'", ProductName)
	}
}

func TestCompany(t *testing.T) {
	if Company != "Mustard Seed Networks" {
		t.Errorf("Expected Company 'Mustard Seed Networks', got '%s'", Company)
	}
}

// exampleTestTypes extracts every test type named by a `-t` argument in the
// EXAMPLES block of printUsage.
func exampleTestTypes(t *testing.T) []string {
	t.Helper()

	usage := captureOutput(t, printUsage)
	start := strings.Index(usage, "EXAMPLES:")
	if start < 0 {
		t.Fatal("printUsage has no EXAMPLES block")
	}
	examples := usage[start:]
	re := regexp.MustCompile(`-t ([\w,]+)`)

	var names []string
	for _, m := range re.FindAllStringSubmatch(examples, -1) {
		names = append(names, strings.Split(m[1], ",")...)
	}
	return names
}

// TestUsageExamplesNameRegisteredTestTypes ties the documented invocations to
// the module registry the CLI validates against. An example naming a type no
// module registers is an instruction that cannot work (#1093).
func TestUsageExamplesNameRegisteredTestTypes(t *testing.T) {
	names := exampleTestTypes(t)
	if len(names) == 0 {
		t.Fatal("no -t examples found in printUsage; the extraction is broken, not the docs")
	}

	for _, name := range names {
		if services.GetModuleForTest(name) == nil {
			t.Errorf("EXAMPLES documents 'stem test -t %s', which no module registers", name)
		}
	}
}

func TestParseFrameSizes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []uint32
	}{
		{"single", "64", []uint32{64}},
		{"three", "64,128,256", []uint32{64, 128, 256}},
		{
			"rfc2544 defaults",
			"64,128,256,512,1024,1280,1518",
			[]uint32{64, 128, 256, 512, 1024, 1280, 1518},
		},
		{"jumbo", "1518,9000", []uint32{1518, 9000}},
		{"jumbo limit", "9000,9216", []uint32{9000, 9216}},
		{"empty", "", []uint32{}},
		{"spaces only", "  ", []uint32{}},
		{"padded", " 64 , 128 ", []uint32{64, 128}},
		{"trailing comma", "64,128,", []uint32{64, 128}},
		{"leading comma", ",64,128", []uint32{64, 128}},
		{"duplicates preserved", "64,64,128", []uint32{64, 64, 128}},
		{"non-numeric dropped", "64,invalid,128", []uint32{64, 128}},
		{"below minimum dropped", "63,64", []uint32{64}},
		{"above maximum dropped", "9216,9217", []uint32{9216}},
		{"far above maximum dropped", "16384", []uint32{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFrameSizes(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("parseFrameSizes(%q) = %v, want %v", tt.input, result, tt.expected)
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf(
						"parseFrameSizes(%q)[%d] = %d, want %d",
						tt.input,
						i,
						v,
						tt.expected[i],
					)
				}
			}
		})
	}
}

func TestPrintVersion(t *testing.T) {
	output := captureOutput(t, printVersion)

	if !strings.Contains(output, ProductName) {
		t.Error("printVersion should contain ProductName")
	}
	if !strings.Contains(output, version.GetVersion()) {
		t.Error("printVersion should contain Version")
	}
	if !strings.Contains(output, Company) {
		t.Error("printVersion should contain Company")
	}
	if !strings.Contains(output, version.GetCommit()) {
		t.Error("printVersion should contain the commit")
	}
}

func TestPrintUsage(t *testing.T) {
	output := captureOutput(t, printUsage)

	// Every verb dispatchSubcommand accepts must be discoverable here.
	expectedSections := []string{
		"USAGE:", "COMMANDS:", "EXAMPLES:",
		subReflect, subTest, "web", "license",
		"list-tests", installCACommandName, "tutorial", "glossary", "version",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("printUsage should contain '%s'", section)
		}
	}
}

// TestDispatchSubcommandRejectsUnknownVerbs is what makes main() print usage
// and exit 1 rather than silently succeeding.
func TestDispatchSubcommandRejectsUnknownVerbs(t *testing.T) {
	for _, cmd := range []string{"", "throughput", "reflectt", "--bogus"} {
		t.Run(cmd, func(t *testing.T) {
			if dispatchSubcommand(cmd, nil) {
				t.Errorf("dispatchSubcommand(%q) = true, want false for an unknown verb", cmd)
			}
		})
	}
}

// TestDispatchSubcommandVersionAliases covers the three spellings that reach
// printVersion; each must actually print, not merely return true.
func TestDispatchSubcommandVersionAliases(t *testing.T) {
	for _, cmd := range []string{"version", "--version", "-v"} {
		t.Run(cmd, func(t *testing.T) {
			var handled bool
			out := captureStdout(t, func() { handled = dispatchSubcommand(cmd, nil) })
			if !handled {
				t.Fatalf("dispatchSubcommand(%q) = false, want true", cmd)
			}
			if !strings.Contains(out, version.GetVersion()) {
				t.Errorf("dispatchSubcommand(%q) printed %q, which omits the version", cmd, out)
			}
		})
	}
}

// TestDispatchSubcommandHelpAliasesPrintUsage: `help` with no topic falls back
// to the usage screen, and --help/-h are the same door.
func TestDispatchSubcommandHelpAliasesPrintUsage(t *testing.T) {
	for _, cmd := range []string{"help", "--help", "-h"} {
		t.Run(cmd, func(t *testing.T) {
			var handled bool
			out := captureStdout(t, func() { handled = dispatchSubcommand(cmd, nil) })
			if !handled {
				t.Fatalf("dispatchSubcommand(%q) = false, want true", cmd)
			}
			if !strings.Contains(out, "COMMANDS:") {
				t.Errorf(
					"dispatchSubcommand(%q) printed %q, which is not the usage screen",
					cmd,
					out,
				)
			}
		})
	}
}

// TestHelpCmdUnknownTopicListsTopics: an operator who mistypes a topic gets
// the topic list, not silence.
func TestHelpCmdUnknownTopicListsTopics(t *testing.T) {
	out := captureStdout(t, func() { helpCmd([]string{"no-such-topic"}) })

	if !strings.Contains(out, "No help found for 'no-such-topic'") {
		t.Errorf("helpCmd printed %q, which does not name the missing topic", out)
	}
	for _, hint := range []string{"Available help topics:", "stem glossary", "stem tutorial"} {
		if !strings.Contains(out, hint) {
			t.Errorf("helpCmd output is missing %q:\n%s", hint, out)
		}
	}
}

// TestListTestsCmdListsEveryRegisteredType: `stem list-tests` is the answer
// the CLI gives when it rejects a test type, so it must name every type the
// registry will accept.
func TestListTestsCmdListsEveryRegisteredType(t *testing.T) {
	out := captureStdout(t, func() { listTestsCmd(nil) })

	total := 0
	for _, mod := range services.GetAllModules() {
		if !strings.Contains(out, mod.DisplayName()) {
			t.Errorf("list-tests omits module %q", mod.DisplayName())
		}
		for _, testType := range mod.TestTypes() {
			if !strings.Contains(out, testType) {
				t.Errorf("list-tests omits test type %q", testType)
			}
			total++
		}
	}
	if !strings.Contains(out, "Total: "+strconv.Itoa(total)+" test types") {
		t.Errorf("list-tests did not report %d test types:\n%s", total, out)
	}
}

// Benchmark tests.
func BenchmarkParseFrameSizes(b *testing.B) {
	input := "64,128,256,512,1024,1280,1518"
	for b.Loop() {
		parseFrameSizes(input)
	}
}

// TestTutorialCmdUnknownIDFallsBackToTheList: a mistyped tutorial name must
// leave the operator with the list of real ones.
func TestTutorialCmdUnknownIDFallsBackToTheList(t *testing.T) {
	unknown := captureStdout(t, func() { tutorialCmd([]string{"no-such-tutorial"}) })
	if !strings.Contains(unknown, "Tutorial 'no-such-tutorial' not found") {
		t.Errorf("tutorialCmd printed %q, which does not name the missing tutorial", unknown)
	}

	listing := captureStdout(t, func() { tutorialCmd(nil) })
	if listing == "" {
		t.Fatal("tutorialCmd with no arguments printed nothing")
	}
	if !strings.Contains(unknown, listing) {
		t.Error("the not-found path does not fall back to the same tutorial list")
	}
}

// TestGlossaryCmdSearchAndLookup: --search narrows the glossary and an
// unknown term says so rather than printing an empty definition.
func TestGlossaryCmdSearchAndLookup(t *testing.T) {
	hit := captureStdout(t, func() { glossaryCmd([]string{"--search", "latency"}) })
	if !strings.Contains(hit, "Terms matching 'latency'") {
		t.Errorf("glossary --search latency printed %q", hit)
	}

	miss := captureStdout(t, func() { glossaryCmd([]string{"--search", "zzzznotaterm"}) })
	if !strings.Contains(miss, "No terms found matching 'zzzznotaterm'") {
		t.Errorf("glossary --search on a nonsense keyword printed %q", miss)
	}

	unknown := captureStdout(t, func() { glossaryCmd([]string{"zzzznotaterm"}) })
	if !strings.Contains(unknown, "not found in glossary") {
		t.Errorf("glossary on an unknown term printed %q", unknown)
	}
}
