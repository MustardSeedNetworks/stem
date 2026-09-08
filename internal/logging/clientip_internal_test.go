// SPDX-License-Identifier: BUSL-1.1

package logging

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
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
			// The leftmost entry is whatever the client sent; with no
			// trusted proxies configured, 10.0.0.2 is the only address a
			// trusted hop actually observed.
			name:       "loopback peer: the last untrusted entry of a chain wins",
			remoteAddr: "127.0.0.1:8443",
			xff:        "198.51.100.7, 10.0.0.1, 10.0.0.2",
			want:       "10.0.0.2",
		},
		{
			// The #807 spoof, reachable through a proxy that appends: the
			// client pre-populates the header and the proxy adds the address
			// it saw. Reading leftmost would hand the attacker the bucket key.
			name:       "loopback peer: a client-supplied prefix cannot claim the bucket",
			remoteAddr: "127.0.0.1:8443",
			xff:        "1.2.3.4, 198.51.100.7",
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
			if got := SecurityClientIP(r, nil); got != tt.want {
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
	auditor := NewAuditor(nil)
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
	auditor := NewAuditor(nil)
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

// mustPrefixes parses a trusted-proxy list that the test author knows is valid.
func mustPrefixes(t *testing.T, list string) []netip.Prefix {
	t.Helper()
	prefixes, err := ParseTrustedProxies(list)
	if err != nil {
		t.Fatalf("ParseTrustedProxies(%q) = %v", list, err)
	}
	return prefixes
}

// #962: behind a non-loopback reverse proxy every client arrived as the
// proxy, so the auth limiter and the failed-login tracker bucketed everyone
// together. The trust list is what lets the header be believed — and only
// from the peers the operator named.
func TestSecurityClientIPTrustedProxies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		trusted    string
		remoteAddr string
		xff        string
		xri        string
		want       string
	}{
		{
			name:       "peer inside a trusted CIDR: the header is honoured",
			trusted:    "10.0.0.0/24",
			remoteAddr: "10.0.0.5:44321",
			xff:        "198.51.100.7",
			want:       "198.51.100.7",
		},
		{
			name:       "peer outside every trusted CIDR: the header is ignored",
			trusted:    "10.0.0.0/24",
			remoteAddr: "10.0.1.5:44321",
			xff:        "198.51.100.7",
			want:       "10.0.1.5",
		},
		{
			name:       "empty list behaves exactly as before: loopback only",
			trusted:    "",
			remoteAddr: "10.0.0.5:44321",
			xff:        "198.51.100.7",
			want:       "10.0.0.5",
		},
		{
			name:       "a chain of trusted hops unwinds to the client",
			trusted:    "10.0.0.0/24",
			remoteAddr: "10.0.0.5:44321",
			xff:        "198.51.100.7, 10.0.0.9, 10.0.0.7",
			want:       "198.51.100.7",
		},
		{
			name:       "an untrusted hop in the chain is as far back as it goes",
			trusted:    "10.0.0.0/24",
			remoteAddr: "10.0.0.5:44321",
			xff:        "198.51.100.7, 203.0.113.9, 10.0.0.7",
			want:       "203.0.113.9",
		},
		{
			name:       "a malformed entry stops the walk instead of being read past",
			trusted:    "10.0.0.0/24",
			remoteAddr: "10.0.0.5:44321",
			xff:        "198.51.100.7, not-an-address, 10.0.0.7",
			want:       "not-an-address",
		},
		{
			name:       "a dual-stack listener reports the peer as IPv4-mapped",
			trusted:    "10.0.0.0/24",
			remoteAddr: "[::ffff:10.0.0.5]:44321",
			xff:        "198.51.100.7",
			want:       "198.51.100.7",
		},
		{
			name:       "IPv6 proxy inside a trusted CIDR",
			trusted:    "2001:db8::/32",
			remoteAddr: "[2001:db8::5]:44321",
			xff:        "198.51.100.7",
			want:       "198.51.100.7",
		},
		{
			name:       "trusted peer, no X-Forwarded-For: X-Real-IP is believed",
			trusted:    "10.0.0.0/24",
			remoteAddr: "10.0.0.5:44321",
			xri:        "198.51.100.7",
			want:       "198.51.100.7",
		},
		{
			name:       "untrusted peer, no X-Forwarded-For: X-Real-IP is not",
			trusted:    "10.0.0.0/24",
			remoteAddr: "203.0.113.9:44321",
			xri:        "198.51.100.7",
			want:       "203.0.113.9",
		},
		{
			name:       "trusted peer, every entry a trusted hop: falls back to the peer",
			trusted:    "10.0.0.0/24",
			remoteAddr: "10.0.0.5:44321",
			xff:        "10.0.0.9, 10.0.0.7",
			want:       "10.0.0.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := loginRequest(tt.remoteAddr, tt.xff)
			if tt.xri != "" {
				r.Header.Set("X-Real-IP", tt.xri)
			}
			if got := SecurityClientIP(r, mustPrefixes(t, tt.trusted)); got != tt.want {
				t.Errorf("SecurityClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseTrustedProxiesRefusesWildcard(t *testing.T) {
	t.Parallel()

	for _, wildcard := range []string{"0.0.0.0/0", "::/0", "10.0.0.0/24,0.0.0.0/0"} {
		if _, err := ParseTrustedProxies(wildcard); err == nil {
			t.Errorf("ParseTrustedProxies(%q) accepted a list that trusts every peer", wildcard)
		}
	}
}

func TestParseTrustedProxiesRefusesMalformed(t *testing.T) {
	t.Parallel()

	for _, bad := range []string{"10.0.0.0", "10.0.0.0/33", "not-a-cidr", "10.0.0.0/24,garbage"} {
		if _, err := ParseTrustedProxies(bad); err == nil {
			t.Errorf("ParseTrustedProxies(%q) accepted a malformed list", bad)
		}
	}
}

func TestParseTrustedProxiesEmptyIsLoopbackOnly(t *testing.T) {
	t.Parallel()

	for _, empty := range []string{"", "   ", ",  ,"} {
		got, err := ParseTrustedProxies(empty)
		if err != nil {
			t.Errorf("ParseTrustedProxies(%q) = %v, want no error", empty, err)
		}
		if len(got) != 0 {
			t.Errorf("ParseTrustedProxies(%q) = %v, want empty", empty, got)
		}
	}
}

func TestParseTrustedProxiesMasksHostBits(t *testing.T) {
	t.Parallel()

	got, err := ParseTrustedProxies(" 10.0.0.5/24 , 192.168.7.5/32 ")
	if err != nil {
		t.Fatalf("ParseTrustedProxies() = %v", err)
	}
	want := []string{"10.0.0.0/24", "192.168.7.5/32"}
	if len(got) != len(want) {
		t.Fatalf("parsed %d prefixes, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].String() != want[i] {
			t.Errorf("prefix %d = %q, want %q", i, got[i].String(), want[i])
		}
	}
}

// The tracker is the second consumer of the trust list (#962): behind a
// configured proxy, two clients must not share one failed-login counter.
func TestLoginFailureSeparatesClientsBehindATrustedProxy(t *testing.T) {
	auditor := NewAuditor(mustPrefixes(t, "10.0.0.0/24"))
	t.Cleanup(auditor.Stop)

	ctx := context.Background()
	const proxy = "10.0.0.5:44321"

	for range 3 {
		auditor.LoginFailure(ctx, loginRequest(proxy, "198.51.100.7"), "admin", "bad password")
	}
	auditor.LoginFailure(ctx, loginRequest(proxy, "203.0.113.9"), "admin", "bad password")

	if got := auditor.Tracker().GetAttemptCount("198.51.100.7"); got != 3 {
		t.Errorf("first client's attempts = %d, want 3", got)
	}
	if got := auditor.Tracker().GetAttemptCount("203.0.113.9"); got != 1 {
		t.Errorf("second client's attempts = %d, want 1", got)
	}
	if got := auditor.Tracker().GetAttemptCount("10.0.0.5"); got != 0 {
		t.Errorf("the proxy itself was counted %d times, want 0", got)
	}
}

// The same proxy without the trust list keeps the pre-#962 behaviour: one
// shared bucket. The setting is additive and cannot loosen a deployment that
// does not use it.
func TestLoginFailureSharesOneBucketWithoutTheTrustList(t *testing.T) {
	auditor := NewAuditor(nil)
	t.Cleanup(auditor.Stop)

	ctx := context.Background()
	const proxy = "10.0.0.5:44321"

	auditor.LoginFailure(ctx, loginRequest(proxy, "198.51.100.7"), "admin", "bad password")
	auditor.LoginFailure(ctx, loginRequest(proxy, "203.0.113.9"), "admin", "bad password")

	if got := auditor.Tracker().GetAttemptCount("10.0.0.5"); got != 2 {
		t.Errorf("attempts against the proxy = %d, want 2", got)
	}
	if got := auditor.Tracker().GetAttemptCount("198.51.100.7"); got != 0 {
		t.Errorf("a forwarded address got its own counter (%d attempts)", got)
	}
}
