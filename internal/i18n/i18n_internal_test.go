/*
 * The Stem - i18n Internal Tests
 *
 * Tests for unexported functions using package-internal access.
 *
 * Copyright (c) 2025 Mustard Seed Networks. All rights reserved.
 */

package i18n

import (
	"os"
	"testing"
)

func TestParseLanguage(t *testing.T) {
	tests := []struct {
		name     string
		locale   string
		expected Language
	}{
		// Empty/special values
		{"empty string", "", ""},
		{"C locale", "C", ""},
		{"POSIX locale", "POSIX", ""},

		// English variants
		{"en", "en", English},
		{"EN uppercase", "EN", English},
		{"english", "english", English},
		{"ENGLISH uppercase", "ENGLISH", English},
		{"en_US", "en_US", English},
		{"en_US.UTF-8", "en_US.UTF-8", English},
		{"en_GB.UTF-8", "en_GB.UTF-8", English},
		{"en.UTF-8", "en.UTF-8", English},

		// Spanish variants
		{"es", "es", Spanish},
		{"ES uppercase", "ES", Spanish},
		{"spanish", "spanish", Spanish},
		{"SPANISH uppercase", "SPANISH", Spanish},
		{"espanol", "espanol", Spanish},
		{"ESPANOL uppercase", "ESPANOL", Spanish},
		{"es_ES", "es_ES", Spanish},
		{"es_ES.UTF-8", "es_ES.UTF-8", Spanish},
		{"es_MX.UTF-8", "es_MX.UTF-8", Spanish},
		{"es.UTF-8", "es.UTF-8", Spanish},

		// Unsupported languages
		{"fr", "fr", ""},
		{"de_DE.UTF-8", "de_DE.UTF-8", ""},
		{"ja_JP.UTF-8", "ja_JP.UTF-8", ""},
		{"zh_CN", "zh_CN", ""},
		{"unknown", "unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLanguage(tt.locale)
			if got != tt.expected {
				t.Errorf("parseLanguage(%q) = %q, want %q", tt.locale, got, tt.expected)
			}
		})
	}
}

func TestDetectLanguage(t *testing.T) {
	// Store original env vars
	origStemLang := os.Getenv("STEM_LANG")
	origLang := os.Getenv("LANG")
	origLanguage := os.Getenv("LANGUAGE")
	origLcAll := os.Getenv("LC_ALL")
	origLcMessages := os.Getenv("LC_MESSAGES")

	// Cleanup function
	defer func() {
		restoreEnvVar("STEM_LANG", origStemLang)
		restoreEnvVar("LANG", origLang)
		restoreEnvVar("LANGUAGE", origLanguage)
		restoreEnvVar("LC_ALL", origLcAll)
		restoreEnvVar("LC_MESSAGES", origLcMessages)
	}()

	tests := []struct {
		name     string
		envVars  map[string]string
		expected Language
	}{
		{
			name:     "STEM_LANG takes precedence over LANG",
			envVars:  map[string]string{"STEM_LANG": "es", "LANG": "en_US.UTF-8"},
			expected: Spanish,
		},
		{
			name:     "STEM_LANG with English",
			envVars:  map[string]string{"STEM_LANG": "en"},
			expected: English,
		},
		{
			name:     "LANGUAGE env var",
			envVars:  map[string]string{"LANGUAGE": "es_ES.UTF-8"},
			expected: Spanish,
		},
		{
			name:     "LANG env var",
			envVars:  map[string]string{"LANG": "es_MX.UTF-8"},
			expected: Spanish,
		},
		{
			name:     "LC_ALL env var",
			envVars:  map[string]string{"LC_ALL": "es_AR.UTF-8"},
			expected: Spanish,
		},
		{
			name:     "LC_MESSAGES env var",
			envVars:  map[string]string{"LC_MESSAGES": "es"},
			expected: Spanish,
		},
		{
			name:     "English from LANG",
			envVars:  map[string]string{"LANG": "en_US.UTF-8"},
			expected: English,
		},
		{
			name:     "Invalid STEM_LANG falls through to LANG",
			envVars:  map[string]string{"STEM_LANG": "fr", "LANG": "es_ES.UTF-8"},
			expected: Spanish,
		},
		{
			name:     "All invalid falls to default",
			envVars:  map[string]string{"STEM_LANG": "fr", "LANG": "de_DE.UTF-8"},
			expected: DefaultLanguage,
		},
		{
			name:     "No env vars defaults to English",
			envVars:  map[string]string{},
			expected: DefaultLanguage,
		},
		{
			name:     "C locale falls through",
			envVars:  map[string]string{"LANG": "C", "LC_ALL": "es"},
			expected: Spanish,
		},
		{
			name:     "POSIX locale falls through",
			envVars:  map[string]string{"LANG": "POSIX", "LC_MESSAGES": "en"},
			expected: English,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars
			clearAllEnvVars()

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v) //nolint:usetesting // Need manual control for detectLanguage
			}

			got := detectLanguage()
			if got != tt.expected {
				t.Errorf("detectLanguage() = %q, want %q", got, tt.expected)
			}
		})
	}
}

