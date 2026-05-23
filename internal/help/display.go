/*
 * Seed Test Suite - Help Display Formatting
 *
 * Terminal output formatting for help content.
 */

package help

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	modules "github.com/krisarmstrong/stem/internal/services"
)

const (
	// Colors and formatting (ANSI escape codes).
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorRed    = "\033[31m"
	colorOrange = "\033[38;5;208m" // Extended color for orange.

	// separatorWidth is the width of section separators.
	separatorWidth = 60

	// wideColumnWidth is the width of wide table separators.
	wideColumnWidth = 70

	// headerPadding is the padding added to header text for separator width.
	headerPadding = 2
)

// DisplayTest shows detailed help for a test.
func DisplayTest(test TestHelp, simple bool) {
	DisplayTestTo(os.Stdout, test, simple)
}

// DisplayTestTo shows detailed help for a test using the provided writer.
func DisplayTestTo(w io.Writer, test TestHelp, simple bool) {
	_, _ = fmt.Fprintln(w)
	printHeader(w, test.Name)
	_, _ = fmt.Fprintf(w, "%s%s%s\n", colorDim, test.Standard, colorReset)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, test.Summary)
	_, _ = fmt.Fprintln(w)

	displayTestDescription(w, test, simple)
	printSection(w, "When to Use This Test", test.WhenToUse)
	displayTestWhenNotToUse(w, test)
	displayTestParameters(w, test, simple)
	displayTestMetrics(w, test)
	displayTestExamples(w, test)
	displayTestTips(w, test)
	displayTestCommonIssues(w, test)
	displayTestSeeAlso(w, test)
	_, _ = fmt.Fprintln(w)
}

func displayTestDescription(w io.Writer, test TestHelp, simple bool) {
	if simple {
		printSection(w, "What This Test Does", test.LaymanDesc)
	} else {
		printSection(w, "Technical Description", test.TechDesc)
	}
}

func displayTestWhenNotToUse(w io.Writer, test TestHelp) {
	if test.WhenNotToUse != "" {
		printSection(w, "When NOT to Use", test.WhenNotToUse)
	}
}

func displayTestParameters(w io.Writer, test TestHelp, simple bool) {
	if len(test.Parameters) == 0 {
		return
	}
	_, _ = fmt.Fprintf(w, "\n%s%sParameters%s\n", colorBold, colorCyan, colorReset)
	_, _ = fmt.Fprintln(w, strings.Repeat("─", separatorWidth))
	for _, p := range test.Parameters {
		printParameter(w, p, simple)
	}
}

func printParameter(w io.Writer, p Parameter, simple bool) {
	_, _ = fmt.Fprintf(w, "\n  %s%s%s", colorBold, p.Flag, colorReset)
	if p.Required {
		_, _ = fmt.Fprintf(w, " %s(required)%s", colorYellow, colorReset)
	}
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "  Type: %s, Default: %s\n", p.Type, p.Default)
	if simple {
		_, _ = fmt.Fprintf(w, "  %s\n", p.LaymanDesc)
	} else {
		_, _ = fmt.Fprintf(w, "  %s\n", p.TechDesc)
	}
	if p.Example != "" {
		_, _ = fmt.Fprintf(w, "  Example: %s%s%s\n", colorGreen, p.Example, colorReset)
	}
}

func displayTestMetrics(w io.Writer, test TestHelp) {
	if len(test.Metrics) == 0 {
		return
	}
	_, _ = fmt.Fprintf(w, "\n%s%sMetrics%s\n", colorBold, colorCyan, colorReset)
	_, _ = fmt.Fprintln(w, strings.Repeat("─", separatorWidth))
	for _, m := range test.Metrics {
		_, _ = fmt.Fprintf(w, "\n  %s%s%s (%s)\n", colorBold, m.Name, colorReset, m.Unit)
		_, _ = fmt.Fprintf(w, "  Good: %s%s%s\n", colorGreen, m.GoodRange, colorReset)
		_, _ = fmt.Fprintf(w, "  Bad: %s%s%s\n", colorYellow, m.BadMeaning, colorReset)
	}
}

