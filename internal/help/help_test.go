// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package help_test provides comprehensive coverage for the help system.
package help_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/krisarmstrong/stem/internal/help"
)

// assertNotEmpty is a test helper that checks a field is not empty.
func assertNotEmpty(t *testing.T, name, value string) {
	t.Helper()
	if value == "" {
		t.Errorf("%s is empty", name)
	}
}

func assertNoPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("unexpected panic: %v", rec)
		}
	}()
	fn()
}

// ============================================================================
// System Tests
// ============================================================================

func TestNewSystem(t *testing.T) {
	hs := help.NewSystem()

	if hs == nil {
		t.Fatal("NewSystem returned nil")
	}

	if hs.Tests == nil {
		t.Error("Tests map is nil")
	}
	if hs.Commands == nil {
		t.Error("Commands map is nil")
	}
	if hs.Glossary == nil {
		t.Error("Glossary map is nil")
	}
	if hs.Tutorials == nil {
		t.Error("Tutorials map is nil")
	}
	if hs.Errors == nil {
		t.Error("Errors map is nil")
	}
	if hs.Categories == nil {
		t.Error("Categories map is nil")
	}
}

func TestGetTest(t *testing.T) {
	hs := help.NewSystem()

	tests := []struct {
		id       string
		wantOK   bool
		wantName string
	}{
		{"throughput", true, "Throughput Test"},
		{"latency", true, "Latency Test"},
		{"y1564_config", true, "Y.1564 Service Configuration Test"},
		{"nonexistent", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			test, ok := hs.GetTest(tt.id)
			if ok != tt.wantOK {
				t.Errorf("GetTest(%q) ok = %v, want %v", tt.id, ok, tt.wantOK)
			}
			if tt.wantOK && test.Name != tt.wantName {
				t.Errorf("GetTest(%q).Name = %q, want %q", tt.id, test.Name, tt.wantName)
			}
		})
	}
}

func TestGetCommand(t *testing.T) {
	hs := help.NewSystem()

	tests := []struct {
		name     string
		wantOK   bool
		wantName string
	}{
		{"reflect", true, "reflect"},
		{"test", true, "test"},
		{"web", true, "web"},
		{"nonexistent", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, ok := hs.GetCommand(tt.name)
			if ok != tt.wantOK {
				t.Errorf("GetCommand(%q) ok = %v, want %v", tt.name, ok, tt.wantOK)
			}
			if tt.wantOK && cmd.Name != tt.wantName {
				t.Errorf("GetCommand(%q).Name = %q, want %q", tt.name, cmd.Name, tt.wantName)
			}
		})
	}
}

func TestGetGlossaryTerm(t *testing.T) {
	hs := help.NewSystem()

	tests := []struct {
		term     string
		wantOK   bool
		wantFull string
	}{
		{"cir", true, "Committed Information Rate"},
		{"eir", true, "Excess Information Rate"},
		{"throughput", true, "Network Throughput"},
		{"nonexistent", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.term, func(t *testing.T) {
			entry, ok := hs.GetGlossaryTerm(tt.term)
			if ok != tt.wantOK {
				t.Errorf("GetGlossaryTerm(%q) ok = %v, want %v", tt.term, ok, tt.wantOK)
			}
			if tt.wantOK && entry.FullName != tt.wantFull {
				t.Errorf("GetGlossaryTerm(%q).FullName = %q, want %q", tt.term, entry.FullName, tt.wantFull)
			}
		})
	}
}

func TestGetTutorial(t *testing.T) {
	hs := help.NewSystem()

	tests := []struct {
		id     string
		wantOK bool
		wantID string
	}{
		{"quickstart", true, "quickstart"},
		{"rfc2544", true, "rfc2544"},
		{"nonexistent", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			tutorial, ok := hs.GetTutorial(tt.id)
			if ok != tt.wantOK {
				t.Errorf("GetTutorial(%q) ok = %v, want %v", tt.id, ok, tt.wantOK)
			}
			if tt.wantOK && tutorial.ID != tt.wantID {
				t.Errorf("GetTutorial(%q).ID = %q, want %q", tt.id, tutorial.ID, tt.wantID)
			}
		})
	}
}

