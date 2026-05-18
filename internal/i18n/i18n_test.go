/*
 * The Stem - i18n Tests
 */

package i18n_test

import (
	"testing"

	"github.com/krisarmstrong/stem/internal/i18n"
)

func TestT(t *testing.T) {
	// Ensure default language works
	i18n.SetLanguage(i18n.English)

	tests := []struct {
		key      string
		expected string
	}{
		{"app.title", "The Stem"},
		{"status.running", "Running"},
		{"results.pass", "PASS"},
		{"labels.dashboard", "Dashboard"},
	}

	for _, tt := range tests {
		got := i18n.T(tt.key)
		if got != tt.expected {
			t.Errorf("T(%q) = %q, want %q", tt.key, got, tt.expected)
		}
	}
}

func TestTWithSpanish(t *testing.T) {
	i18n.SetLanguage(i18n.Spanish)
	defer i18n.SetLanguage(i18n.English)

	tests := []struct {
		key      string
		expected string
	}{
		{"app.title", "The Stem"},
		{"status.running", "Ejecutando"},
		{"results.pass", "APROBADO"},
		{"labels.dashboard", "Panel de Control"},
	}

	for _, tt := range tests {
		got := i18n.T(tt.key)
		if got != tt.expected {
			t.Errorf("T(%q) = %q, want %q", tt.key, got, tt.expected)
		}
	}
}

func TestTMissingKey(t *testing.T) {
	i18n.SetLanguage(i18n.English)

	// Unknown key should return the key itself
	key := "unknown.key.that.does.not.exist"
	got := i18n.T(key)
	if got != key {
		t.Errorf("T(%q) = %q, want %q (fallback to key)", key, got, key)
	}
}

func TestTL(t *testing.T) {
	tests := []struct {
		key      string
		lang     i18n.Language
		expected string
	}{
		{"status.running", i18n.English, "Running"},
		{"status.running", i18n.Spanish, "Ejecutando"},
		{"results.pass", i18n.English, "PASS"},
		{"results.pass", i18n.Spanish, "APROBADO"},
		{"labels.dashboard", i18n.English, "Dashboard"},
		{"labels.dashboard", i18n.Spanish, "Panel de Control"},
	}

	for _, tt := range tests {
		got := i18n.TL(tt.key, tt.lang)
		if got != tt.expected {
			t.Errorf("TL(%q, %q) = %q, want %q", tt.key, tt.lang, got, tt.expected)
		}
	}
}

func TestSetLanguage(t *testing.T) {
	i18n.SetLanguage(i18n.Spanish)
	if i18n.GetLanguage() != i18n.Spanish {
		t.Errorf("GetLanguage() = %q, want %q", i18n.GetLanguage(), i18n.Spanish)
	}

	i18n.SetLanguage(i18n.English)
	if i18n.GetLanguage() != i18n.English {
		t.Errorf("GetLanguage() = %q, want %q", i18n.GetLanguage(), i18n.English)
	}
}

func TestSpanishFallback(t *testing.T) {
	i18n.SetLanguage(i18n.Spanish)
	defer i18n.SetLanguage(i18n.English)

	// All English keys should have Spanish translations
	// This is a spot check of important keys
	keysToCheck := []string{
		"app.title",
		"status.running",
		"status.completed",
		"status.failed",
		"results.pass",
		"results.fail",
		"labels.dashboard",
		"labels.tests",
		"labels.settings",
		"auth.invalidCredentials",
		"license.invalid",
	}

	for _, key := range keysToCheck {
		en := i18n.TL(key, i18n.English)
		es := i18n.TL(key, i18n.Spanish)

		// Both should return something other than the key
		if en == key {
			t.Errorf("English translation missing for %q", key)
		}
		if es == key {
			t.Errorf("Spanish translation missing for %q", key)
		}

		// For most keys, Spanish should be different from English
		// (app.title is an exception as it's a proper noun)
		if key != "app.title" && en == es {
			t.Logf("Warning: %q has same translation in English and Spanish: %q", key, en)
		}
	}
}

func TestSupportedLanguages(t *testing.T) {
	langs := i18n.SupportedLanguages()
	if len(langs) != 2 {
		t.Errorf("SupportedLanguages() returned %d languages, want 2", len(langs))
	}

	hasEnglish := false
	hasSpanish := false
	for _, l := range langs {
		if l == i18n.English {
			hasEnglish = true
		}
		if l == i18n.Spanish {
			hasSpanish = true
		}
	}

	if !hasEnglish {
		t.Error("English not in supported languages")
	}
	if !hasSpanish {
		t.Error("Spanish not in supported languages")
	}
}

func TestLanguageName(t *testing.T) {
	tests := []struct {
		lang     i18n.Language
		expected string
	}{
		{i18n.English, "English"},
		{i18n.Spanish, "Espanol"},
		{i18n.Language("fr"), "fr"}, // Unknown language returns the code
		{i18n.Language("de"), "de"}, // Unknown language returns the code
		{i18n.Language(""), ""},     // Empty language
	}

	for _, tt := range tests {
		got := i18n.LanguageName(tt.lang)
		if got != tt.expected {
			t.Errorf("LanguageName(%q) = %q, want %q", tt.lang, got, tt.expected)
		}
	}
}

func TestLanguageNativeName(t *testing.T) {
	tests := []struct {
		lang     i18n.Language
		expected string
	}{
		{i18n.English, "English"},
		{i18n.Spanish, "Espanol"},
		{i18n.Language("fr"), "fr"}, // Unknown language returns the code
		{i18n.Language("de"), "de"}, // Unknown language returns the code
		{i18n.Language(""), ""},     // Empty language
	}

	for _, tt := range tests {
		got := i18n.LanguageNativeName(tt.lang)
		if got != tt.expected {
			t.Errorf("LanguageNativeName(%q) = %q, want %q", tt.lang, got, tt.expected)
		}
	}
}