func displayTestExamples(w io.Writer, test TestHelp) {
	if len(test.Examples) == 0 {
		return
	}
	_, _ = fmt.Fprintf(w, "\n%s%sExamples%s\n", colorBold, colorCyan, colorReset)
	_, _ = fmt.Fprintln(w, strings.Repeat("─", separatorWidth))
	for _, ex := range test.Examples {
		_, _ = fmt.Fprintf(w, "\n  %s%s%s\n", colorDim, ex.Desc, colorReset)
		_, _ = fmt.Fprintf(w, "  $ %s%s%s\n", colorGreen, ex.Command, colorReset)
		if ex.Output != "" {
			_, _ = fmt.Fprintf(w, "  %s\n", ex.Output)
		}
	}
}

func displayTestTips(w io.Writer, test TestHelp) {
	if len(test.Tips) == 0 {
		return
	}
	_, _ = fmt.Fprintf(w, "\n%s%sTips%s\n", colorBold, colorCyan, colorReset)
	_, _ = fmt.Fprintln(w, strings.Repeat("─", separatorWidth))
	for _, tip := range test.Tips {
		_, _ = fmt.Fprintf(w, "  • %s\n", tip)
	}
}

func displayTestCommonIssues(w io.Writer, test TestHelp) {
	if len(test.CommonIssues) == 0 {
		return
	}
	_, _ = fmt.Fprintf(w, "\n%s%sCommon Issues%s\n", colorBold, colorCyan, colorReset)
	_, _ = fmt.Fprintln(w, strings.Repeat("─", separatorWidth))
	for _, issue := range test.CommonIssues {
		_, _ = fmt.Fprintf(w, "\n  %sProblem:%s %s\n", colorYellow, colorReset, issue.Problem)
		_, _ = fmt.Fprintf(w, "  Cause: %s\n", issue.Cause)
		_, _ = fmt.Fprintf(w, "  %sSolution:%s %s\n", colorGreen, colorReset, issue.Solution)
	}
}

func displayTestSeeAlso(w io.Writer, test TestHelp) {
	if len(test.SeeAlso) == 0 {
		return
	}
	_, _ = fmt.Fprintf(w, "\n%s%sSee Also%s\n", colorBold, colorCyan, colorReset)
	_, _ = fmt.Fprintln(w, strings.Repeat("─", separatorWidth))
	_, _ = fmt.Fprintf(w, "  Related tests: %s\n", strings.Join(test.SeeAlso, ", "))
	_, _ = fmt.Fprint(w, "  Run: stem help <test-name>\n")
}

// DisplayCommand shows detailed help for a command.
func DisplayCommand(cmd CommandHelp) {
	DisplayCommandTo(os.Stdout, cmd)
}

// DisplayCommandTo shows detailed help for a command using the provided writer.
func DisplayCommandTo(w io.Writer, cmd CommandHelp) {
	_, _ = fmt.Fprintln(w)
	printHeader(w, "stem "+cmd.Name)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, cmd.Summary)
	_, _ = fmt.Fprintln(w)

	printSection(w, "Description", cmd.Description)

	_, _ = fmt.Fprintf(w, "\n%s%sUsage%s\n", colorBold, colorCyan, colorReset)
	_, _ = fmt.Fprintln(w, strings.Repeat("─", separatorWidth))
	_, _ = fmt.Fprintf(w, "  %s\n", cmd.Usage)

	if len(cmd.Flags) > 0 {
		_, _ = fmt.Fprintf(w, "\n%s%sFlags%s\n", colorBold, colorCyan, colorReset)
		_, _ = fmt.Fprintln(w, strings.Repeat("─", separatorWidth))
		for _, f := range cmd.Flags {
			flag := f.Long
			if f.Short != "" {
				flag = f.Short + ", " + f.Long
			}
			_, _ = fmt.Fprintf(w, "\n  %s%s%s", colorBold, flag, colorReset)
			if f.Required {
				_, _ = fmt.Fprintf(w, " %s(required)%s", colorYellow, colorReset)
			}
			_, _ = fmt.Fprintln(w)
			_, _ = fmt.Fprintf(w, "  Type: %s, Default: %s\n", f.Type, f.Default)
			_, _ = fmt.Fprintf(w, "  %s\n", f.LaymanDesc)
		}
	}

	if len(cmd.Examples) > 0 {
		_, _ = fmt.Fprintf(w, "\n%s%sExamples%s\n", colorBold, colorCyan, colorReset)
		_, _ = fmt.Fprintln(w, strings.Repeat("─", separatorWidth))
		for _, ex := range cmd.Examples {
			_, _ = fmt.Fprintf(w, "\n  %s%s%s\n", colorDim, ex.Desc, colorReset)
			_, _ = fmt.Fprintf(w, "  $ %s%s%s\n", colorGreen, ex.Command, colorReset)
		}
	}

	if len(cmd.SeeAlso) > 0 {
		_, _ = fmt.Fprintf(w, "\n%s%sSee Also%s\n", colorBold, colorCyan, colorReset)
		_, _ = fmt.Fprintln(w, strings.Repeat("─", separatorWidth))
		for _, s := range cmd.SeeAlso {
			_, _ = fmt.Fprintf(w, "  • stem help %s\n", s)
		}
	}

	_, _ = fmt.Fprintln(w)
}

