/*
 * Seed Test Suite - Help Display Formatting
 *
 * Terminal output formatting for help content.
 *
 * Copyright (c) 2025 Mustard Seed Networks. All rights reserved.
 */

package help

import (
	"fmt"
	"sort"
	"strings"

	"github.com/krisarmstrong/stem/internal/modules"
)

const (
	// Colors and formatting (ANSI escape codes)
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorRed    = "\033[31m"
	colorOrange = "\033[38;5;208m" // Extended color for orange
)

// DisplayTest shows detailed help for a test
func DisplayTest(test TestHelp, simple bool) {
	fmt.Println()
	printHeader(test.Name)
	fmt.Printf("%s%s%s\n", colorDim, test.Standard, colorReset)
	fmt.Println()

	// Summary
	fmt.Println(test.Summary)
	fmt.Println()

	// Description (technical or layman based on mode)
	if simple {
		printSection("What This Test Does", test.LaymanDesc)
	} else {
		printSection("Technical Description", test.TechDesc)
	}

	// When to use
	printSection("When to Use This Test", test.WhenToUse)

	// When not to use
	if test.WhenNotToUse != "" {
		printSection("When NOT to Use", test.WhenNotToUse)
	}

	// Parameters
	if len(test.Parameters) > 0 {
		fmt.Printf("\n%s%sParameters%s\n", colorBold, colorCyan, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		for _, p := range test.Parameters {
			fmt.Printf("\n  %s%s%s", colorBold, p.Flag, colorReset)
			if p.Required {
				fmt.Printf(" %s(required)%s", colorYellow, colorReset)
			}
			fmt.Println()
			fmt.Printf("  Type: %s, Default: %s\n", p.Type, p.Default)
			if simple {
				fmt.Printf("  %s\n", p.LaymanDesc)
			} else {
				fmt.Printf("  %s\n", p.TechDesc)
			}
			if p.Example != "" {
				fmt.Printf("  Example: %s%s%s\n", colorGreen, p.Example, colorReset)
			}
		}
	}

	// Metrics
	if len(test.Metrics) > 0 {
		fmt.Printf("\n%s%sMetrics%s\n", colorBold, colorCyan, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		for _, m := range test.Metrics {
			fmt.Printf("\n  %s%s%s (%s)\n", colorBold, m.Name, colorReset, m.Unit)
			fmt.Printf("  Good: %s%s%s\n", colorGreen, m.GoodRange, colorReset)
			fmt.Printf("  Bad: %s%s%s\n", colorYellow, m.BadMeaning, colorReset)
		}
	}

	// Examples
	if len(test.Examples) > 0 {
		fmt.Printf("\n%s%sExamples%s\n", colorBold, colorCyan, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		for _, ex := range test.Examples {
			fmt.Printf("\n  %s%s%s\n", colorDim, ex.Desc, colorReset)
			fmt.Printf("  $ %s%s%s\n", colorGreen, ex.Command, colorReset)
			if ex.Output != "" {
				fmt.Printf("  %s\n", ex.Output)
			}
		}
	}

	// Tips
	if len(test.Tips) > 0 {
		fmt.Printf("\n%s%sTips%s\n", colorBold, colorCyan, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		for _, tip := range test.Tips {
			fmt.Printf("  • %s\n", tip)
		}
	}

	// Common issues
	if len(test.CommonIssues) > 0 {
		fmt.Printf("\n%s%sCommon Issues%s\n", colorBold, colorCyan, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		for _, issue := range test.CommonIssues {
			fmt.Printf("\n  %sProblem:%s %s\n", colorYellow, colorReset, issue.Problem)
			fmt.Printf("  Cause: %s\n", issue.Cause)
			fmt.Printf("  %sSolution:%s %s\n", colorGreen, colorReset, issue.Solution)
		}
	}

	// See also
	if len(test.SeeAlso) > 0 {
		fmt.Printf("\n%s%sSee Also%s\n", colorBold, colorCyan, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("  Related tests: %s\n", strings.Join(test.SeeAlso, ", "))
		fmt.Printf("  Run: stem help <test-name>\n")
	}

	fmt.Println()
}

// DisplayCommand shows detailed help for a command
func DisplayCommand(cmd CommandHelp) {
	fmt.Println()
	printHeader(fmt.Sprintf("stem %s", cmd.Name))
	fmt.Println()
	fmt.Println(cmd.Summary)
	fmt.Println()

	printSection("Description", cmd.Description)

	fmt.Printf("\n%s%sUsage%s\n", colorBold, colorCyan, colorReset)
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("  %s\n", cmd.Usage)

	if len(cmd.Flags) > 0 {
		fmt.Printf("\n%s%sFlags%s\n", colorBold, colorCyan, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		for _, f := range cmd.Flags {
			flag := f.Long
			if f.Short != "" {
				flag = fmt.Sprintf("%s, %s", f.Short, f.Long)
			}
			fmt.Printf("\n  %s%s%s", colorBold, flag, colorReset)
			if f.Required {
				fmt.Printf(" %s(required)%s", colorYellow, colorReset)
			}
			fmt.Println()
			fmt.Printf("  Type: %s, Default: %s\n", f.Type, f.Default)
			fmt.Printf("  %s\n", f.LaymanDesc)
		}
	}

	if len(cmd.Examples) > 0 {
		fmt.Printf("\n%s%sExamples%s\n", colorBold, colorCyan, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		for _, ex := range cmd.Examples {
			fmt.Printf("\n  %s%s%s\n", colorDim, ex.Desc, colorReset)
			fmt.Printf("  $ %s%s%s\n", colorGreen, ex.Command, colorReset)
		}
	}

	if len(cmd.SeeAlso) > 0 {
		fmt.Printf("\n%s%sSee Also%s\n", colorBold, colorCyan, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		for _, s := range cmd.SeeAlso {
			fmt.Printf("  • stem help %s\n", s)
		}
	}

	fmt.Println()
}

// DisplayCategory shows a test category overview
func DisplayCategory(cat Category) {
	fmt.Println()
	printHeader(cat.Name)
	fmt.Printf("%s%s%s\n", colorDim, cat.FullName, colorReset)
	fmt.Println()

	fmt.Println(cat.Summary)
	fmt.Println()

	printSection("Description", cat.Description)

	fmt.Printf("\n%s%sTests in this Category%s\n", colorBold, colorCyan, colorReset)
	fmt.Println(strings.Repeat("─", 60))

	hs := NewHelpSystem()
	for _, testID := range cat.Tests {
		if test, ok := hs.Tests[testID]; ok {
			fmt.Printf("  %s%-20s%s %s\n", colorBold, testID, colorReset, test.Summary)
		}
	}

	printSection("When to Use", cat.WhenToUse)

	if len(cat.SeeAlso) > 0 {
		fmt.Printf("\n%s%sSee Also%s\n", colorBold, colorCyan, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		for _, s := range cat.SeeAlso {
			fmt.Printf("  • stem help %s\n", s)
		}
	}

	fmt.Println()
}

// DisplayGlossaryTerm shows a glossary entry
func DisplayGlossaryTerm(entry GlossaryEntry, simple bool) {
	fmt.Println()
	printHeader(entry.Term)
	fmt.Printf("%s%s%s\n", colorDim, entry.FullName, colorReset)
	fmt.Println()

	if simple {
		fmt.Printf("%s%sSimple Definition%s\n", colorBold, colorCyan, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("  %s\n", entry.LaymanDef)
	} else {
		fmt.Printf("%s%sTechnical Definition%s\n", colorBold, colorCyan, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("  %s\n", entry.TechDef)
		fmt.Println()
		fmt.Printf("%s%sSimple Definition%s\n", colorBold, colorCyan, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("  %s\n", entry.LaymanDef)
	}

	if len(entry.Related) > 0 {
		fmt.Printf("\n%s%sRelated Terms%s\n", colorBold, colorCyan, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("  %s\n", strings.Join(entry.Related, ", "))
	}

	fmt.Println()
}

// DisplayGlossaryList shows all glossary terms by category
func DisplayGlossaryList() {
	fmt.Println()
	printHeader("Network Glossary")
	fmt.Println()
	fmt.Println("Look up network testing terminology.")
	fmt.Printf("Usage: %sstem glossary <term>%s\n", colorGreen, colorReset)
	fmt.Println()

	categories := GetGlossaryTermsByCategory()

	// Sort category names
	catNames := make([]string, 0, len(categories))
	for name := range categories {
		catNames = append(catNames, name)
	}
	sort.Strings(catNames)

	for _, catName := range catNames {
		terms := categories[catName]
		fmt.Printf("%s%s%s\n", colorBold, catName, colorReset)
		fmt.Printf("  %s\n\n", strings.Join(terms, ", "))
	}

	fmt.Println("Use 'stem glossary <term>' for definition.")
	fmt.Println()
}

// DisplayTutorial shows a tutorial
func DisplayTutorial(tutorial Tutorial) {
	fmt.Println()
	printHeader(tutorial.Title)
	fmt.Printf("%s%s | %s | %s%s\n", colorDim, tutorial.Duration, tutorial.Level, tutorial.Description, colorReset)
	fmt.Println()

	for i, step := range tutorial.Steps {
		fmt.Printf("%s%sStep %d: %s%s\n", colorBold, colorCyan, i+1, step.Title, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println()
		fmt.Println(step.Content)

		if step.Command != "" {
			fmt.Println()
			fmt.Printf("  $ %s%s%s\n", colorGreen, step.Command, colorReset)
			if step.Expected != "" {
				fmt.Printf("  %s%s%s\n", colorDim, step.Expected, colorReset)
			}
		}

		if step.Tip != "" {
			fmt.Println()
			fmt.Printf("  💡 %s%sTip:%s %s\n", colorYellow, colorBold, colorReset, step.Tip)
		}

		fmt.Println()
	}
}

// DisplayTutorialList shows available tutorials
func DisplayTutorialList() {
	fmt.Println()
	printHeader("Available Tutorials")
	fmt.Println()
	fmt.Println("Step-by-step guides to help you learn Seed Test Suite.")
	fmt.Printf("Usage: %sstem tutorial <name>%s\n", colorGreen, colorReset)
	fmt.Println()

	tutorials := GetAllTutorials()

	// Sort by ID
	ids := make([]string, 0, len(tutorials))
	for id := range tutorials {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	fmt.Printf("%s%-15s %-30s %-10s %s%s\n", colorBold, "ID", "Title", "Duration", "Level", colorReset)
	fmt.Println(strings.Repeat("─", 70))

	for _, id := range ids {
		t := tutorials[id]
		fmt.Printf("%-15s %-30s %-10s %s\n", id, t.Title, t.Duration, t.Level)
	}

	fmt.Println()
	fmt.Printf("Run a tutorial: %sstem tutorial quickstart%s\n", colorGreen, colorReset)
	fmt.Println()
}

// DisplayTestList shows all available tests
func DisplayTestList() {
	fmt.Println()
	printHeader("Available Tests")
	fmt.Println()
	fmt.Println("Seed Test Suite supports 27 test types across 7 categories.")
	fmt.Printf("Usage: %sstem test -i <interface> -t <test-type>%s\n", colorGreen, colorReset)
	fmt.Println()

	categories := GetAllCategories()

	// Define category order
	catOrder := []string{"rfc2544", "y1564", "rfc2889", "rfc6349", "y1731", "mef", "tsn"}

	hs := NewHelpSystem()

	for _, catID := range catOrder {
		cat := categories[catID]
		fmt.Printf("%s%s%s - %s\n", colorBold, cat.Name, colorReset, cat.Summary)
		fmt.Println(strings.Repeat("─", 70))

		for _, testID := range cat.Tests {
			if test, ok := hs.Tests[testID]; ok {
				fmt.Printf("  %s%-20s%s %s\n", colorCyan, testID, colorReset, test.Summary)
			}
		}
		fmt.Println()
	}

	fmt.Printf("For details: %sstem help <test-name>%s\n", colorGreen, colorReset)
	fmt.Printf("For category: %sstem help rfc2544%s\n", colorGreen, colorReset)
	fmt.Println()
}

// DisplayTestListByModule shows all available tests grouped by module
func DisplayTestListByModule() {
	fmt.Println()
	printHeader("Available Tests by Module")
	fmt.Println()

	allModules := modules.GetAllModules()
	totalTests := 0
	for _, m := range allModules {
		totalTests += len(m.TestTypes())
	}

	fmt.Printf("Stem supports %d test types across %d modules.\n", totalTests, len(allModules))
	fmt.Printf("Usage: %sstem test -i <interface> -t <test-type>%s\n", colorGreen, colorReset)
	fmt.Println()

	// Define module display order (Tier 1 first, then Tier 2 by color) - use lowercase module names
	moduleOrder := []string{"reflector", "benchmark", "servicetest", "trafficgen", "measure", "certify"}

	hs := NewHelpSystem()

	for _, modName := range moduleOrder {
		mod := modules.GetModule(modName)
		if mod == nil {
			continue
		}

		modColor := moduleColorToANSI(mod.Color())
		fmt.Printf("%s%s%s%s - %s\n", modColor, colorBold, mod.DisplayName(), colorReset, mod.Description())
		fmt.Printf("%sStandard: %s%s\n", colorDim, mod.Standard(), colorReset)
		fmt.Println(strings.Repeat("─", 70))

		testTypes := mod.TestTypes()
		// Sort test types for consistent display
		sort.Strings(testTypes)

		for _, testType := range testTypes {
			if test, ok := hs.Tests[testType]; ok {
				fmt.Printf("  %s%-20s%s %s\n", modColor, testType, colorReset, test.Summary)
			} else {
				// Fallback if test not in help system
				fmt.Printf("  %s%-20s%s\n", modColor, testType, colorReset)
			}
		}
		fmt.Println()
	}

	fmt.Printf("For details: %sstem help <test-name>%s\n", colorGreen, colorReset)
	fmt.Printf("For module view: %sstem list-tests --by-module%s\n", colorGreen, colorReset)
	fmt.Println()
}

// moduleColorToANSI converts a hex color code to ANSI escape code
func moduleColorToANSI(hexColor string) string {
	// Map module colors to ANSI codes
	switch hexColor {
	case "#0891b2": // Reflector - Cyan
		return colorCyan
	case "#dc2626": // Benchmark - Red
		return colorRed
	case "#ea580c": // ServiceTest - Orange
		return colorOrange
	case "#ca8a04": // TrafficGen - Yellow
		return colorYellow
	case "#2563eb": // Measure - Blue
		return colorBlue
	case "#16a34a": // Certify - Green
		return colorGreen
	default:
		return colorCyan
	}
}

// Helper functions

func printHeader(text string) {
	fmt.Printf("%s%s%s\n", colorBold, text, colorReset)
	fmt.Println(strings.Repeat("═", len(text)+2))
}

func printSection(title, content string) {
	fmt.Printf("\n%s%s%s%s\n", colorBold, colorCyan, title, colorReset)
	fmt.Println(strings.Repeat("─", 60))
	// Indent content
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		fmt.Printf("  %s\n", line)
	}
}

// ShowHelp is the main entry point for the help system
func ShowHelp(topic string, simple bool) bool {
	hs := NewHelpSystem()

	// Check if topic is a test
	if test, ok := hs.GetTest(topic); ok {
		DisplayTest(test, simple)
		return true
	}

	// Check if topic is a command
	if cmd, ok := hs.GetCommand(topic); ok {
		DisplayCommand(cmd)
		return true
	}

	// Check if topic is a category
	if cat, ok := hs.GetCategory(topic); ok {
		DisplayCategory(cat)
		return true
	}

	// Check if topic is "tests"
	if topic == "tests" || topic == "list" {
		DisplayTestList()
		return true
	}

	// Check if topic is module-related
	if topic == "modules" || topic == "by-module" || topic == "tests-by-module" {
		DisplayTestListByModule()
		return true
	}

	// Not found
	return false
}

// ShowGlossary handles glossary lookups
func ShowGlossary(term string, simple bool) bool {
	hs := NewHelpSystem()

	if term == "" {
		DisplayGlossaryList()
		return true
	}

	// Normalize term
	term = strings.ToLower(strings.TrimSpace(term))

	if entry, ok := hs.GetGlossaryTerm(term); ok {
		DisplayGlossaryTerm(entry, simple)
		return true
	}

	// Search for partial matches
	results := hs.SearchGlossary(term)
	if len(results) > 0 {
		fmt.Printf("\nNo exact match for '%s'. Did you mean:\n", term)
		for _, r := range results {
			fmt.Printf("  • %s - %s\n", r.Term, r.FullName)
		}
		fmt.Println()
		return true
	}

	return false
}

// ShowTutorial handles tutorial display
func ShowTutorial(id string) bool {
	hs := NewHelpSystem()

	if id == "" {
		DisplayTutorialList()
		return true
	}

	if tutorial, ok := hs.GetTutorial(id); ok {
		DisplayTutorial(tutorial)
		return true
	}

	return false
}
