/*
 * The Stem - Internationalization (i18n) Support
 *
 * Provides localization for English (en) and Spanish (es).
 *
 * Copyright (c) 2025 Mustard Seed Networks. All rights reserved.
 */

package i18n

import (
	"os"
	"strings"
	"sync"
)

// Language represents a supported language.
type Language string

const (
	English Language = "en"
	Spanish Language = "es"
)

// DefaultLanguage is the fallback language.
const DefaultLanguage = English

//nolint:gochecknoglobals // Required for message catalog.
var messages = map[Language]map[string]string{
	English: englishMessages,
	Spanish: spanishMessages,
}

//nolint:gochecknoglobals // Required for language state.
var (
	currentLang Language
	langMu      sync.RWMutex
)

//nolint:gochecknoinits // Required for language detection at startup.
func init() {
	currentLang = detectLanguage()
}

// detectLanguage attempts to detect the user's language from environment.
func detectLanguage() Language {
	// Check STEM_LANG first (explicit override)
	if lang := os.Getenv("STEM_LANG"); lang != "" {
		if l := parseLanguage(lang); l != "" {
			return l
		}
	}

	// Check LANGUAGE, LANG, LC_ALL, LC_MESSAGES
	for _, env := range []string{"LANGUAGE", "LANG", "LC_ALL", "LC_MESSAGES"} {
		if lang := os.Getenv(env); lang != "" {
			if l := parseLanguage(lang); l != "" {
				return l
			}
		}
	}

	return DefaultLanguage
}

// parseLanguage extracts language code from locale string (e.g., "es_ES.UTF-8" -> "es").
func parseLanguage(locale string) Language {
	// Handle empty string
	if locale == "" || locale == "C" || locale == "POSIX" {
		return ""
	}

	// Extract language part (before _ or .)
	locale = strings.ToLower(locale)
	if idx := strings.Index(locale, "_"); idx > 0 {
		locale = locale[:idx]
	}
	if idx := strings.Index(locale, "."); idx > 0 {
		locale = locale[:idx]
	}

	switch locale {
	case "en", "english":
		return English
	case "es", "spanish", "espanol":
		return Spanish
	default:
		return ""
	}
}

// SetLanguage sets the current language.
func SetLanguage(lang Language) {
	langMu.Lock()
	defer langMu.Unlock()
	if _, ok := messages[lang]; ok {
		currentLang = lang
	}
}

// GetLanguage returns the current language.
func GetLanguage() Language {
	langMu.RLock()
	defer langMu.RUnlock()
	return currentLang
}

// SupportedLanguages returns all supported languages.
func SupportedLanguages() []Language {
	return []Language{English, Spanish}
}

// T translates a message key to the current language.
// If the key is not found, returns the key itself.
func T(key string) string {
	langMu.RLock()
	lang := currentLang
	langMu.RUnlock()

	if msgs, found := messages[lang]; found {
		if msg, exists := msgs[key]; exists {
			return msg
		}
	}

	// Fallback to English
	if msgs, found := messages[English]; found {
		if msg, exists := msgs[key]; exists {
			return msg
		}
	}

	// Return key as last resort
	return key
}

// TL translates a message key to a specific language.
func TL(key string, lang Language) string {
	if msgs, found := messages[lang]; found {
		if msg, exists := msgs[key]; exists {
			return msg
		}
	}

	// Fallback to English
	if msgs, found := messages[English]; found {
		if msg, exists := msgs[key]; exists {
			return msg
		}
	}

	return key
}

// LanguageName returns the display name for a language.
func LanguageName(lang Language) string {
	switch lang {
	case English:
		return "English"
	case Spanish:
		return "Espanol"
	default:
		return string(lang)
	}
}

// LanguageNativeName returns the native name for a language.
func LanguageNativeName(lang Language) string {
	switch lang {
	case English:
		return "English"
	case Spanish:
		return "Espanol"
	default:
		return string(lang)
	}
}
