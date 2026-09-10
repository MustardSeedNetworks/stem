// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"bufio"
	"os"
	"os/exec"
	"sort"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/help"
)

func TestCommandHelpFlagsMatchRuntimeFlagSets(t *testing.T) {
	t.Parallel()

	for _, command := range []string{"reflect", "test", "web", "license", "help", "glossary", "list-tests", "install-ca"} {
		commandHelp := help.GetAllCommands()[command]
		t.Run(command, func(t *testing.T) {
			t.Parallel()

			process := exec.Command(os.Args[0], "-test.run=TestCLIHelpContractProcess")
			process.Env = append(os.Environ(), "STEM_CLI_HELP_COMMAND="+command)
			output, err := process.CombinedOutput()
			if err != nil {
				t.Fatalf("run %s --help: %v\n%s", command, err, output)
			}
			got := outputFlagNames(string(output))
			want := helpFlagNames(commandHelp.Flags)
			if strings.Join(got, ",") != strings.Join(want, ",") {
				t.Fatalf("runtime flags = %v, documented flags = %v", got, want)
			}
		})
	}
}

func TestCLIHelpContractProcess(_ *testing.T) {
	command := os.Getenv("STEM_CLI_HELP_COMMAND")
	if command == "" {
		return
	}

	dispatchSubcommand(command, []string{"--help"})
}

func outputFlagNames(output string) []string {
	names := make([]string, 0)
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "-") {
			continue
		}
		name := strings.TrimLeft(strings.Fields(line)[0], "-")
		if name != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func helpFlagNames(flags []help.FlagHelp) []string {
	names := make([]string, 0, len(flags)*2)
	for _, current := range flags {
		if current.Long != "" {
			names = append(names, strings.TrimPrefix(current.Long, "--"))
		}
		if current.Short != "" {
			names = append(names, strings.TrimPrefix(current.Short, "-"))
		}
	}
	sort.Strings(names)
	return names
}