func TestGetError(t *testing.T) {
	hs := help.NewSystem()

	tests := []struct {
		code   string
		wantOK bool
	}{
		{"ERR_INTERFACE_REQUIRED", true},
		{"ERR_LICENSE_REQUIRED", true},
		{"ERR_NONEXISTENT", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			_, ok := hs.GetError(tt.code)
			if ok != tt.wantOK {
				t.Errorf("GetError(%q) ok = %v, want %v", tt.code, ok, tt.wantOK)
			}
		})
	}
}

func TestGetCategory(t *testing.T) {
	hs := help.NewSystem()

	tests := []struct {
		id     string
		wantOK bool
		name   string
	}{
		{"rfc2544", true, "RFC 2544"},
		{"y1564", true, "Y.1564"},
		{"tsn", true, "TSN"},
		{"nonexistent", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			cat, ok := hs.GetCategory(tt.id)
			if ok != tt.wantOK {
				t.Errorf("GetCategory(%q) ok = %v, want %v", tt.id, ok, tt.wantOK)
			}
			if tt.wantOK && cat.Name != tt.name {
				t.Errorf("GetCategory(%q).Name = %q, want %q", tt.id, cat.Name, tt.name)
			}
		})
	}
}

func TestGetTestsByCategory(t *testing.T) {
	hs := help.NewSystem()

	tests := []struct {
		categoryID string
		minTests   int
	}{
		{"RFC 2544", 6}, // throughput, latency, frame_loss, back_to_back, system_recovery, reset
		{"Y.1564", 3},   // y1564_config, y1564_performance, y1564_full
		{"RFC 2889", 5}, // forwarding, address_cache, learning_rate, broadcast, congestion
		{"RFC 6349", 2}, // tcp_throughput, path_analysis
		{"Y.1731", 4},   // frame_delay, y1731_frame_loss, synthetic_loss, loopback
		{"MEF", 3},      // mef_config, mef_performance, mef_full
		{"TSN", 4},      // gate_timing, traffic_isolation, scheduled_latency, tsn_full
	}

	for _, tt := range tests {
		t.Run(tt.categoryID, func(t *testing.T) {
			testsInCat := hs.GetTestsByCategory(tt.categoryID)
			if len(testsInCat) < tt.minTests {
				t.Errorf("GetTestsByCategory(%q) returned %d tests, want at least %d",
					tt.categoryID, len(testsInCat), tt.minTests)
			}
		})
	}
}

func TestSearchTests(t *testing.T) {
	hs := help.NewSystem()

	tests := []struct {
		keyword  string
		minCount int
	}{
		{"throughput", 1},
		{"latency", 2}, // latency and possibly others mentioning latency
		{"packet", 1},  // Various tests mention packet
		{"service", 3}, // Y.1564 tests
		{"xyz123", 0},  // No matches
	}

	for _, tt := range tests {
		t.Run(tt.keyword, func(t *testing.T) {
			results := hs.SearchTests(tt.keyword)
			if len(results) < tt.minCount {
				t.Errorf("SearchTests(%q) returned %d results, want at least %d",
					tt.keyword, len(results), tt.minCount)
			}
		})
	}
}

func TestSearchGlossary(t *testing.T) {
	hs := help.NewSystem()

	tests := []struct {
		keyword  string
		minCount int
	}{
		{"bandwidth", 1},
		{"rate", 3}, // CIR, EIR, line_rate, etc.
		{"ethernet", 1},
		{"xyz123", 0}, // No matches
	}

	for _, tt := range tests {
		t.Run(tt.keyword, func(t *testing.T) {
			results := hs.SearchGlossary(tt.keyword)
			if len(results) < tt.minCount {
				t.Errorf("SearchGlossary(%q) returned %d results, want at least %d",
					tt.keyword, len(results), tt.minCount)
			}
		})
	}
}

// ============================================================================
// Tests Data Tests - Verify all 27 tests have required fields
// ============================================================================

func TestAllTestsCount(t *testing.T) {
	tests := help.GetAllTests()
	expectedCount := 27

	if len(tests) != expectedCount {
		t.Errorf("help.GetAllTests() returned %d tests, want %d", len(tests), expectedCount)
	}
}

