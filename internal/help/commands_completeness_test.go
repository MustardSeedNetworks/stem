package help_test

// commands_completeness_test.go — locks CLI help completeness in CI.
//
// stem uses Go stdlib `flag` (not cobra) and dispatches verbs from a switch
// in cmd/stem/main.go. The help entries live in internal/help. This test
// enforces three invariants:
//
//   1. Every verb in main.go's dispatch switch has a CommandHelp entry in
//      GetAllCommands(). install-ca / tui / list-tests were missing before
//      PR-B — this catches the next gap automatically.
//   2. Every CommandHelp has non-empty Name, Summary, Description, Usage,
//      and at least one Example.
//   3. Every FlagHelp has a non-empty TechDesc (the authoritative usage
//      string CI exposes to operators).

import (
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/help"
)

// canonicalVerbs lists the top-level CLI verbs cmd/stem/main.go dispatches
// (excluding "-h" / "--help" / "--version" synonyms). When a new verb lands
// in main.go's dispatch switch, add it here AND add a CommandHelp entry in
// GetAllCommands().
func canonicalVerbs() []string {
	return []string{
		"reflect", "test", "web", "tui", "license", "list-tests",
		"help", "tutorial", "glossary", "version", "install-ca",
	}
}

func TestEveryDispatchVerbHasHelpEntry(t *testing.T) {
	all := help.GetAllCommands()
	var missing []string
	for _, v := range canonicalVerbs() {
		if _, ok := all[v]; !ok {
			missing = append(missing, v)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("dispatch verbs missing help-package entries: %s",
			strings.Join(missing, ", "))
	}
}

func TestEveryHelpEntryHasMatchingDispatchVerb(t *testing.T) {
	known := map[string]bool{}
	for _, v := range canonicalVerbs() {
		known[v] = true
	}
	var stray []string
	for name := range help.GetAllCommands() {
		if !known[name] {
			stray = append(stray, name)
		}
	}
	if len(stray) > 0 {
		sort.Strings(stray)
		t.Fatalf("help entries with no dispatch verb (drift): %s",
			strings.Join(stray, ", "))
	}
}

func TestEveryHelpEntryIsComplete(t *testing.T) {
	var gaps []string
	for name, h := range help.GetAllCommands() {
		if h.Name == "" {
			gaps = append(gaps, name+": empty Name")
		}
		if h.Summary == "" {
			gaps = append(gaps, name+": empty Summary")
		}
		if h.Description == "" {
			gaps = append(gaps, name+": empty Description")
		}
		if h.Usage == "" {
			gaps = append(gaps, name+": empty Usage")
		}
		if len(h.Examples) == 0 {
			gaps = append(gaps, name+": no Examples")
		}
		for i, f := range h.Flags {
			if f.TechDesc == "" {
				gaps = append(gaps, name+": flag #"+strconv.Itoa(i)+" empty TechDesc")
			}
		}
	}
	if len(gaps) > 0 {
		sort.Strings(gaps)
		t.Fatalf("incomplete help entries:\n  %s", strings.Join(gaps, "\n  "))
	}
}