//nolint:usetesting // Manual env var control needed for detectLanguage priority testing
func TestDetectLanguagePriority(t *testing.T) {
	// Store original env vars
	origStemLang := os.Getenv("STEM_LANG")
	origLang := os.Getenv("LANG")
	origLanguage := os.Getenv("LANGUAGE")
	origLcAll := os.Getenv("LC_ALL")
	origLcMessages := os.Getenv("LC_MESSAGES")

	// Cleanup function
	defer func() {
		restoreEnvVar("STEM_LANG", origStemLang)
		restoreEnvVar("LANG", origLang)
		restoreEnvVar("LANGUAGE", origLanguage)
		restoreEnvVar("LC_ALL", origLcAll)
		restoreEnvVar("LC_MESSAGES", origLcMessages)
	}()

	// Clear all env vars
	clearAllEnvVars()

	// Test priority: STEM_LANG > LANGUAGE > LANG > LC_ALL > LC_MESSAGES

	// Set all to English except STEM_LANG to Spanish
	os.Setenv("STEM_LANG", "es")
	os.Setenv("LANGUAGE", "en")
	os.Setenv("LANG", "en_US.UTF-8")
	os.Setenv("LC_ALL", "en")
	os.Setenv("LC_MESSAGES", "en")

	got := detectLanguage()
	if got != Spanish {
		t.Errorf("STEM_LANG should take precedence, got %q, want %q", got, Spanish)
	}

	// Clear STEM_LANG, now LANGUAGE should take precedence
	os.Unsetenv("STEM_LANG")
	os.Setenv("LANGUAGE", "es")

	got = detectLanguage()
	if got != Spanish {
		t.Errorf("LANGUAGE should take precedence after STEM_LANG, got %q, want %q", got, Spanish)
	}

	// Clear LANGUAGE, now LANG should take precedence
	os.Unsetenv("LANGUAGE")
	os.Setenv("LANG", "es_ES.UTF-8")

	got = detectLanguage()
	if got != Spanish {
		t.Errorf("LANG should take precedence after LANGUAGE, got %q, want %q", got, Spanish)
	}

	// Clear LANG, now LC_ALL should take precedence
	os.Unsetenv("LANG")
	os.Setenv("LC_ALL", "es")

	got = detectLanguage()
	if got != Spanish {
		t.Errorf("LC_ALL should take precedence after LANG, got %q, want %q", got, Spanish)
	}

	// Clear LC_ALL, now LC_MESSAGES should be used
	os.Unsetenv("LC_ALL")
	os.Setenv("LC_MESSAGES", "es")

	got = detectLanguage()
	if got != Spanish {
		t.Errorf("LC_MESSAGES should be used after LC_ALL, got %q, want %q", got, Spanish)
	}
}

func clearAllEnvVars() {
	os.Unsetenv("STEM_LANG")
	os.Unsetenv("LANG")
	os.Unsetenv("LANGUAGE")
	os.Unsetenv("LC_ALL")
	os.Unsetenv("LC_MESSAGES")
}

func restoreEnvVar(key, value string) {
	if value == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, value)
	}
}

func TestParseLanguageEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		locale   string
		expected Language
	}{
		// Edge cases with dots before underscores
		{"en with dot only", "en.UTF-8", English},
		{"es with dot only", "es.UTF-8", Spanish},

		// Mixed case
		{"En mixed case", "En", English},
		{"eS mixed case", "eS", Spanish},
		{"EnGlIsH mixed case", "EnGlIsH", English},
		{"EsPaNoL mixed case", "EsPaNoL", Spanish},

		// With various region codes
		{"en_AU.UTF-8", "en_AU.UTF-8", English},
		{"en_CA.UTF-8", "en_CA.UTF-8", English},
		{"en_NZ.UTF-8", "en_NZ.UTF-8", English},
		{"es_CL.UTF-8", "es_CL.UTF-8", Spanish},
		{"es_CO.UTF-8", "es_CO.UTF-8", Spanish},
		{"es_PE.UTF-8", "es_PE.UTF-8", Spanish},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLanguage(tt.locale)
			if got != tt.expected {
				t.Errorf("parseLanguage(%q) = %q, want %q", tt.locale, got, tt.expected)
			}
		})
	}
}

func TestTFallbackToEnglish(t *testing.T) {
	// We need to test the case where:
	// 1. Current language is Spanish
	// 2. Key doesn't exist in Spanish
	// 3. Key exists in English
	// This tests the fallback branch in T()

	// Temporarily add a key only to English messages
	testKey := "test.only.in.english.for.coverage"
	testValue := "This key only exists in English"
	englishMessages[testKey] = testValue
	defer delete(englishMessages, testKey)

	// Set language to Spanish
	SetLanguage(Spanish)
	defer SetLanguage(English)

	// T should fall back to English
	got := T(testKey)
	if got != testValue {
		t.Errorf("T(%q) with Spanish set = %q, want %q (should fallback to English)", testKey, got, testValue)
	}
}

func TestTWithCurrentLangMissingKey(t *testing.T) {
	// Test T when the key exists in current language
	SetLanguage(English)

	// Use an existing key
	got := T("app.name")
	if got != "The Stem" {
		t.Errorf("T(\"app.name\") = %q, want \"The Stem\"", got)
	}

	// Switch to Spanish and test
	SetLanguage(Spanish)
	defer SetLanguage(English)

	got = T("app.name")
	if got != "The Stem" {
		t.Errorf("T(\"app.name\") with Spanish = %q, want \"The Stem\"", got)
	}
}

func TestTKeyNotFoundAnywhere(t *testing.T) {
	SetLanguage(Spanish)
	defer SetLanguage(English)

	// A key that doesn't exist anywhere
	key := "this.key.does.not.exist.anywhere"
	got := T(key)
	if got != key {
		t.Errorf("T(%q) = %q, want %q (should return key itself)", key, got, key)
	}
}