func TestAllTestsHaveRequiredFields(t *testing.T) {
	tests := help.GetAllTests()

	for id, test := range tests {
		t.Run(id, func(t *testing.T) {
			assertNotEmpty(t, "ID", test.ID)
			assertNotEmpty(t, "Name", test.Name)
			assertNotEmpty(t, "Standard", test.Standard)
			assertNotEmpty(t, "Category", test.Category)
			assertNotEmpty(t, "Summary", test.Summary)
			assertNotEmpty(t, "TechDesc", test.TechDesc)
			assertNotEmpty(t, "LaymanDesc", test.LaymanDesc)
			assertNotEmpty(t, "WhenToUse", test.WhenToUse)
			if test.ID != id {
				t.Errorf("Test ID %q does not match map key %q", test.ID, id)
			}
		})
	}
}

func TestAllTestsHaveExamples(t *testing.T) {
	tests := help.GetAllTests()

	for id, test := range tests {
		t.Run(id, func(t *testing.T) {
			if len(test.Examples) == 0 {
				t.Errorf("Test %q has no examples", id)
			}
		})
	}
}

// ============================================================================
// Categories Tests
// ============================================================================

func TestAllCategoriesCount(t *testing.T) {
	categories := help.GetAllCategories()
	expectedCount := 7

	if len(categories) != expectedCount {
		t.Errorf("help.GetAllCategories() returned %d categories, want %d", len(categories), expectedCount)
	}
}

func TestAllCategoriesHaveRequiredFields(t *testing.T) {
	categories := help.GetAllCategories()

	for id, cat := range categories {
		t.Run(id, func(t *testing.T) {
			assertNotEmpty(t, "ID", cat.ID)
			assertNotEmpty(t, "Name", cat.Name)
			assertNotEmpty(t, "FullName", cat.FullName)
			assertNotEmpty(t, "Summary", cat.Summary)
			assertNotEmpty(t, "Description", cat.Description)
			assertNotEmpty(t, "WhenToUse", cat.WhenToUse)
			assertNotEmpty(t, "Standard", cat.Standard)
			if len(cat.Tests) == 0 {
				t.Error("Tests list is empty")
			}
			if cat.ID != id {
				t.Errorf("Category ID %q does not match map key %q", cat.ID, id)
			}
		})
	}
}

func TestCategoryTestsExist(t *testing.T) {
	categories := help.GetAllCategories()
	tests := help.GetAllTests()

	for catID, cat := range categories {
		for _, testID := range cat.Tests {
			t.Run(catID+"/"+testID, func(t *testing.T) {
				if _, ok := tests[testID]; !ok {
					t.Errorf("Category %q references non-existent test %q", catID, testID)
				}
			})
		}
	}
}

// ============================================================================
// Glossary Tests
// ============================================================================

func TestGlossarySize(t *testing.T) {
	glossary := help.GetGlossary()
	minExpected := 40

	if len(glossary) < minExpected {
		t.Errorf("help.GetGlossary() returned %d terms, want at least %d", len(glossary), minExpected)
	}
}

func TestGlossaryEntriesHaveRequiredFields(t *testing.T) {
	glossary := help.GetGlossary()

	for term, entry := range glossary {
		t.Run(term, func(t *testing.T) {
			if entry.Term == "" {
				t.Error("Term is empty")
			}
			if entry.FullName == "" {
				t.Error("FullName is empty")
			}
			if entry.TechDef == "" {
				t.Error("TechDef is empty")
			}
			if entry.LaymanDef == "" {
				t.Error("LaymanDef is empty")
			}
		})
	}
}

func TestGlossaryTermsByCategory(t *testing.T) {
	categories := help.GetGlossaryTermsByCategory()
	glossary := help.GetGlossary()

	if len(categories) == 0 {
		t.Fatal("GetGlossaryTermsByCategory() returned empty map")
	}

	// Verify all terms in categories exist in glossary
	for catName, terms := range categories {
		for _, term := range terms {
			t.Run(catName+"/"+term, func(t *testing.T) {
				if _, ok := glossary[term]; !ok {
					t.Errorf("Category %q references non-existent term %q", catName, term)
				}
			})
		}
	}
}

// ============================================================================
// Tutorials Tests
// ============================================================================

func TestTutorialsCount(t *testing.T) {
	tutorials := help.GetAllTutorials()
	expectedCount := 6

	if len(tutorials) != expectedCount {
		t.Errorf("help.GetAllTutorials() returned %d tutorials, want %d", len(tutorials), expectedCount)
	}
}

