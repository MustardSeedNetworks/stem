// SPDX-License-Identifier: BUSL-1.1

package support_test

import (
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/support"
)

func TestScrubText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		text    string
		secrets []string
		leaked  string
		kept    string
	}{
		{
			"known value anywhere",
			"connect failed for hunter2-value in the middle",
			[]string{"hunter2-value"},
			"hunter2-value",
			"connect failed",
		},
		{"short known value is left to the patterns", "a b c", []string{"b"}, "", "a b c"},
		{"bearer header", "Authorization: Bearer abc.DEF-123_xyz", nil, "abc.DEF", "Authorization"},
		{"JSON password", `{"level":"info","password":"s3cr3t-pw","user":"admin"}`, nil, "s3cr3t-pw", `"user":"admin"`},
		{"JSON token", `{"api_token": "tok-12345"}`, nil, "tok-12345", ""},
		{"key=value", "login token=tok-98765 ok", nil, "tok-98765", "login"},
		{"snake_case prefix", "mail smtp_password=sm-pw-1 sent", nil, "sm-pw-1", "mail"},
		{"digit in the prefix", "db2_password=db-pw-2 up", nil, "db-pw-2", "up"},
		{"kebab header", "x-api-key: kebab-key-3 seen", nil, "kebab-key-3", "seen"},
		{"upper-case kebab", "SMTP-PASSWORD: up-pw-4 seen", nil, "up-pw-4", "seen"},
		{
			"webhook url assignment",
			"alert webhook_url=https://hooks.example/XYZ-SECRET sent",
			nil,
			"XYZ-SECRET",
			"alert",
		},
		{"licence key", "activating MSN1.eyJwYXlsb2FkIjoxfQ.c2lnbmF0dXJl now", nil, "eyJwYXlsb2FkIjoxfQ", "activating"},
		{
			"argon2 hash",
			"stored $argon2id$v=19$m=65536,t=3,p=4$c29tZXNhbHQ$aGFzaGhhc2g done",
			nil,
			"aGFzaGhhc2g",
			"done",
		},
		{
			"bcrypt hash",
			"stored $2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ01234 done",
			nil,
			"abcdefghijklmnopqrstuu",
			"done",
		},
		{
			"private key block",
			"key:\n-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkq\nhkiG9w0BAQEFAASC\n-----END PRIVATE KEY-----\nafter",
			nil,
			"MIIEvQIBADANBgkq",
			"after",
		},
		{
			"EC private key block",
			"-----BEGIN EC PRIVATE KEY-----\nQUJDREVGRw==\n-----END EC PRIVATE KEY-----",
			nil,
			"QUJDREVGRw",
			"",
		},
		{"URL userinfo", "fetching https://deploy:pa55w0rd@mirror.example/stem", nil, "pa55w0rd", "mirror.example"},
		{"plain text is untouched", "listening on https://0.0.0.0:8444", nil, "", "listening on https://0.0.0.0:8444"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := support.ScrubText(tc.text, tc.secrets)
			if tc.leaked != "" && strings.Contains(got, tc.leaked) {
				t.Errorf("ScrubText leaked %q: %q", tc.leaked, got)
			}
			if !strings.Contains(got, tc.kept) {
				t.Errorf("ScrubText removed %q: %q", tc.kept, got)
			}
		})
	}
}

func TestRedactJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		everyString bool
		leaked      []string
		kept        []string
		removed     []string
	}{
		{
			name:        "credential file loses every string",
			input:       `{"username":"admin","passwordHash":"$argon2id$x","extra":"later-field"}`,
			everyString: true,
			leaked:      []string{"admin", "$argon2id$x", "later-field"},
			kept:        []string{`"username"`, `"passwordHash"`, `"extra"`},
			removed:     []string{"$argon2id$x", "admin", "later-field"},
		},
		{
			name:    "descriptor keeps the address, loses the token",
			input:   `{"url":"https://127.0.0.1:8444","token":"bearer-value","caFile":"/var/lib/stem/certs/server.crt"}`,
			leaked:  []string{"bearer-value"},
			kept:    []string{"https://127.0.0.1:8444", "/var/lib/stem/certs/server.crt"},
			removed: []string{"bearer-value"},
		},
		{
			name:    "nested and listed values",
			input:   `{"sources":[{"name":"a","apiKey":"k-1"},{"name":"b","apiKey":"k-2"}],"automatic":true}`,
			leaked:  []string{"k-1", "k-2"},
			kept:    []string{`"automatic": true`, `"name": "a"`},
			removed: []string{"k-1", "k-2"},
		},
		{
			name:    "URL userinfo in a plain field",
			input:   `{"source":"https://u:p4ss@mirror.example/stem"}`,
			leaked:  []string{"p4ss", "u:"},
			kept:    []string{"https://mirror.example/stem"},
			removed: []string{"p4ss", "u"},
		},
		{
			name:    "webhook and receiver URLs are credentials in themselves",
			input:   `{"alertWebhook":"https://hooks.example/T000/B000/XXXXSECRET","receiverUrl":"https://r.example/x?k=SECRET2","note":"kept"}`,
			leaked:  []string{"XXXXSECRET", "SECRET2"},
			kept:    []string{`"note": "kept"`},
			removed: []string{"https://hooks.example/T000/B000/XXXXSECRET", "https://r.example/x?k=SECRET2"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			out, removed, err := support.RedactJSON([]byte(tc.input), tc.everyString)
			if err != nil {
				t.Fatalf("RedactJSON: %v", err)
			}
			assertAbsent(t, string(out), tc.leaked)
			assertPresent(t, string(out), tc.kept)
			if strings.Join(removed, "|") != strings.Join(tc.removed, "|") {
				t.Errorf("removed = %v, want %v", removed, tc.removed)
			}
		})
	}
}

func TestRedactJSONRejectsAFileThatIsNotJSON(t *testing.T) {
	t.Parallel()

	if _, _, err := support.RedactJSON([]byte("password=hunter2"), true); err == nil {
		t.Fatal("RedactJSON accepted a file that is not JSON")
	}
}

func assertAbsent(t *testing.T, text string, items []string) {
	t.Helper()
	for _, item := range items {
		if strings.Contains(text, item) {
			t.Errorf("output leaked %q: %s", item, text)
		}
	}
}

func assertPresent(t *testing.T, text string, items []string) {
	t.Helper()
	for _, item := range items {
		if !strings.Contains(text, item) {
			t.Errorf("output lost %q: %s", item, text)
		}
	}
}
