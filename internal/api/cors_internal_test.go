// SPDX-License-Identifier: BUSL-1.1

package api

import "testing"

func TestIsLocalhostOrigin(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		// Valid localhost origins - should accept.
		{
			name:   "localhost with port",
			origin: "http://localhost:8080",
			want:   true,
		},
		{
			name:   "localhost without port",
			origin: "http://localhost",
			want:   true,
		},
		{
			name:   "localhost HTTPS",
			origin: "https://localhost:8443",
			want:   true,
		},
		{
			name:   "IPv4 loopback with port",
			origin: "http://127.0.0.1:8080",
			want:   true,
		},
		{
			name:   "IPv4 loopback without port",
			origin: "http://127.0.0.1",
			want:   true,
		},
		{
			name:   "IPv6 loopback with port",
			origin: "http://[::1]:8080",
			want:   true,
		},
		{
			name:   "IPv6 loopback without port",
			origin: "http://[::1]",
			want:   true,
		},

		// CORS bypass attempts - should reject.
		{
			name:   "bypass via subdomain prefix",
			origin: "http://localhost.evil.com",
			want:   false,
		},
		{
			name:   "bypass via subdomain suffix",
			origin: "http://evil.localhost.com",
			want:   false,
		},
		{
			name:   "bypass via prefix without dot",
			origin: "http://notlocalhost:8080",
			want:   false,
		},
		{
			name:   "bypass via malicious subdomain",
			origin: "http://localhost.attacker.com:8080",
			want:   false,
		},
		{
			name:   "external domain",
			origin: "http://example.com",
			want:   false,
		},
		{
			name:   "external domain with port",
			origin: "http://example.com:8080",
			want:   false,
		},

		// Edge cases.
		{
			name:   "empty origin",
			origin: "",
			want:   false,
		},
		{
			name:   "malformed URL",
			origin: "not-a-valid-url",
			want:   false,
		},
		{
			name:   "localhost in path only",
			origin: "http://evil.com/localhost",
			want:   false,
		},
		{
			name:   "localhost in query",
			origin: "http://evil.com?host=localhost",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLocalhostOrigin(tt.origin)
			if got != tt.want {
				t.Errorf("isLocalhostOrigin(%q) = %v, want %v", tt.origin, got, tt.want)
			}
		})
	}
}

// TestIsRFC1918Origin tests the RFC 1918 private network address validation.
func TestIsRFC1918Origin(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		// Valid Class C addresses (192.168.x.x).
		{
			name:   "class C with port",
			origin: "http://192.168.1.1:8080",
			want:   true,
		},
		{
			name:   "class C without port",
			origin: "http://192.168.1.1",
			want:   true,
		},
		{
			name:   "class C HTTPS",
			origin: "https://192.168.0.100:8443",
			want:   true,
		},
		{
			name:   "class C edge 192.168.0.0",
			origin: "http://192.168.0.0",
			want:   true,
		},
		{
			name:   "class C edge 192.168.255.255",
			origin: "http://192.168.255.255",
			want:   true,
		},

		// Valid Class A addresses (10.x.x.x).
		{
			name:   "class A with port",
			origin: "http://10.0.0.1:8080",
			want:   true,
		},
		{
			name:   "class A without port",
			origin: "http://10.0.0.1",
			want:   true,
		},
		{
			name:   "class A edge 10.0.0.0",
			origin: "http://10.0.0.0",
			want:   true,
		},
		{
			name:   "class A edge 10.255.255.255",
			origin: "http://10.255.255.255",
			want:   true,
		},

		// Valid Class B addresses (172.16-31.x.x).
		{
			name:   "class B 172.16.x.x",
			origin: "http://172.16.0.1:8080",
			want:   true,
		},
		{
			name:   "class B 172.31.x.x",
			origin: "http://172.31.255.255",
			want:   true,
		},
		{
			name:   "class B 172.20.x.x",
			origin: "http://172.20.100.50",
			want:   true,
		},

		// Invalid Class B addresses (outside 172.16-31.x.x).
		{
			name:   "class B outside range 172.15.x.x",
			origin: "http://172.15.0.1",
			want:   false,
		},
		{
			name:   "class B outside range 172.32.x.x",
			origin: "http://172.32.0.1",
			want:   false,
		},

		// Subdomain bypass attacks - should reject.
		{
			name:   "bypass 192.168 subdomain",
			origin: "http://192.168.1.1.evil.com",
			want:   false,
		},
		{
			name:   "bypass 10.x subdomain",
			origin: "http://10.0.0.1.evil.com",
			want:   false,
		},
		{
			name:   "bypass 172.16 subdomain",
			origin: "http://172.16.0.1.evil.com",
			want:   false,
		},

		// Public addresses - should reject.
		{
			name:   "public IP 8.8.8.8",
			origin: "http://8.8.8.8",
			want:   false,
		},
		{
			name:   "public domain",
			origin: "http://example.com",
			want:   false,
		},

		// Localhost should NOT be matched by RFC 1918 (handled by isLocalhostOrigin).
		{
			name:   "localhost not RFC 1918",
			origin: "http://localhost",
			want:   false,
		},
		{
			name:   "127.0.0.1 not RFC 1918",
			origin: "http://127.0.0.1",
			want:   false,
		},

		// Edge cases.
		{
			name:   "null origin",
			origin: "null",
			want:   false,
		},
		{
			name:   "empty origin",
			origin: "",
			want:   false,
		},
		{
			name:   "malformed URL",
			origin: "not-a-valid-url",
			want:   false,
		},
		{
			name:   "invalid octet 256",
			origin: "http://192.168.1.256",
			want:   false,
		},
		{
			name:   "invalid octet negative",
			origin: "http://192.168.-1.1",
			want:   false,
		},
		{
			name:   "too few octets",
			origin: "http://192.168.1",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRFC1918Origin(tt.origin)
			if got != tt.want {
				t.Errorf("isRFC1918Origin(%q) = %v, want %v", tt.origin, got, tt.want)
			}
		})
	}
}