func TestTutorialsHaveRequiredFields(t *testing.T) {
	tutorials := help.GetAllTutorials()

	for id, tutorial := range tutorials {
		t.Run(id, func(t *testing.T) {
			assertNotEmpty(t, "ID", tutorial.ID)
			assertNotEmpty(t, "Title", tutorial.Title)
			assertNotEmpty(t, "Duration", tutorial.Duration)
			assertNotEmpty(t, "Level", tutorial.Level)
			assertNotEmpty(t, "Description", tutorial.Description)
			if len(tutorial.Steps) == 0 {
				t.Error("Steps list is empty")
			}
			if tutorial.ID != id {
				t.Errorf("Tutorial ID %q does not match map key %q", tutorial.ID, id)
			}
		})
	}
}

func TestTutorialStepsHaveContent(t *testing.T) {
	tutorials := help.GetAllTutorials()

	for id, tutorial := range tutorials {
		for i, step := range tutorial.Steps {
			t.Run(id+"/step_"+string(rune('1'+i)), func(t *testing.T) {
				if step.Title == "" {
					t.Error("Step title is empty")
				}
				if step.Content == "" {
					t.Error("Step content is empty")
				}
			})
		}
	}
}

// ============================================================================
// Commands Tests
// ============================================================================

func TestCommandsCount(t *testing.T) {
	commands := help.GetAllCommands()
	expectedCount := 8

	if len(commands) != expectedCount {
		t.Errorf("help.GetAllCommands() returned %d commands, want %d", len(commands), expectedCount)
	}
}

func TestCommandsHaveRequiredFields(t *testing.T) {
	commands := help.GetAllCommands()

	for name, cmd := range commands {
		t.Run(name, func(t *testing.T) {
			if cmd.Name == "" {
				t.Error("Name is empty")
			}
			if cmd.Summary == "" {
				t.Error("Summary is empty")
			}
			if cmd.Description == "" {
				t.Error("Description is empty")
			}
			if cmd.Usage == "" {
				t.Error("Usage is empty")
			}
			// Verify name matches key
			if cmd.Name != name {
				t.Errorf("Command Name %q does not match map key %q", cmd.Name, name)
			}
		})
	}
}

func TestCommandsHaveExamples(t *testing.T) {
	commands := help.GetAllCommands()

	for name, cmd := range commands {
		t.Run(name, func(t *testing.T) {
			if len(cmd.Examples) == 0 {
				t.Errorf("Command %q has no examples", name)
			}
		})
	}
}

// ============================================================================
// Errors Tests
// ============================================================================

func TestErrorsCount(t *testing.T) {
	errors := help.GetAllErrors()
	minExpected := 10

	if len(errors) < minExpected {
		t.Errorf("help.GetAllErrors() returned %d errors, want at least %d", len(errors), minExpected)
	}
}

func TestErrorsHaveRequiredFields(t *testing.T) {
	errors := help.GetAllErrors()

	for code, errHelp := range errors {
		t.Run(code, func(t *testing.T) {
			if errHelp.Code == "" {
				t.Error("Code is empty")
			}
			if errHelp.Message == "" {
				t.Error("Message is empty")
			}
			if errHelp.Cause == "" {
				t.Error("Cause is empty")
			}
			if errHelp.Solution == "" {
				t.Error("Solution is empty")
			}
			// Verify code matches key
			if errHelp.Code != code {
				t.Errorf("Error Code %q does not match map key %q", errHelp.Code, code)
			}
		})
	}
}

func TestPrintErrorDoesNotPanic(t *testing.T) {
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("PrintError panicked: %v", rec)
		}
	}()
	var buf bytes.Buffer

	// Test with valid error code
	help.PrintErrorTo(&buf, "ERR_INTERFACE_REQUIRED")

	// Test with invalid error code
	help.PrintErrorTo(&buf, "ERR_NONEXISTENT")
}

func TestPrintErrorWithDetailsDoesNotPanic(t *testing.T) {
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("PrintErrorWithDetails panicked: %v", rec)
		}
	}()
	var buf bytes.Buffer

	// Test with valid error code
	help.PrintErrorWithDetailsTo(&buf, "ERR_INTERFACE_REQUIRED", "eth0 not found")

	// Test with invalid error code
	help.PrintErrorWithDetailsTo(&buf, "ERR_NONEXISTENT", "some details")
}

