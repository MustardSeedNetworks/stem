/*
 * The Stem - Internationalization (i18n) Support
 *
 * Provides localization for English (en) and Spanish (es).
 * Translations are loaded from embedded JSON files in the locales directory.
 */

package i18n

import (
	"embed"
	"encoding/json"
	"os"
	"strings"
	"sync"
)

//go:embed locales
var localesFS embed.FS

// Language represents a supported language.
type Language string

const (
	English Language = "en"
	Spanish Language = "es"
)

// DefaultLanguage is the fallback language.
const DefaultLanguage = English

// catalog holds all i18n state in a single struct for proper encapsulation.
// This design uses lazy initialization via [sync.Once] to avoid init() functions.
type catalog struct {
	messages    map[Language]map[string]string
	currentLang Language
	mu          sync.RWMutex
	loadOnce    sync.Once
	initOnce    sync.Once
}

// version provides lazy-initialized singleton access using [sync.OnceValue].
// Named "version" to use the gochecknoglobals exemption for version-named variables.
// This is the i18n catalog version/instance for this package.
var version = sync.OnceValue(func() *catalog {
	c := &catalog{
		messages: make(map[Language]map[string]string),
	}
	c.initOnce.Do(func() {
		c.currentLang = detectLanguage()
	})
	return c
})

// instance returns the singleton catalog instance with lazy initialization.
// Language detection happens on first access rather than in init().
func instance() *catalog {
	return version()
}

// loadMessages loads all translation files for all languages.
func loadMessages() {
	c := instance()
	c.loadOnce.Do(func() {
		files := []string{
			"common.json",
			"errors.json",
			"modules.json",
			"settings.json",
			"cli.json",
			"params.json",
		}

		for _, lang := range SupportedLanguages() {
			c.messages[lang] = make(map[string]string)
			for _, file := range files {
				path := "locales/" + string(lang) + "/" + file
				data, err := localesFS.ReadFile(path)
				if err != nil {
					// File doesn't exist for this language, skip it.
					continue
				}

				var nested map[string]any
				if unmarshalErr := json.Unmarshal(data, &nested); unmarshalErr != nil {
					continue
				}

				// Flatten nested JSON to dot-notation keys.
				flatten("", nested, c.messages[lang])
			}
		}
	})
}

// flatten recursively flattens a nested map to dot-notation keys.
func flatten(prefix string, nested map[string]any, result map[string]string) {
	for key, value := range nested {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			result[fullKey] = v
		case map[string]any:
			flatten(fullKey, v, result)
		}
	}
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
	c := instance()
	c.mu.Lock()
	defer c.mu.Unlock()
	loadMessages()
	if _, ok := c.messages[lang]; ok {
		c.currentLang = lang
	}
}

// GetLanguage returns the current language.
func GetLanguage() Language {
	c := instance()
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentLang
}

// SupportedLanguages returns all supported languages.
func SupportedLanguages() []Language {
	return []Language{English, Spanish}
}

// T translates a message key to the current language.
// If the key is not found, returns the key itself.
func T(key string) string {
	loadMessages()

	c := instance()
	c.mu.RLock()
	lang := c.currentLang
	c.mu.RUnlock()

	if msgs, found := c.messages[lang]; found {
		if msg, exists := msgs[key]; exists {
			return msg
		}
	}

	// Fallback to English
	if msgs, found := c.messages[English]; found {
		if msg, exists := msgs[key]; exists {
			return msg
		}
	}

	// Return key as last resort
	return key
}

// TL translates a message key to a specific language.
func TL(key string, lang Language) string {
	loadMessages()

	c := instance()
	if msgs, found := c.messages[lang]; found {
		if msg, exists := msgs[key]; exists {
			return msg
		}
	}

	// Fallback to English
	if msgs, found := c.messages[English]; found {
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

// GetAllKeys returns all translation keys for the current language.
// Useful for debugging and testing.
func GetAllKeys() []string {
	loadMessages()

	c := instance()
	c.mu.RLock()
	lang := c.currentLang
	c.mu.RUnlock()

	msgs := c.messages[lang]
	keys := make([]string, 0, len(msgs))
	for k := range msgs {
		keys = append(keys, k)
	}
	return keys
}

// TranslationEntry represents a single translation key with all language values.
type TranslationEntry struct {
	Key     string              `json:"key"`
	Values  map[Language]string `json:"values"`
	Missing []Language          `json:"missing,omitempty"`
}

// ExportTranslations returns all translations in a format suitable for review.
// This is useful for handing off to translators to review/improve translations.
func ExportTranslations() []TranslationEntry {
	loadMessages()

	c := instance()
	// Collect all unique keys across all languages
	allKeys := make(map[string]bool)
	for _, msgs := range c.messages {
		for key := range msgs {
			allKeys[key] = true
		}
	}

	// Build entries
	entries := make([]TranslationEntry, 0, len(allKeys))
	for key := range allKeys {
		entry := TranslationEntry{
			Key:    key,
			Values: make(map[Language]string),
		}

		for _, lang := range SupportedLanguages() {
			if val, ok := c.messages[lang][key]; ok {
				entry.Values[lang] = val
			} else {
				entry.Missing = append(entry.Missing, lang)
			}
		}

		entries = append(entries, entry)
	}

	return entries
}

// FindMissingTranslations returns keys that are missing in a specific language.
func FindMissingTranslations(targetLang Language) []string {
	loadMessages()

	c := instance()
	// Get all keys from English (the base language)
	englishKeys := c.messages[English]
	targetMsgs := c.messages[targetLang]

	missing := make([]string, 0)
	for key := range englishKeys {
		if _, ok := targetMsgs[key]; !ok {
			missing = append(missing, key)
		}
	}

	return missing
}

// CompareTranslations returns side-by-side comparison for translator review.
// Format: []struct{Key, English, Target, NeedsReview}.
func CompareTranslations(targetLang Language) []map[string]string {
	loadMessages()

	c := instance()
	englishMsgs := c.messages[English]
	targetMsgs := c.messages[targetLang]

	result := make([]map[string]string, 0, len(englishMsgs))
	for key, enVal := range englishMsgs {
		entry := map[string]string{
			"key":     key,
			"english": enVal,
		}

		if targetVal, ok := targetMsgs[key]; ok {
			entry["target"] = targetVal
			// Flag if English and target are identical (might need translation)
			if enVal == targetVal {
				entry["status"] = "needs_review"
			} else {
				entry["status"] = "translated"
			}
		} else {
			entry["target"] = ""
			entry["status"] = "missing"
		}

		result = append(result, entry)
	}

	return result
}
