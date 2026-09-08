// SPDX-License-Identifier: BUSL-1.1

package logging

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"slices"
	"strings"
)

// TrustedProxiesEnv names the environment variable carrying the trusted-proxy
// CIDR list, comma separated (for example "10.0.0.0/24,192.168.7.5/32").
// Unset — the default — means loopback is the only trusted hop.
const TrustedProxiesEnv = "STEM_TRUSTED_PROXIES"

// ParseTrustedProxies parses a comma-separated CIDR list into prefixes.
//
// A prefix covering every address ("0.0.0.0/0", "::/0", or any other zero-bit
// spelling) is refused: trusting every peer is exactly the spoofable
// behaviour #807 removed, written a different way. An empty or whitespace-only
// list parses to nil, which keeps the loopback-only default.
func ParseTrustedProxies(list string) ([]netip.Prefix, error) {
	var prefixes []netip.Prefix
	for entry := range strings.SplitSeq(list, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(entry)
		if err != nil {
			return nil, fmt.Errorf("trusted proxy %q: %w", entry, err)
		}
		if prefix.Bits() == 0 {
			return nil, fmt.Errorf("trusted proxy %q trusts every peer, which makes "+
				"X-Forwarded-For spoofable by any client", entry)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return prefixes, nil
}

// SecurityClientIP returns the client address that may be used for security
// decisions: rate-limit buckets, failed-login counters, ban lists.
//
// It is the counterpart to [GetClientIP], and the difference is the whole
// point. GetClientIP returns whatever the client claimed, which is right for
// a log field and wrong for anything that decides. This returns the immediate
// TCP peer, and honours X-Forwarded-For / X-Real-IP only when that peer is a
// trusted hop: a loopback address (the single-host reverse-proxy deployment)
// or a member of trusted, the operator's [TrustedProxiesEnv] list as parsed
// by [ParseTrustedProxies]. The list is passed in rather than held in a
// package variable so each consumer that keys on this — the auth rate
// limiter, the failed-login tracker — owns its copy, and a consumer that was
// never given one falls back to loopback-only rather than to nothing.
//
// It fails closed: an unparseable peer address is not loopback and is in no
// CIDR, so forwarding headers are ignored rather than trusted from an unknown
// source, and an unset list leaves loopback as the only trusted hop.
func SecurityClientIP(r *http.Request, trusted []netip.Prefix) string {
	peer := r.RemoteAddr
	if host, _, err := net.SplitHostPort(peer); err == nil {
		// RemoteAddr does not always carry a port (httptest, unix sockets).
		peer = host
	}

	addr, err := netip.ParseAddr(peer)
	if err != nil || !isTrustedHop(addr, trusted) {
		return peer
	}
	if forwarded := forwardedClientIP(r, trusted); forwarded != "" {
		return forwarded
	}
	return peer
}

// forwardedClientIP returns the client IP conveyed by X-Forwarded-For or
// X-Real-IP, or "" if neither carries a usable value. Callers MUST gate its
// use on the immediate peer being a trusted hop.
//
// The header is read right to left, not left to right. Every proxy in the
// chain appends the peer it saw (nginx's $proxy_add_x_forwarded_for, and every
// load balancer that follows RFC 7239's model), so the rightmost entries are
// the ones our own trusted hops observed and the leftmost is whatever the
// original client chose to send — a client that pre-populates the header
// prepends an address of its choosing. Walking from the right and discarding
// entries that are themselves trusted hops stops at the last address the
// trusted chain actually vouched for, which is as far back as it can be
// believed. An entry that will not parse also stops the walk, so a malformed
// header cannot be used to reach past it.
func forwardedClientIP(r *http.Request, trusted []netip.Prefix) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for _, raw := range slices.Backward(strings.Split(xff, ",")) {
			entry := strings.TrimSpace(raw)
			if entry == "" {
				continue
			}
			addr, err := netip.ParseAddr(entry)
			if err != nil || !isTrustedHop(addr, trusted) {
				return entry
			}
		}
	}
	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return xri
	}
	return ""
}

// isTrustedHop reports whether addr is a hop whose forwarding headers may be
// believed: a loopback address, or a member of the configured list.
// An invalid address is not trusted, so the caller fails closed.
func isTrustedHop(addr netip.Addr, trusted []netip.Prefix) bool {
	if !addr.IsValid() {
		return false
	}
	if addr.IsLoopback() {
		return true
	}
	// A dual-stack listener reports an IPv4 peer as ::ffff:a.b.c.d, which no
	// IPv4 prefix contains, and Prefix.Contains rejects any zoned address.
	addr = addr.Unmap().WithZone("")
	for _, prefix := range trusted {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}