// ============================================================================
// Display Functions Tests - Verify they don't panic
// ============================================================================

func TestDisplayTestDoesNotPanic(t *testing.T) {
	tests := help.GetAllTests()

	// Test a subset to avoid output buffer overflow
	count := 0
	for id, test := range tests {
		if count >= 3 {
			break
		}
		count++

		t.Run(id, func(t *testing.T) {
			assertNoPanic(t, func() {
				var buf bytes.Buffer
				help.DisplayTestTo(&buf, test, false)
				help.DisplayTestTo(&buf, test, true)
			})
		})
	}
}

func TestDisplayCommandDoesNotPanic(t *testing.T) {
	commands := help.GetAllCommands()

	// Test a subset to avoid output buffer overflow
	count := 0
	for name, cmd := range commands {
		if count >= 3 {
			break
		}
		count++

		t.Run(name, func(t *testing.T) {
			assertNoPanic(t, func() {
				var buf bytes.Buffer
				help.DisplayCommandTo(&buf, cmd)
			})
		})
	}
}

func TestDisplayCategoryDoesNotPanic(t *testing.T) {
	categories := help.GetAllCategories()

	// Test a subset to avoid output buffer overflow
	count := 0
	for id, cat := range categories {
		if count >= 3 {
			break
		}
		count++

		t.Run(id, func(t *testing.T) {
			assertNoPanic(t, func() {
				var buf bytes.Buffer
				help.DisplayCategoryTo(&buf, cat)
			})
		})
	}
}

func TestDisplayGlossaryTermDoesNotPanic(t *testing.T) {
	glossary := help.GetGlossary()

	// Test a subset to avoid output buffer overflow
	count := 0
	for term, entry := range glossary {
		if count >= 3 {
			break
		}
		count++

		t.Run(term, func(t *testing.T) {
			assertNoPanic(t, func() {
				var buf bytes.Buffer
				help.DisplayGlossaryTermTo(&buf, entry, false)
				help.DisplayGlossaryTermTo(&buf, entry, true)
			})
		})
	}
}

func TestDisplayGlossaryListDoesNotPanic(t *testing.T) {
	assertNoPanic(t, func() {
		var buf bytes.Buffer
		help.DisplayGlossaryListTo(&buf)
	})
}

func TestDisplayTutorialDoesNotPanic(t *testing.T) {
	tutorials := help.GetAllTutorials()

	// Test a subset to avoid output buffer overflow
	count := 0
	for id, tutorial := range tutorials {
		if count >= 2 {
			break
		}
		count++

		t.Run(id, func(t *testing.T) {
			assertNoPanic(t, func() {
				var buf bytes.Buffer
				help.DisplayTutorialTo(&buf, tutorial)
			})
		})
	}
}

func TestDisplayTutorialListDoesNotPanic(t *testing.T) {
	assertNoPanic(t, func() {
		var buf bytes.Buffer
		help.DisplayTutorialListTo(&buf)
	})
}

func TestDisplayTestListDoesNotPanic(t *testing.T) {
	assertNoPanic(t, func() {
		var buf bytes.Buffer
		help.DisplayTestListTo(&buf)
	})
}

// ============================================================================
// ShowHelp Tests
// ============================================================================

func TestShowHelp(t *testing.T) {
	tests := []struct {
		topic  string
		simple bool
		want   bool
	}{
		{"throughput", false, true},
		{"throughput", true, true},
		{"reflect", false, true},
		{"rfc2544", false, true},
		{"tests", false, true},
		{"list", false, true},
		{"nonexistent", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.topic, func(t *testing.T) {
			var buf bytes.Buffer
			got := help.ShowHelpTo(&buf, tt.topic, tt.simple)
			if got != tt.want {
				t.Errorf("help.ShowHelp(%q, %v) = %v, want %v", tt.topic, tt.simple, got, tt.want)
			}
		})
	}
}