// TestIsPrivateNetworkAddress tests the private network address validation helper.
func TestIsPrivateNetworkAddress(t *testing.T) {
	tests := []struct {
		name string
		host string
		want bool
	}{
		// Class C.
		{"class C valid", "192.168.1.1", true},
		{"class C zero", "192.168.0.0", true},
		{"class C max", "192.168.255.255", true},

		// Class A.
		{"class A valid", "10.0.0.1", true},
		{"class A zero", "10.0.0.0", true},
		{"class A max", "10.255.255.255", true},

		// Class B.
		{"class B 172.16", "172.16.0.1", true},
		{"class B 172.31", "172.31.255.255", true},
		{"class B 172.20", "172.20.100.50", true},

		// Invalid.
		{"class B 172.15 invalid", "172.15.0.1", false},
		{"class B 172.32 invalid", "172.32.0.1", false},
		{"public IP", "8.8.8.8", false},
		{"localhost", "127.0.0.1", false},
		{"localhost name", "localhost", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPrivateNetworkAddress(tt.host)
			if got != tt.want {
				t.Errorf("isPrivateNetworkAddress(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

// TestIsValidIPOctet tests the IP octet validation helper.
func TestIsValidIPOctet(t *testing.T) {
	tests := []struct {
		name  string
		octet string
		want  bool
	}{
		{"zero", "0", true},
		{"single digit", "5", true},
		{"double digit", "42", true},
		{"triple digit", "255", true},
		{"max valid", "255", true},
		{"min valid", "0", true},

		// Invalid.
		{"too large", "256", false},
		{"way too large", "999", false},
		{"empty", "", false},
		{"negative", "-1", false},
		{"letters", "abc", false},
		{"mixed", "12a", false},
		{"too long", "1234", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidIPOctet(tt.octet)
			if got != tt.want {
				t.Errorf("isValidIPOctet(%q) = %v, want %v", tt.octet, got, tt.want)
			}
		})
	}
}

// TestParseOctetInRange tests the octet parsing with range validation.
func TestParseOctetInRange(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		minVal int
		maxVal int
		want   int
		wantOk bool
	}{
		{"in range", "20", 16, 31, 20, true},
		{"at min", "16", 16, 31, 16, true},
		{"at max", "31", 16, 31, 31, true},
		{"below min", "15", 16, 31, 15, false},
		{"above max", "32", 16, 31, 32, false},
		{"empty string", "", 0, 255, 0, false},
		{"invalid chars", "abc", 0, 255, 0, false},
		{"too long", "1234", 0, 255, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseOctetInRange(tt.s, tt.minVal, tt.maxVal)
			if got != tt.want || ok != tt.wantOk {
				t.Errorf("parseOctetInRange(%q, %d, %d) = (%d, %v), want (%d, %v)",
					tt.s, tt.minVal, tt.maxVal, got, ok, tt.want, tt.wantOk)
			}
		})
	}
}
