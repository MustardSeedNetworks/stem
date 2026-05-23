/*
 * Seed Test Suite - Unified Help System
 *
 * This package provides help content for both CLI and WebUI.
 * All test explanations include both technical and layman-friendly descriptions.
 */

package help

// TestHelp contains comprehensive documentation for a single test type.
type TestHelp struct {
	ID       string // Unique identifier: "throughput", "latency", etc.
	Name     string // Display name: "Throughput Test"
	Standard string // Reference standard: "RFC 2544 Section 26.1"
	Category string // Category: "RFC 2544", "Y.1564", etc.

	// Two-tier explanations
	Summary    string // 1 sentence for tooltips
	TechDesc   string // Technical explanation for engineers
	LaymanDesc string // Plain English for non-technical users

	// Usage guidance
	WhenToUse    string // Scenarios where this test applies
	WhenNotToUse string // When a different test is better

	// Parameters
	Parameters []Parameter

	// Results interpretation
	Metrics            []Metric
	SuccessCriteria    string
	FailureExplanation string

	// Practical guidance
	Examples     []Example
	Tips         []string
	CommonIssues []Issue

	// References
	RFCSection string   // e.g., "Section 26.1"
	SeeAlso    []string // Related test IDs
}

// Parameter describes a configurable test parameter.
type Parameter struct {
	Name       string // "Frame Sizes"
	Flag       string // "--frame-sizes"
	Type       string // "comma-separated integers"
	Default    string // "64,128,256,512,1024,1280,1518"
	Required   bool
	TechDesc   string // Technical description
	LaymanDesc string // Plain English description
	Example    string // "--frame-sizes 64,512,1518"
}

// Metric describes an output metric from a test.
type Metric struct {
	Name       string // "Max Rate"
	Unit       string // "% of line rate"
	GoodRange  string // ">95% is excellent, >80% is acceptable"
	BadMeaning string // "Below 80% indicates bottleneck"
}

// Example shows a practical usage example.
type Example struct {
	Desc    string // "Basic throughput test"
	Command string // "stem test -i eth0 -t throughput"
	Output  string // Expected output
}

// Issue documents a common problem and solution.
type Issue struct {
	Problem  string // "Test shows 0% throughput"
	Cause    string // "Interface not connected or wrong interface specified"
	Solution string // "Verify cable connection and interface name with 'ip link show'"
}

// CommandHelp documents a CLI command.
type CommandHelp struct {
	Name        string // "reflect"
	Summary     string // "Run packet reflection mode"
	Description string // Detailed description
	Usage       string // "stem reflect [flags]"
	Flags       []FlagHelp
	Examples    []Example
	SeeAlso     []string
}

// FlagHelp documents a command-line flag.
type FlagHelp struct {
	Short      string // "-i"
	Long       string // "--interface"
	Type       string // "string"
	Default    string
	Required   bool
	TechDesc   string
	LaymanDesc string
}

// GlossaryEntry defines a network term.
type GlossaryEntry struct {
	Term      string   // "CIR"
	FullName  string   // "Committed Information Rate"
	TechDef   string   // Technical definition
	LaymanDef string   // Plain English definition
	Related   []string // Related terms
}

// Tutorial provides step-by-step guidance.
type Tutorial struct {
	ID          string         // "quickstart"
	Title       string         // "Your First Test in 5 Minutes"
	Duration    string         // "5 min read"
	Level       string         // "Beginner", "Intermediate", "Advanced"
	Description string         // Brief description
	Steps       []TutorialStep // Ordered steps
}

// TutorialStep is a single step in a tutorial.
type TutorialStep struct {
	Title    string // Step title
	Content  string // Markdown content
	Command  string // Optional CLI command
	Expected string // Expected output
	Tip      string // Pro tip
}

// ErrorHelp provides context for error messages.
type ErrorHelp struct {
	Code       string    // "ERR_INTERFACE_REQUIRED"
	Message    string    // "Network interface is required"
	Cause      string    // Why this happens
	Solution   string    // How to fix it
	Examples   []Example // Example commands
	RelatedCmd string    // "stem help reflect"
}