// DisplayCategory shows a test category overview.
func DisplayCategory(cat Category) {
	DisplayCategoryTo(os.Stdout, cat)
}

// DisplayCategoryTo shows a test category overview using the provided writer.
func DisplayCategoryTo(w io.Writer, cat Category) {
	_, _ = fmt.Fprintln(w)
	printHeader(w, cat.Name)
	_, _ = fmt.Fprintf(w, "%s%s%s\n", colorDim, cat.FullName, colorReset)
	_, _ = fmt.Fprintln(w)

	_, _ = fmt.Fprintln(w, cat.Summary)
	_, _ = fmt.Fprintln(w)

	printSection(w, "Description", cat.Description)

	_, _ = fmt.Fprintf(w, "\n%s%sTests in this Category%s\n", colorBold, colorCyan, colorReset)
	_, _ = fmt.Fprintln(w, strings.Repeat("─", separatorWidth))

	hs := NewSystem()
	for _, testID := range cat.Tests {
		if test, ok := hs.Tests[testID]; ok {
			_, _ = fmt.Fprintf(w, "  %s%-20s%s %s\n", colorBold, testID, colorReset, test.Summary)
		}
	}

	printSection(w, "When to Use", cat.WhenToUse)

	if len(cat.SeeAlso) > 0 {
		_, _ = fmt.Fprintf(w, "\n%s%sSee Also%s\n", colorBold, colorCyan, colorReset)
		_, _ = fmt.Fprintln(w, strings.Repeat("─", separatorWidth))
		for _, s := range cat.SeeAlso {
			_, _ = fmt.Fprintf(w, "  • stem help %s\n", s)
		}
	}

	_, _ = fmt.Fprintln(w)
}

// DisplayGlossaryTerm shows a glossary entry.
func DisplayGlossaryTerm(entry GlossaryEntry, simple bool) {
	DisplayGlossaryTermTo(os.Stdout, entry, simple)
}

// DisplayGlossaryTermTo shows a glossary entry using the provided writer.
func DisplayGlossaryTermTo(w io.Writer, entry GlossaryEntry, simple bool) {
	_, _ = fmt.Fprintln(w)
	printHeader(w, entry.Term)
	_, _ = fmt.Fprintf(w, "%s%s%s\n", colorDim, entry.FullName, colorReset)
	_, _ = fmt.Fprintln(w)

	if simple {
		_, _ = fmt.Fprintf(w, "%s%sSimple Definition%s\n", colorBold, colorCyan, colorReset)
		_, _ = fmt.Fprintln(w, strings.Repeat("─", separatorWidth))
		_, _ = fmt.Fprintf(w, "  %s\n", entry.LaymanDef)
	} else {
		_, _ = fmt.Fprintf(w, "%s%sTechnical Definition%s\n", colorBold, colorCyan, colorReset)
		_, _ = fmt.Fprintln(w, strings.Repeat("─", separatorWidth))
		_, _ = fmt.Fprintf(w, "  %s\n", entry.TechDef)
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintf(w, "%s%sSimple Definition%s\n", colorBold, colorCyan, colorReset)
		_, _ = fmt.Fprintln(w, strings.Repeat("─", separatorWidth))
		_, _ = fmt.Fprintf(w, "  %s\n", entry.LaymanDef)
	}

	if len(entry.Related) > 0 {
		_, _ = fmt.Fprintf(w, "\n%s%sRelated Terms%s\n", colorBold, colorCyan, colorReset)
		_, _ = fmt.Fprintln(w, strings.Repeat("─", separatorWidth))
		_, _ = fmt.Fprintf(w, "  %s\n", strings.Join(entry.Related, ", "))
	}

	_, _ = fmt.Fprintln(w)
}

