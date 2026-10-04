// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/help"
	"github.com/MustardSeedNetworks/stem/internal/services"
)

// uiHelpDir holds the web UI's copy of the help topics and tutorials.
const uiHelpDir = "../../ui/src/data/help"

// helpTestExample is one documented `stem test` invocation and where it lives.
type helpTestExample struct {
	source  string
	command string
}

// helpTestExamples gathers documented `stem test` invocations by source.
type helpTestExamples []helpTestExample

func (e *helpTestExamples) add(source, command string) {
	if strings.HasPrefix(command, "stem test ") {
		*e = append(*e, helpTestExample{source: source, command: command})
	}
}

// addLines adds each line of prose that is itself an invocation.
func (e *helpTestExamples) addLines(source, text string) {
	for line := range strings.Lines(text) {
		e.add(source, strings.TrimSpace(line))
	}
}

func (e *helpTestExamples) addExamples(source string, examples []help.Example) {
	for _, example := range examples {
		e.add(source, example.Command)
	}
}

// goHelpTestExamples collects every `stem test` invocation the CLI help
// shows: topic, command and error examples, and tutorial step commands.
func goHelpTestExamples() helpTestExamples {
	var examples helpTestExamples
	for id, topic := range help.GetAllTests() {
		examples.addExamples("help topic "+id, topic.Examples)
		for _, issue := range topic.CommonIssues {
			examples.addLines("help topic "+id, issue.Solution)
		}
	}
	for name, command := range help.GetAllCommands() {
		examples.addExamples("help command "+name, command.Examples)
	}
	for code, errHelp := range help.GetAllErrors() {
		examples.addExamples("help error "+code, errHelp.Examples)
		examples.addLines("help error "+code, errHelp.Solution)
	}
	for id, tutorial := range help.GetAllTutorials() {
		for _, step := range tutorial.Steps {
			examples.add("tutorial "+id, step.Command)
			examples.addLines("tutorial "+id, step.Content)
		}
	}
	return examples
}

// uiHelpTestExamples collects every `command:` string in the web UI's help
// data that runs `stem test`, with or without sudo.
func uiHelpTestExamples(t *testing.T) helpTestExamples {
	t.Helper()

	files, err := filepath.Glob(filepath.Join(uiHelpDir, "*.ts"))
	if err != nil {
		t.Fatal(err)
	}
	nested, err := filepath.Glob(filepath.Join(uiHelpDir, "tests", "*.ts"))
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`command:\s*'(?:sudo )?(stem test [^']*)'`)

	var examples helpTestExamples
	for _, file := range append(files, nested...) {
		data, readErr := os.ReadFile(file)
		if readErr != nil {
			t.Fatal(readErr)
		}
		for _, m := range re.FindAllStringSubmatch(string(data), -1) {
			examples.add(file, m[1])
		}
	}
	return examples
}

// isRunFlag reports whether a flag selects the run rather than a test's
// configuration, so that every test type honours it.
func isRunFlag(arg string) bool {
	switch arg {
	case "-i", "--interface", "--peer", "--peer-port", "-t", "--type", "--json", "--csv":
		return true
	}
	return false
}

// exampleArgs returns the arguments after `stem test`, up to any shell
// redirect or pipe, and the configuration flags among them.
func exampleArgs(command string) ([]string, []string) {
	args := strings.Fields(command)[2:]
	for i, arg := range args {
		if arg == ">" || arg == "|" {
			args = args[:i]
			break
		}
	}
	var configFlags []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") && !isRunFlag(arg) {
			configFlags = append(configFlags, arg)
		}
	}
	return args, configFlags
}

// checkHelpExample applies to one example the checks testCmd applies before
// it contacts the daemon, and checks that every configuration flag reaches
// each test type the example runs.
func checkHelpExample(t *testing.T, example helpTestExample) {
	t.Helper()

	args, configFlags := exampleArgs(example.command)
	flags, err := parseTestFlags(args)
	if err != nil {
		t.Errorf("%s: %q does not parse: %v", example.source, example.command, err)
		return
	}
	if flags.iface == "" || flags.peer == "" {
		t.Errorf("%s: %q lacks -i or --peer", example.source, example.command)
	}
	seconds, err := validateDuration(flags.duration)
	if err != nil {
		t.Errorf("%s: %q: %v", example.source, example.command, err)
	}
	frameSizes := parseFrameSizes(flags.frameSizes)
	if len(frameSizes) == 0 {
		t.Errorf("%s: %q has no valid frame size", example.source, example.command)
	}
	for name := range strings.SplitSeq(flags.testTypes, ",") {
		switch {
		case services.GetModuleForTest(name) == nil:
			t.Errorf("%s: %q names test type %q, which no module registers",
				example.source, example.command, name)
		case len(configFlags) > 0 && stepConfig(flags, name, frameSizes, seconds) == nil:
			t.Errorf("%s: %q sets %v, which the CLI does not pass to %s",
				example.source, example.command, configFlags, name)
		}
	}
}

// TestHelpExamplesRunAsWritten holds every `stem test` example in the CLI and
// web UI help to the real flag set and module registry. An example an
// operator copies must not fail with "Unknown test type" or "flag provided
// but not defined", nor pass a setting the CLI drops for that type (#1435).
func TestHelpExamplesRunAsWritten(t *testing.T) {
	goExamples := goHelpTestExamples()
	uiExamples := uiHelpTestExamples(t)
	if len(goExamples) == 0 || len(uiExamples) == 0 {
		t.Fatalf("found %d CLI and %d UI help examples; the extraction is broken, not the docs",
			len(goExamples), len(uiExamples))
	}

	for _, example := range append(goExamples, uiExamples...) {
		checkHelpExample(t, example)
	}
}
