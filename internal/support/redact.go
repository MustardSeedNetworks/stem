// SPDX-License-Identifier: BUSL-1.1

package support

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// Placeholder replaces every credential a bundle would otherwise carry. It is a
// fixed string rather than a length-preserving mask, so a reader sees at a
// glance that a value was removed and learns nothing about its length.
const Placeholder = logging.RedactedPlaceholder

// minSecretLen is the shortest known secret scrubbed by value. Replacing a
// one- or two-character value would shred unrelated text; the patterns still
// catch it when it is labelled.
const minSecretLen = 4

var (
	// sensitiveKey names a JSON field whose value is a credential.
	sensitiveKey = regexp.MustCompile(
		`(?i)pass(word|wd|phrase)?|hash|token|secret|key|license|credential|salt|jwt|auth|webhook|receiver|hook`,
	)

	// jsonCredential catches `"password":"x"` in a JSON log line, which the
	// logging package's key=value patterns do not.
	jsonCredential = regexp.MustCompile(
		`(?i)("[A-Za-z0-9_.-]*(?:pass(?:word|wd|phrase)?|hash|token|secret|key|license|credential|jwt|webhook|receiver)` +
			`[A-Za-z0-9_.-]*"\s*:\s*)"[^"]*"`,
	)

	// credentialAssignment catches `token=x`, `license_key: x`, `smtp_password=x`
	// and `x-api-key: x`. There is no leading boundary: the prefix is part of
	// the key's name and is kept, and a boundary would stop `db2_password`.
	credentialAssignment = regexp.MustCompile(
		`(?i)([A-Za-z0-9_-]*(?:token|password|passphrase|secret|api[_-]?key|jwt|license[_-]?key|hash|` +
			`webhook[_-]?url|receiver[_-]?url))(\s*[:=]\s*)"?[^"\s,}]+"?`,
	)

	// licenseKey is the shape of an activation key: MSN1.<payload>.<signature>.
	licenseKey = regexp.MustCompile(`MSN1\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`)

	// passwordHash is the PHC string an argon2 or bcrypt hash is stored as.
	passwordHash = regexp.MustCompile(`\$(?:argon2(?:id|i|d)|2[abxy])\$[^\s"']+`)

	// privateKey is a whole PEM private key block, whatever its algorithm.
	privateKey = regexp.MustCompile(`(?s)-----BEGIN [A-Z ]*PRIVATE KEY-----.*?-----END [A-Z ]*PRIVATE KEY-----`)

	// urlUserinfo is the user:password part of a URL.
	urlUserinfo = regexp.MustCompile(`(://)[^/\s:@"]+:[^/\s@"]+@`)
)

// ScrubText removes credentials from free text: the known values first, because
// they are exact and a secret that looks like a word is removed by value rather
// than by pattern, then everything still shaped like one.
func ScrubText(text string, secrets []string) string {
	for _, secret := range secrets {
		if len(secret) >= minSecretLen {
			text = strings.ReplaceAll(text, secret, Placeholder)
		}
	}
	text = privateKey.ReplaceAllString(text, Placeholder)
	text = licenseKey.ReplaceAllString(text, Placeholder)
	text = passwordHash.ReplaceAllString(text, Placeholder)
	text = urlUserinfo.ReplaceAllString(text, "${1}"+Placeholder+"@")
	text = jsonCredential.ReplaceAllString(text, `$1"`+Placeholder+`"`)
	text = credentialAssignment.ReplaceAllString(text, "$1$2"+Placeholder)
	return logging.RedactString(text)
}

// RedactJSON returns data with its credentials replaced by the placeholder,
// the structure intact, together with the credential values it removed so log
// text elsewhere in the bundle can be scrubbed of them too.
//
// With everyString set every string leaf is removed, which is what the
// administrator credential file gets: its shape is known, and a field added
// to it later is hidden by default rather than leaked by default. Otherwise a
// leaf is removed when its key names a credential.
func RedactJSON(data []byte, everyString bool) ([]byte, []string, error) {
	var document any
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, nil, fmt.Errorf("decode: %w", err)
	}
	var removed []string
	document = redactValue(document, "", everyString, &removed)
	out, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("encode: %w", err)
	}
	slices.Sort(removed)
	return append(out, '\n'), slices.Compact(removed), nil
}

func redactValue(value any, key string, everyString bool, removed *[]string) any {
	switch typed := value.(type) {
	case map[string]any:
		for name, child := range typed {
			typed[name] = redactValue(child, name, everyString, removed)
		}
		return typed
	case []any:
		for i, child := range typed {
			typed[i] = redactValue(child, key, everyString, removed)
		}
		return typed
	case string:
		if typed == "" {
			return typed
		}
		if everyString || sensitiveKey.MatchString(key) {
			*removed = append(*removed, typed)
			return Placeholder
		}
		return stripURLCredentials(typed, removed)
	default:
		return value
	}
}

// stripURLCredentials removes the password from a URL-valued string, and the
// whole userinfo with it, and records what it removed so the same text in the
// log is scrubbed by value. A string that is not a URL passes through the
// pattern scrub.
func stripURLCredentials(value string, removed *[]string) string {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || parsed.User == nil {
		return ScrubText(value, nil)
	}
	if password, set := parsed.User.Password(); set {
		*removed = append(*removed, password)
	}
	if name := parsed.User.Username(); name != "" {
		*removed = append(*removed, name)
	}
	parsed.User = nil
	return ScrubText(parsed.String(), nil)
}