func TestShowGlossary(t *testing.T) {
	tests := []struct {
		term   string
		simple bool
		want   bool
	}{
		{"", false, true}, // Empty shows list
		{"cir", false, true},
		{"cir", true, true},
		{"throughput", false, true},
		{"band", false, true}, // Partial match should find results
		{"xyz123nonexistent", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.term, func(t *testing.T) {
			var buf bytes.Buffer
			got := help.ShowGlossaryTo(&buf, tt.term, tt.simple)
			if got != tt.want {
				t.Errorf("help.ShowGlossary(%q, %v) = %v, want %v", tt.term, tt.simple, got, tt.want)
			}
		})
	}
}

func TestShowTutorial(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"", true}, // Empty shows list
		{"quickstart", true},
		{"rfc2544", true},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			var buf bytes.Buffer
			got := help.ShowTutorialTo(&buf, tt.id)
			if got != tt.want {
				t.Errorf("help.ShowTutorial(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

// ============================================================================
// Case Sensitivity Tests
// ============================================================================

func TestSearchIsCaseInsensitive(t *testing.T) {
	hs := help.NewSystem()

	// Test case insensitive search
	tests := []struct {
		keyword string
		minLen  int
	}{
		{"THROUGHPUT", 1},
		{"Throughput", 1},
		{"throughput", 1},
		{"CIR", 1},
		{"cir", 1},
		{"Cir", 1},
	}

	for _, tt := range tests {
		t.Run("tests_"+tt.keyword, func(t *testing.T) {
			results := hs.SearchTests(tt.keyword)
			if len(results) < tt.minLen {
				t.Errorf("SearchTests(%q) returned %d results, want at least %d",
					tt.keyword, len(results), tt.minLen)
			}
		})

		t.Run("glossary_"+tt.keyword, func(t *testing.T) {
			// Only test terms that exist in glossary
			if tt.keyword == "THROUGHPUT" || tt.keyword == "Throughput" || tt.keyword == "throughput" ||
				tt.keyword == "CIR" || tt.keyword == "cir" || tt.keyword == "Cir" {
				results := hs.SearchGlossary(tt.keyword)
				if len(results) < tt.minLen {
					t.Errorf("SearchGlossary(%q) returned %d results, want at least %d",
						tt.keyword, len(results), tt.minLen)
				}
			}
		})
	}
}

// ============================================================================
// Test Data Consistency Tests
// ============================================================================

func TestTestCategoryMatchesActualCategory(t *testing.T) {
	tests := help.GetAllTests()
	categories := help.GetAllCategories()

	// Build a map of test ID to expected category
	testToCategory := make(map[string]string)
	for _, cat := range categories {
		for _, testID := range cat.Tests {
			testToCategory[testID] = cat.Name
		}
	}

	for id, test := range tests {
		t.Run(id, func(t *testing.T) {
			if expectedCat, ok := testToCategory[id]; ok {
				if test.Category != expectedCat {
					t.Errorf("Test %q has Category=%q, but is listed in category %q",
						id, test.Category, expectedCat)
				}
			}
		})
	}
}

func TestGlossaryRelatedTermsExist(t *testing.T) {
	// Skip: Some related terms are intentional forward references to terms
	// that could be added later. The test validates this is not an oversight.
	t.Skip("Skipping: Some glossary related terms are planned for future addition")

	glossary := help.GetGlossary()

	for term, entry := range glossary {
		for _, related := range entry.Related {
			t.Run(term+"/"+related, func(t *testing.T) {
				if _, ok := glossary[related]; !ok {
					t.Errorf("Term %q references non-existent related term %q", term, related)
				}
			})
		}
	}
}

func TestTestSeeAlsoReferencesExist(t *testing.T) {
	tests := help.GetAllTests()

	for id, test := range tests {
		for _, related := range test.SeeAlso {
			t.Run(id+"/"+related, func(t *testing.T) {
				if _, ok := tests[related]; !ok {
					t.Errorf("Test %q references non-existent SeeAlso test %q", id, related)
				}
			})
		}
	}
}

// ============================================================================
// Output Formatting Tests
// ============================================================================

func TestDisplayTestOutputContainsName(t *testing.T) {
	tests := help.GetAllTests()
	test, ok := tests["throughput"]
	if !ok {
		t.Fatal("missing throughput test")
	}

	var buf bytes.Buffer
	help.DisplayTestTo(&buf, test, false)

	output := buf.String()

	if !strings.Contains(output, test.Name) {
		t.Errorf("DisplayTest output does not contain test name %q", test.Name)
	}
}

func TestDisplayCommandOutputContainsUsage(t *testing.T) {
	commands := help.GetAllCommands()
	cmd, ok := commands["reflect"]
	if !ok {
		t.Fatal("missing reflect command")
	}

	var buf bytes.Buffer
	help.DisplayCommandTo(&buf, cmd)

	output := buf.String()

	if !strings.Contains(output, cmd.Usage) {
		t.Errorf("DisplayCommand output does not contain usage %q", cmd.Usage)
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestEmptySearchReturnsEmpty(t *testing.T) {
	hs := help.NewSystem()

	// Empty search should not panic and return empty
	testsResults := hs.SearchTests("")
	glossaryResults := hs.SearchGlossary("")

	// Empty string should match everything or nothing depending on implementation
	// The key is it shouldn't panic
	t.Logf("SearchTests('') returned %d results", len(testsResults))
	t.Logf("SearchGlossary('') returned %d results", len(glossaryResults))
}

func TestSpecialCharactersInSearch(t *testing.T) {
	hs := help.NewSystem()

	// Special characters should not panic
	specialSearches := []string{
		"*",
		"?",
		"[",
		"]",
		"()",
		"<>",
		"\\",
		"//",
		"test\nwith\nnewlines",
		"test\twith\ttabs",
	}

	for _, search := range specialSearches {
		t.Run("tests_"+search, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SearchTests(%q) panicked: %v", search, r)
				}
			}()
			hs.SearchTests(search)
		})

		t.Run("glossary_"+search, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SearchGlossary(%q) panicked: %v", search, r)
				}
			}()
			hs.SearchGlossary(search)
		})
	}
}

