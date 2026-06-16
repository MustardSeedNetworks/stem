// SPDX-License-Identifier: BUSL-1.1

// Package cors classifies HTTP Origin header values for the api transport
// layer's CORS policy: localhost detection, same-origin matching, and RFC 1918
// private-network validation. The classification is deliberately strict —
// validating the complete IP structure — to prevent CORS-bypass tricks such as
// "localhost.evil.com" or "192.168.1.1.evil.com".
//
// It is a leaf of internal/api (ADR-0011): it depends only on the standard
// library (net/url, strings) — never on the api transport layer itself. The
// boundary is enforced by depguard (api-cors-isolated). The HTTP middleware
// that consumes these classifiers (corsMiddleware), the opt-in env read
// (corsAllowPrivateEnabled), and the response-header wiring stay in the api
// transport layer.
package cors

import (
	"net/url"
	"strings"
)

// RFC 1918 validation constants.
const (
	// ipPartsClassC is the expected number of IP parts for Class C address validation.
	ipPartsClassC = 2

	// ipPartsClassAB is the expected number of IP parts for Class A/B address validation.
	ipPartsClassAB = 3

	// decimalParseBase is the base for decimal digit parsing.
	decimalParseBase = 10

	// maxIPOctetValue is the maximum valid value for an IP address octet (255).
	maxIPOctetValue = 255

	// classBMinOctet is the minimum second octet for 172.x.x.x private range.
	classBMinOctet = 16

	// classBMaxOctet is the maximum second octet for 172.x.x.x private range.
	classBMaxOctet = 31
)

// IsLocalhostOrigin validates that the origin is actually localhost.
// Prevents CORS bypass via origins like "localhost.evil.com".
func IsLocalhostOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := u.Hostname()
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// IsSameOrigin checks if the Origin header matches the request's Host.
// This allows browsers to access the server from its actual address (e.g., 10.0.0.210:8444).
func IsSameOrigin(origin string, requestHost string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	// Compare origin host:port with request host.
	originHost := u.Host // Includes port if present.
	return originHost == requestHost
}

// IsRFC1918Origin checks if the origin is an RFC 1918 private network address.
// Ported from Seed for CORS validation - allows connections from private networks.
//
// Allowed addresses:
//   - Class A private: 10.0.0.0/8 (10.x.x.x)
//   - Class B private: 172.16.0.0/12 (172.16.x.x through 172.31.x.x)
//   - Class C private: 192.168.0.0/16 (192.168.x.x)
//
// Uses proper IP validation to prevent subdomain bypass attacks.
// Rejects malicious origins like "http://192.168.1.1.evil.com".
func IsRFC1918Origin(origin string) bool {
	// Reject null origin
	if origin == "null" {
		return false
	}

	u, err := url.Parse(origin)
	if err != nil {
		return false
	}

	host := u.Hostname()
	if host == "" {
		return false
	}

	// Check for RFC 1918 private network ranges
	return isPrivateNetworkAddress(host)
}

// isPrivateNetworkAddress checks if the host is an RFC 1918 private network address.
// This prevents subdomain attacks like "192.168.1.1.evil.com" by validating
// the complete IP address structure.
func isPrivateNetworkAddress(host string) bool {
	// Class C: 192.168.0.0/16
	if strings.HasPrefix(host, "192.168.") {
		return isValidClassCAddress(host)
	}

	// Class A: 10.0.0.0/8
	if strings.HasPrefix(host, "10.") {
		return isValidClassAAddress(host)
	}

	// Class B: 172.16.0.0/12 (172.16.0.0 - 172.31.255.255)
	if strings.HasPrefix(host, "172.") {
		return isValidClassBAddress(host)
	}

	return false
}

// isValidClassCAddress validates a 192.168.x.x address.
// Returns true if the host is a valid Class C private address.
func isValidClassCAddress(host string) bool {
	remainder := host[8:] // After "192.168."
	// Should be X.Y where X and Y are 0-255
	parts := strings.Split(remainder, ".")
	if len(parts) != ipPartsClassC {
		return false
	}
	return isValidIPOctet(parts[0]) && isValidIPOctet(parts[1])
}

// isValidClassAAddress validates a 10.x.x.x address.
// Returns true if the host is a valid Class A private address.
func isValidClassAAddress(host string) bool {
	remainder := host[3:] // After "10."
	parts := strings.Split(remainder, ".")
	if len(parts) != ipPartsClassAB {
		return false
	}
	return isValidIPOctet(parts[0]) && isValidIPOctet(parts[1]) && isValidIPOctet(parts[2])
}

// isValidClassBAddress validates a 172.16-31.x.x address.
// Returns true if the host is a valid Class B private address (172.16.0.0/12).
func isValidClassBAddress(host string) bool {
	remainder := host[4:] // After "172."
	parts := strings.Split(remainder, ".")
	if len(parts) != ipPartsClassAB {
		return false
	}

	// Validate and parse second octet to verify range 16-31
	secondOctet, ok := parseOctetInRange(parts[0], classBMinOctet, classBMaxOctet)
	if !ok || secondOctet < classBMinOctet || secondOctet > classBMaxOctet {
		return false
	}

	return isValidIPOctet(parts[1]) && isValidIPOctet(parts[2])
}

// parseOctetInRange parses an octet string and checks if it's within the given range.
// Returns the parsed value and true if valid, 0 and false otherwise.
func parseOctetInRange(s string, minVal, maxVal int) (int, bool) {
	if s == "" || len(s) > 3 {
		return 0, false
	}

	val := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
		val = val*decimalParseBase + int(c-'0')
		if val > maxIPOctetValue {
			return 0, false
		}
	}

	if val < minVal || val > maxVal {
		return val, false
	}

	return val, true
}

// isValidIPOctet checks if a string is a valid IP octet (0-255).
// Helper function for proper IP validation.
func isValidIPOctet(s string) bool {
	if s == "" || len(s) > 3 {
		return false
	}

	val := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
		val = val*decimalParseBase + int(c-'0')
		if val > maxIPOctetValue {
			return false
		}
	}

	return true
}