// DisplayGlossaryList shows all glossary terms by category.
func DisplayGlossaryList() {
	DisplayGlossaryListTo(os.Stdout)
}

// DisplayGlossaryListTo shows all glossary terms by category using the provided writer.
func DisplayGlossaryListTo(w io.Writer) {
	_, _ = fmt.Fprintln(w)
	printHeader(w, "Network Glossary")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Look up network testing terminology.")
	_, _ = fmt.Fprintf(w, "Usage: %sstem glossary <term>%s\n", colorGreen, colorReset)
	_, _ = fmt.Fprintln(w)

	categories := GetGlossaryTermsByCategory()

	// Sort category names
	catNames := make([]string, 0, len(categories))
	for name := range categories {
		catNames = append(catNames, name)
	}
	sort.Strings(catNames)

	for _, catName := range catNames {
		terms := categories[catName]
		_, _ = fmt.Fprintf(w, "%s%s%s\n", colorBold, catName, colorReset)
		_, _ = fmt.Fprintf(w, "  %s\n\n", strings.Join(terms, ", "))
	}

	_, _ = fmt.Fprintln(w, "Use 'stem glossary <term>' for definition.")
	_, _ = fmt.Fprintln(w)
}

// DisplayTutorial shows a tutorial.
func DisplayTutorial(tutorial Tutorial) {
	DisplayTutorialTo(os.Stdout, tutorial)
}

// DisplayTutorialTo shows a tutorial using the provided writer.
func DisplayTutorialTo(w io.Writer, tutorial Tutorial) {
	_, _ = fmt.Fprintln(w)
	printHeader(w, tutorial.Title)
	_, _ = fmt.Fprintf(
		w, "%s%s | %s | %s%s\n",
		colorDim, tutorial.Duration, tutorial.Level, tutorial.Description, colorReset,
	)
	_, _ = fmt.Fprintln(w)

	for i, step := range tutorial.Steps {
		_, _ = fmt.Fprintf(w, "%s%sStep %d: %s%s\n", colorBold, colorCyan, i+1, step.Title, colorReset)
		_, _ = fmt.Fprintln(w, strings.Repeat("─", separatorWidth))
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, step.Content)

		if step.Command != "" {
			_, _ = fmt.Fprintln(w)
			_, _ = fmt.Fprintf(w, "  $ %s%s%s\n", colorGreen, step.Command, colorReset)
			if step.Expected != "" {
				_, _ = fmt.Fprintf(w, "  %s%s%s\n", colorDim, step.Expected, colorReset)
			}
		}

		if step.Tip != "" {
			_, _ = fmt.Fprintln(w)
			_, _ = fmt.Fprintf(w, "  💡 %s%sTip:%s %s\n", colorYellow, colorBold, colorReset, step.Tip)
		}

		_, _ = fmt.Fprintln(w)
	}
}

// DisplayTutorialList shows available tutorials.
func DisplayTutorialList() {
	DisplayTutorialListTo(os.Stdout)
}

// DisplayTutorialListTo shows available tutorials using the provided writer.
func DisplayTutorialListTo(w io.Writer) {
	_, _ = fmt.Fprintln(w)
	printHeader(w, "Available Tutorials")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Step-by-step guides to help you learn Seed Test Suite.")
	_, _ = fmt.Fprintf(w, "Usage: %sstem tutorial <name>%s\n", colorGreen, colorReset)
	_, _ = fmt.Fprintln(w)

	tutorials := GetAllTutorials()

	// Sort by ID
	ids := make([]string, 0, len(tutorials))
	for id := range tutorials {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	_, _ = fmt.Fprintf(w, "%s%-15s %-30s %-10s %s%s\n", colorBold, "ID", "Title", "Duration", "Level", colorReset)
	_, _ = fmt.Fprintln(w, strings.Repeat("─", wideColumnWidth))

	for _, id := range ids {
		t := tutorials[id]
		_, _ = fmt.Fprintf(w, "%-15s %-30s %-10s %s\n", id, t.Title, t.Duration, t.Level)
	}

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "Run a tutorial: %sstem tutorial quickstart%s\n", colorGreen, colorReset)
	_, _ = fmt.Fprintln(w)
}

