// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/MustardSeedNetworks/stem/internal/help"
	"github.com/MustardSeedNetworks/stem/internal/services"
)

func listTestsCmd(args []string) {
	fs := flag.NewFlagSet("list-tests", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output in JSON format")

	err := fs.Parse(args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		// JSON output using modules.
		moduleInfos := services.GetAllModuleInfos()
		data, _ := json.MarshalIndent(map[string]any{
			"modules": moduleInfos,
			"count":   len(moduleInfos),
		}, "", "  ")
		_, _ = fmt.Fprintln(os.Stdout, string(data))
		return
	}

	_, _ = fmt.Fprintf(os.Stdout, "%s - Available Test Types by Module\n", ProductName)
	_, _ = fmt.Fprintln(os.Stdout, strings.Repeat("=", bannerWidth))
	hs := help.NewSystem()
	allMods := services.GetAllModules()
	totalTests := 0

	for _, mod := range allMods {
		_, _ = fmt.Fprintf(
			os.Stdout,
			"\n%s [%s] (%s):\n",
			mod.DisplayName(),
			mod.Color(),
			mod.Standard(),
		)
		_, _ = fmt.Fprintf(os.Stdout, "  %s\n", mod.Description())
		_, _ = fmt.Fprintln(os.Stdout)
		for _, t := range mod.TestTypes() {
			desc := ""
			if test, ok := hs.Tests[t]; ok {
				desc = test.Summary
			}
			if desc == "" {
				_, _ = fmt.Fprintf(os.Stdout, "    %-20s\n", t)
			} else {
				_, _ = fmt.Fprintf(os.Stdout, "    %-20s %s\n", t, desc)
			}
			totalTests++
		}
	}
	_, _ = fmt.Fprintf(
		os.Stdout,
		"\nTotal: %d test types across %d modules\n",
		totalTests,
		len(allMods),
	)
}

func helpCmd(args []string) {
	fs := flag.NewFlagSet("help", flag.ExitOnError)
	simple := fs.Bool("simple", false, "Show simplified explanations for non-technical users")
	fs.BoolVar(simple, "s", false, "Show simplified explanations (shorthand)")

	err := fs.Parse(args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// If no topic specified, show general help.
	if fs.NArg() == 0 {
		printUsage(os.Stdout)
		return
	}

	topic := strings.ToLower(fs.Arg(0))

	// Try to find the topic in our help system.
	if help.ShowHelp(topic, *simple) {
		return
	}

	// Not found in help system.
	_, _ = fmt.Fprintf(os.Stdout, "No help found for '%s'\n\n", topic)
	_, _ = fmt.Fprintln(os.Stdout, "Available help topics:")
	_, _ = fmt.Fprintln(os.Stdout, "  Commands:   reflect, test, web, license")
	_, _ = fmt.Fprintln(
		os.Stdout,
		"  Tests:      throughput, latency, frame_loss, y1564_config, ...",
	)
	_, _ = fmt.Fprintln(
		os.Stdout,
		"  Categories: rfc2544, y1564, rfc2889, rfc6349, y1731, mef, tsn",
	)
	_, _ = fmt.Fprintln(os.Stdout, "\nUse 'stem help tests' for a complete list of tests.")
	_, _ = fmt.Fprintln(os.Stdout, "Use 'stem glossary' for network terminology definitions.")
	_, _ = fmt.Fprintln(os.Stdout, "Use 'stem tutorial' for step-by-step guides.")
}

func tutorialCmd(args []string) {
	// If no tutorial specified, list available tutorials.
	if len(args) == 0 {
		help.ShowTutorial("")
		return
	}

	tutorialID := strings.ToLower(strings.TrimSpace(args[0]))

	if !help.ShowTutorial(tutorialID) {
		_, _ = fmt.Fprintf(os.Stdout, "Tutorial '%s' not found.\n\n", tutorialID)
		help.ShowTutorial("") // Show available tutorials.
	}
}

func glossaryCmd(args []string) {
	fs := flag.NewFlagSet("glossary", flag.ExitOnError)
	simple := fs.Bool("simple", false, "Show only simple definitions")
	fs.BoolVar(simple, "s", false, "Show only simple definitions (shorthand)")
	search := fs.String("search", "", "Search for terms containing keyword")

	err := fs.Parse(args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Search mode.
	if *search != "" {
		hs := help.NewSystem()
		results := hs.SearchGlossary(*search)
		if len(results) == 0 {
			_, _ = fmt.Fprintf(os.Stdout, "No terms found matching '%s'\n", *search)
			return
		}
		_, _ = fmt.Fprintf(os.Stdout, "Terms matching '%s':\n\n", *search)
		for _, entry := range results {
			_, _ = fmt.Fprintf(os.Stdout, "  %s - %s\n", entry.Term, entry.FullName)
		}
		_, _ = fmt.Fprintln(os.Stdout, "\nUse 'stem glossary <term>' for full definition.")
		return
	}

	// If no term specified, list all terms.
	if fs.NArg() == 0 {
		help.ShowGlossary("", *simple)
		return
	}

	term := strings.ToLower(fs.Arg(0))

	if !help.ShowGlossary(term, *simple) {
		_, _ = fmt.Fprintf(os.Stdout, "Term '%s' not found in glossary.\n\n", term)
		_, _ = fmt.Fprintln(os.Stdout, "Use 'stem glossary' to see all available terms.")
		_, _ = fmt.Fprintln(os.Stdout, "Use 'stem glossary --search <keyword>' to search.")
	}
}