func TestTLWithFallback(t *testing.T) {
	// Test TL fallback to English for missing Spanish key
	// First, test with a key that exists in English but might not in Spanish
	key := "unknown.key.for.testing"
	got := i18n.TL(key, i18n.Spanish)
	// Should return the key itself since it doesn't exist in either language
	if got != key {
		t.Errorf("TL(%q, Spanish) = %q, want %q", key, got, key)
	}

	// Test with unknown language
	unknownLang := i18n.Language("xx")
	got = i18n.TL("status.running", unknownLang)
	// Should fall back to English
	if got != "Running" {
		t.Errorf("TL(\"status.running\", %q) = %q, want \"Running\"", unknownLang, got)
	}
}

func TestTLMissingKeyInBothLanguages(t *testing.T) {
	key := "completely.nonexistent.key"

	// Test with English
	got := i18n.TL(key, i18n.English)
	if got != key {
		t.Errorf("TL(%q, English) = %q, want %q", key, got, key)
	}

	// Test with Spanish
	got = i18n.TL(key, i18n.Spanish)
	if got != key {
		t.Errorf("TL(%q, Spanish) = %q, want %q", key, got, key)
	}

	// Test with unknown language
	got = i18n.TL(key, i18n.Language("unknown"))
	if got != key {
		t.Errorf("TL(%q, unknown) = %q, want %q", key, got, key)
	}
}

func TestSetLanguageInvalid(t *testing.T) {
	// Set to a valid language first
	i18n.SetLanguage(i18n.English)
	initial := i18n.GetLanguage()

	// Try to set an invalid language
	i18n.SetLanguage(i18n.Language("invalid"))

	// Should still be the initial language
	if i18n.GetLanguage() != initial {
		t.Errorf("SetLanguage with invalid language changed the language")
	}
}

func TestDetectLanguageWithEnvVars(t *testing.T) {
	// Use t.Setenv which automatically cleans up after the test
	t.Setenv("STEM_LANG", "es")
	t.Setenv("LANG", "en_US.UTF-8")

	i18n.SetLanguage(i18n.English) // Reset
	// Note: detectLanguage is called at init, so we can't test it directly
	// But we can verify the language codes work

	// Test Spanish locale parsing
	i18n.SetLanguage(i18n.Spanish)
	if i18n.GetLanguage() != i18n.Spanish {
		t.Errorf("GetLanguage() = %q, want %q", i18n.GetLanguage(), i18n.Spanish)
	}

	// Reset to English
	i18n.SetLanguage(i18n.English)
	if i18n.GetLanguage() != i18n.English {
		t.Errorf("GetLanguage() = %q, want %q", i18n.GetLanguage(), i18n.English)
	}
}

func TestParseLanguageVariants(t *testing.T) {
	// These tests verify that various locale formats are handled
	// We can test this indirectly through SetLanguage and GetLanguage
	tests := []struct {
		name     string
		langCode i18n.Language
		valid    bool
	}{
		{"English", i18n.English, true},
		{"Spanish", i18n.Spanish, true},
		{"French (unsupported)", i18n.Language("fr"), false},
		{"German (unsupported)", i18n.Language("de"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initial := i18n.GetLanguage()
			i18n.SetLanguage(tt.langCode)
			if tt.valid {
				if i18n.GetLanguage() != tt.langCode {
					t.Errorf("SetLanguage(%q) did not change language", tt.langCode)
				}
			} else {
				if i18n.GetLanguage() != initial {
					t.Errorf("SetLanguage(%q) should not change language for unsupported lang", tt.langCode)
				}
			}
			i18n.SetLanguage(i18n.English) // Reset
		})
	}
}

func TestConcurrentAccess(_ *testing.T) {
	// Test thread safety of language operations
	done := make(chan bool, 10)

	for range 5 {
		go func() {
			for range 100 {
				i18n.SetLanguage(i18n.English)
				_ = i18n.GetLanguage()
				_ = i18n.T("app.title")
			}
			done <- true
		}()
	}

	for range 5 {
		go func() {
			for range 100 {
				i18n.SetLanguage(i18n.Spanish)
				_ = i18n.GetLanguage()
				_ = i18n.TL("app.title", i18n.Spanish)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for range 10 {
		<-done
	}
}

func TestAllEnglishKeysHaveTranslations(t *testing.T) {
	// Verify that common keys return expected values
	englishKeys := []string{
		"app.title",
		"app.description",
		"status.starting",
		"status.running",
		"status.completed",
		"status.failed",
		"results.pass",
		"results.fail",
		"results.warning",
		"labels.dashboard",
		"labels.tests",
		"labels.settings",
		"labels.help",
		"buttons.start",
		"buttons.stop",
	}

	i18n.SetLanguage(i18n.English)
	for _, key := range englishKeys {
		val := i18n.T(key)
		if val == key {
			t.Errorf("English translation missing for key %q", key)
		}
	}
}

func TestSpanishTranslationsExist(t *testing.T) {
	// Verify that Spanish translations exist for common keys
	spanishKeys := []string{
		"app.title",
		"status.running",
		"status.completed",
		"status.failed",
		"results.pass",
		"results.fail",
		"labels.dashboard",
		"labels.tests",
	}

	for _, key := range spanishKeys {
		val := i18n.TL(key, i18n.Spanish)
		if val == key {
			t.Errorf("Spanish translation missing for key %q", key)
		}
	}
}