// DisplayTestList shows all available tests.
func DisplayTestList() {
	DisplayTestListTo(os.Stdout)
}

// DisplayTestListTo shows all available tests using the provided writer.
func DisplayTestListTo(w io.Writer) {
	_, _ = fmt.Fprintln(w)
	printHeader(w, "Available Tests")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Seed Test Suite supports 27 test types across 7 categories.")
	_, _ = fmt.Fprintf(w, "Usage: %sstem test -i <interface> -t <test-type>%s\n", colorGreen, colorReset)
	_, _ = fmt.Fprintln(w)

	categories := GetAllCategories()

	// Define category order
	catOrder := []string{CatRFC2544, CatY1564, CatRFC2889, CatRFC6349, CatY1731, CatMEF, CatTSN}

	hs := NewSystem()

	for _, catID := range catOrder {
		cat := categories[catID]
		_, _ = fmt.Fprintf(w, "%s%s%s - %s\n", colorBold, cat.Name, colorReset, cat.Summary)
		_, _ = fmt.Fprintln(w, strings.Repeat("─", wideColumnWidth))

		for _, testID := range cat.Tests {
			if test, ok := hs.Tests[testID]; ok {
				_, _ = fmt.Fprintf(w, "  %s%-20s%s %s\n", colorCyan, testID, colorReset, test.Summary)
			}
		}
		_, _ = fmt.Fprintln(w)
	}

	_, _ = fmt.Fprintf(w, "For details: %sstem help <test-name>%s\n", colorGreen, colorReset)
	_, _ = fmt.Fprintf(w, "For category: %sstem help rfc2544%s\n", colorGreen, colorReset)
	_, _ = fmt.Fprintln(w)
}

// DisplayTestListByModule shows all available tests grouped by module.
func DisplayTestListByModule() {
	DisplayTestListByModuleTo(os.Stdout)
}

// DisplayTestListByModuleTo shows all available tests grouped by module using the provided writer.
func DisplayTestListByModuleTo(w io.Writer) {
	_, _ = fmt.Fprintln(w)
	printHeader(w, "Available Tests by Module")
	_, _ = fmt.Fprintln(w)

	allModules := modules.GetAllModules()
	totalTests := 0
	for _, m := range allModules {
		totalTests += len(m.TestTypes())
	}

	_, _ = fmt.Fprintf(w, "The Stem supports %d test types across %d modules.\n", totalTests, len(allModules))
	_, _ = fmt.Fprintf(w, "Usage: %sstem test -i <interface> -t <test-type>%s\n", colorGreen, colorReset)
	_, _ = fmt.Fprintln(w)

	// Define module display order (Tier 1 first, then Tier 2 by color) - use lowercase module names
	moduleOrder := []string{"reflector", ModuleBenchmark, "servicetest", "trafficgen", "measure", "certify"}

	hs := NewSystem()

	for _, modName := range moduleOrder {
		mod := modules.GetModule(modName)
		if mod == nil {
			continue
		}

		modColor := ModuleColorToANSI(mod.Color())
		_, _ = fmt.Fprintf(w, "%s%s%s%s - %s\n", modColor, colorBold, mod.DisplayName(), colorReset, mod.Description())
		_, _ = fmt.Fprintf(w, "%sStandard: %s%s\n", colorDim, mod.Standard(), colorReset)
		_, _ = fmt.Fprintln(w, strings.Repeat("─", wideColumnWidth))

		testTypes := mod.TestTypes()
		// Sort test types for consistent display
		sort.Strings(testTypes)

		for _, testType := range testTypes {
			if test, ok := hs.Tests[testType]; ok {
				_, _ = fmt.Fprintf(w, "  %s%-20s%s %s\n", modColor, testType, colorReset, test.Summary)
			} else {
				// Fallback if test not in help system
				_, _ = fmt.Fprintf(w, "  %s%-20s%s\n", modColor, testType, colorReset)
			}
		}
		_, _ = fmt.Fprintln(w)
	}

	_, _ = fmt.Fprintf(w, "For details: %sstem help <test-name>%s\n", colorGreen, colorReset)
	_, _ = fmt.Fprintf(w, "For module view: %sstem list-tests%s\n", colorGreen, colorReset)
	_, _ = fmt.Fprintln(w)
}