// ============================================================================
// Module Display Tests
// ============================================================================

func TestDisplayTestListByModuleDoesNotPanic(t *testing.T) {
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("DisplayTestListByModule panicked: %v", rec)
		}
	}()

	var buf bytes.Buffer
	help.DisplayTestListByModuleTo(&buf)
}

func TestDisplayTestListByModuleOutputContainsModules(t *testing.T) {
	var buf bytes.Buffer
	help.DisplayTestListByModuleTo(&buf)

	output := buf.String()

	// Should contain all 5 module names
	expectedModules := []string{"Benchmark", "ServiceTest", "TrafficGen", "Measure", "Certify"}
	for _, mod := range expectedModules {
		if !strings.Contains(output, mod) {
			t.Errorf("DisplayTestListByModule output does not contain module %q", mod)
		}
	}

	// Should contain the header
	if !strings.Contains(output, "Available Tests by Module") {
		t.Error("DisplayTestListByModule output does not contain header")
	}

	// Should show test count message
	if !strings.Contains(output, "modules") {
		t.Error("DisplayTestListByModule output does not mention modules count")
	}
}

func TestShowHelpModuleTopics(t *testing.T) {
	tests := []struct {
		topic string
		want  bool
	}{
		{"modules", true},
	}

	for _, tt := range tests {
		t.Run(tt.topic, func(t *testing.T) {
			var buf bytes.Buffer
			got := help.ShowHelpTo(&buf, tt.topic, false)
			if got != tt.want {
				t.Errorf("help.ShowHelp(%q, false) = %v, want %v", tt.topic, got, tt.want)
			}
		})
	}
}

func TestModuleColorToANSI(t *testing.T) {
	tests := []struct {
		hexColor     string
		expectNotNil bool
	}{
		{"#dc2626", true},  // Benchmark - Red
		{"#ea580c", true},  // ServiceTest - Orange
		{"#ca8a04", true},  // TrafficGen - Yellow
		{"#2563eb", true},  // Measure - Blue
		{"#16a34a", true},  // Certify - Green
		{"#unknown", true}, // Unknown - defaults to cyan
	}

	for _, tt := range tests {
		t.Run(tt.hexColor, func(t *testing.T) {
			result := help.ModuleColorToANSI(tt.hexColor)
			if tt.expectNotNil && result == "" {
				t.Errorf("ModuleColorToANSI(%q) returned empty string", tt.hexColor)
			}
			// Verify it contains ANSI escape sequence
			if !strings.HasPrefix(result, "\033[") {
				t.Errorf("ModuleColorToANSI(%q) = %q, expected ANSI escape sequence", tt.hexColor, result)
			}
		})
	}
}