// Category groups tests.
type Category struct {
	ID          string   // "rfc2544"
	Name        string   // "RFC 2544"
	FullName    string   // "Benchmarking Methodology for Network Interconnect Devices"
	Summary     string   // Brief overview
	Description string   // Full description
	Tests       []string // Test IDs in this category
	WhenToUse   string   // When to use this test suite
	Standard    string   // Standard reference
	SeeAlso     []string // Related categories
}

// System is the main entry point for help content.
type System struct {
	Tests      map[string]TestHelp
	Commands   map[string]CommandHelp
	Glossary   map[string]GlossaryEntry
	Tutorials  map[string]Tutorial
	Errors     map[string]ErrorHelp
	Categories map[string]Category
}

// NewSystem creates and initializes the help system.
func NewSystem() *System {
	return &System{
		Tests:      GetAllTests(),
		Commands:   GetAllCommands(),
		Glossary:   GetGlossary(),
		Tutorials:  GetAllTutorials(),
		Errors:     GetAllErrors(),
		Categories: GetAllCategories(),
	}
}

// GetTest returns help for a specific test.
func (h *System) GetTest(id string) (TestHelp, bool) {
	test, ok := h.Tests[id]
	return test, ok
}

// GetCommand returns help for a specific command.
func (h *System) GetCommand(name string) (CommandHelp, bool) {
	cmd, ok := h.Commands[name]
	return cmd, ok
}

// GetGlossaryTerm returns definition for a term.
func (h *System) GetGlossaryTerm(term string) (GlossaryEntry, bool) {
	entry, ok := h.Glossary[term]
	return entry, ok
}

// GetTutorial returns a specific tutorial.
func (h *System) GetTutorial(id string) (Tutorial, bool) {
	tut, ok := h.Tutorials[id]
	return tut, ok
}

// GetError returns help for an error code.
func (h *System) GetError(code string) (ErrorHelp, bool) {
	err, ok := h.Errors[code]
	return err, ok
}

// GetCategory returns a test category.
func (h *System) GetCategory(id string) (Category, bool) {
	cat, ok := h.Categories[id]
	return cat, ok
}

// GetTestsByCategory returns all tests in a category.
func (h *System) GetTestsByCategory(categoryID string) []TestHelp {
	var tests []TestHelp
	for _, test := range h.Tests {
		if test.Category == categoryID {
			tests = append(tests, test)
		}
	}
	return tests
}

// SearchTests searches tests by keyword.
func (h *System) SearchTests(keyword string) []TestHelp {
	var results []TestHelp
	for _, test := range h.Tests {
		if containsIgnoreCase(test.Name, keyword) ||
			containsIgnoreCase(test.Summary, keyword) ||
			containsIgnoreCase(test.TechDesc, keyword) ||
			containsIgnoreCase(test.LaymanDesc, keyword) {
			results = append(results, test)
		}
	}
	return results
}

// SearchGlossary searches glossary by keyword.
func (h *System) SearchGlossary(keyword string) []GlossaryEntry {
	var results []GlossaryEntry
	for _, entry := range h.Glossary {
		if containsIgnoreCase(entry.Term, keyword) ||
			containsIgnoreCase(entry.FullName, keyword) ||
			containsIgnoreCase(entry.TechDef, keyword) ||
			containsIgnoreCase(entry.LaymanDef, keyword) {
			results = append(results, entry)
		}
	}
	return results
}

// containsIgnoreCase checks if s contains substr (case-insensitive).
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(substr) == 0 ||
			findIgnoreCase(s, substr) >= 0)
}

func findIgnoreCase(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	for i := range len(s) - len(substr) + 1 {
		match := true
		for j := range len(substr) {
			sc := s[i+j]
			pc := substr[j]
			if sc >= 'A' && sc <= 'Z' {
				sc += 'a' - 'A'
			}
			if pc >= 'A' && pc <= 'Z' {
				pc += 'a' - 'A'
			}
			if sc != pc {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