// ModuleColorToANSI converts a hex color code to ANSI escape code.
func ModuleColorToANSI(hexColor string) string {
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

func printHeader(w io.Writer, text string) {
	_, _ = fmt.Fprintf(w, "%s%s%s\n", colorBold, text, colorReset)
	_, _ = fmt.Fprintln(w, strings.Repeat("═", len(text)+headerPadding))
}

func printSection(w io.Writer, title, content string) {
	_, _ = fmt.Fprintf(w, "\n%s%s%s%s\n", colorBold, colorCyan, title, colorReset)
	_, _ = fmt.Fprintln(w, strings.Repeat("─", separatorWidth))
	// Indent content.
	for line := range strings.SplitSeq(content, "\n") {
		_, _ = fmt.Fprintf(w, "  %s\n", line)
	}
}

// ShowHelp is the main entry point for the help system.
func ShowHelp(topic string, simple bool) bool {
	return ShowHelpTo(os.Stdout, topic, simple)
}

// ShowHelpTo is the main entry point for the help system using the provided writer.
func ShowHelpTo(w io.Writer, topic string, simple bool) bool {
	hs := NewSystem()

	// Check if topic is a test
	if test, ok := hs.GetTest(topic); ok {
		DisplayTestTo(w, test, simple)
		return true
	}

	// Check if topic is a command
	if cmd, ok := hs.GetCommand(topic); ok {
		DisplayCommandTo(w, cmd)
		return true
	}

	// Check if topic is a category
	if cat, ok := hs.GetCategory(topic); ok {
		DisplayCategoryTo(w, cat)
		return true
	}

	// Check if topic is "tests"
	if topic == "tests" || topic == "list" {
		DisplayTestListTo(w)
		return true
	}

	// Check if topic is module-related
	if topic == "modules" {
		DisplayTestListByModuleTo(w)
		return true
	}

	// Not found
	return false
}

// ShowGlossary handles glossary lookups.
func ShowGlossary(term string, simple bool) bool {
	return ShowGlossaryTo(os.Stdout, term, simple)
}

// ShowGlossaryTo handles glossary lookups using the provided writer.
func ShowGlossaryTo(w io.Writer, term string, simple bool) bool {
	hs := NewSystem()

	if term == "" {
		DisplayGlossaryListTo(w)
		return true
	}

	// Normalize term
	term = strings.ToLower(strings.TrimSpace(term))

	if entry, ok := hs.GetGlossaryTerm(term); ok {
		DisplayGlossaryTermTo(w, entry, simple)
		return true
	}

	// Search for partial matches
	results := hs.SearchGlossary(term)
	if len(results) > 0 {
		_, _ = fmt.Fprintf(w, "\nNo exact match for '%s'. Did you mean:\n", term)
		for _, r := range results {
			_, _ = fmt.Fprintf(w, "  • %s - %s\n", r.Term, r.FullName)
		}
		_, _ = fmt.Fprintln(w)
		return true
	}

	return false
}

// ShowTutorial handles tutorial display.
func ShowTutorial(id string) bool {
	return ShowTutorialTo(os.Stdout, id)
}

// ShowTutorialTo handles tutorial display using the provided writer.
func ShowTutorialTo(w io.Writer, id string) bool {
	hs := NewSystem()

	if id == "" {
		DisplayTutorialListTo(w)
		return true
	}

	if tutorial, ok := hs.GetTutorial(id); ok {
		DisplayTutorialTo(w, tutorial)
		return true
	}

	return false
}
