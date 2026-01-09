/*
 * The Stem - i18n Tests
 *
 * Copyright (c) 2025 Mustard Seed Networks. All rights reserved.
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
		{"app.name", "The Stem"},
		{"status.running", "Running"},
		{"result.pass", "PASS"},
		{"ui.dashboard", "Dashboard"},
	}

	for _, tt := range tests {
		got := i18n.T(tt.key)
		if got != tt.expected {
			t.Errorf("T(%q) = %q, want %q", tt.key, got, tt.expected)
		}
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
		{"result.pass", i18n.English, "PASS"},
		{"result.pass", i18n.Spanish, "APROBADO"},
		{"ui.dashboard", i18n.English, "Dashboard"},
		{"ui.dashboard", i18n.Spanish, "Panel de Control"},
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

func TestFallback(t *testing.T) {
	// Unknown key should return the key itself
	key := "unknown.key.that.does.not.exist"
	got := i18n.T(key)
	if got != key {
		t.Errorf("T(%q) = %q, want %q (fallback to key)", key, got, key)
	}
}

func TestSpanishFallback(t *testing.T) {
	i18n.SetLanguage(i18n.Spanish)
	defer i18n.SetLanguage(i18n.English)

	// All English keys should have Spanish translations
	// This is a spot check of important keys
	keysToCheck := []string{
		"app.name",
		"status.running",
		"status.completed",
		"status.failed",
		"result.pass",
		"result.fail",
		"ui.dashboard",
		"ui.tests",
		"ui.settings",
		"err.interface_required",
		"err.license_required",
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
		// (app.name is an exception as it's a proper noun)
		if key != "app.name" && en == es {
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
