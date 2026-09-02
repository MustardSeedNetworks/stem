// SPDX-License-Identifier: BUSL-1.1

package logging

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// loginRequest builds a login request from a given TCP peer, optionally
// carrying a forwarding header the client chose for itself.
func loginRequest(remoteAddr, xff string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	r.RemoteAddr = remoteAddr
	if xff != "" {
		r.Header.Set("X-Forwarded-For", xff)
	}
	return r
}

func TestSecurityClientIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		xri        string
		want       string
	}{
		{
			name:       "public peer: forwarding header is ignored",
			remoteAddr: "203.0.113.9:44321",
			xff:        "1.2.3.4",
			want:       "203.0.113.9",
		},
		{
			name:       "public peer: X-Real-IP is ignored too",
			remoteAddr: "203.0.113.9:44321",
			xri:        "1.2.3.4",
			want:       "203.0.113.9",
		},
		{
			name:       "loopback peer: a local reverse proxy may convey the client",
			remoteAddr: "127.0.0.1:8443",
			xff:        "198.51.100.7",
			want:       "198.51.100.7",
		},
		{
			name:       "loopback peer: leftmost entry of a chain wins",
			remoteAddr: "127.0.0.1:8443",
			xff:        "198.51.100.7, 10.0.0.1, 10.0.0.2",
			want:       "198.51.100.7",
		},
		{
			name:       "IPv6 loopback is loopback",
			remoteAddr: "[::1]:8443",
			xff:        "198.51.100.7",
			want:       "198.51.100.7",
		},
		{
			name:       "loopback peer, empty header: falls back to the peer",
			remoteAddr: "127.0.0.1:8443",
			xff:        "   ",
			want:       "127.0.0.1",
		},
		{
			name:       "unparseable peer fails closed — header is not trusted",
			remoteAddr: "not-an-address",
			xff:        "1.2.3.4",
			want:       "not-an-address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := loginRequest(tt.remoteAddr, tt.xff)
			if tt.xri != "" {
				r.Header.Set("X-Real-IP", tt.xri)
			}
			if got := SecurityClientIP(r); got != tt.want {
				t.Errorf("SecurityClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

// The defect in #807: the tracker keyed on GetClientIP, so an attacker who
// varied X-Forwarded-For got a fresh counter on every request and the
// threshold could never be reached. This drives the real LoginFailure path,
// not the tracker directly, because the bug was in the key the caller chose.
func TestLoginFailureCountsRotatingForwardedHeadersAsOneClient(t *testing.T) {
	auditor := NewAuditor()
	t.Cleanup(auditor.Stop)

	ctx := context.Background()
	const peer = "203.0.113.9:44321"

	forged := []string{"1.2.3.4", "5.6.7.8", "9.10.11.12", "13.14.15.16", "17.18.19.20"}
	alerted := false
	for _, xff := range forged {
		if auditor.LoginFailure(ctx, loginRequest(peer, xff), "admin", "bad password") {
			alerted = true
		}
	}

	if got := auditor.Tracker().GetAttemptCount("203.0.113.9"); got != len(forged) {
		t.Errorf("attempts recorded against the peer = %d, want %d", got, len(forged))
	}
	if !alerted {
		t.Errorf("five failures from one peer did not raise the suspicious-activity alert")
	}
	// Each forged value must not have become its own bucket.
	for _, xff := range forged {
		if got := auditor.Tracker().GetAttemptCount(xff); got != 0 {
			t.Errorf("forged header %q got its own counter (%d attempts)", xff, got)
		}
	}
}

// The mirror of the above: LoginSuccess clears a counter, so a spoofable key
// there lets an attacker wipe somebody else's failed-attempt record.
func TestLoginSuccessCannotClearAnotherClientsCounter(t *testing.T) {
	auditor := NewAuditor()
	t.Cleanup(auditor.Stop)

	ctx := context.Background()
	const victimPeer = "198.51.100.7:52000"
	const attackerPeer = "203.0.113.9:44321"

	for range 3 {
		auditor.LoginFailure(ctx, loginRequest(victimPeer, ""), "admin", "bad password")
	}
	if got := auditor.Tracker().GetAttemptCount("198.51.100.7"); got != 3 {
		t.Fatalf("setup: victim attempts = %d, want 3", got)
	}

	// The attacker logs in successfully as themselves, claiming to be the
	// victim in X-Forwarded-For.
	auditor.LoginSuccess(ctx, loginRequest(attackerPeer, "198.51.100.7"), "mallory", "mallory")

	if got := auditor.Tracker().GetAttemptCount("198.51.100.7"); got != 3 {
		t.Errorf("victim's counter was cleared by a forged header: %d attempts left, want 3", got)
	}
}
