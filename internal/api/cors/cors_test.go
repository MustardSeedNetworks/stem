// SPDX-License-Identifier: BUSL-1.1

package cors_test

import (
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api/cors"
)

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
			got := cors.IsLocalhostOrigin(tt.origin)
			if got != tt.want {
				t.Errorf("IsLocalhostOrigin(%q) = %v, want %v", tt.origin, got, tt.want)
			}
		})
	}
}

// TestIsLocalhostOriginAdditional tests additional IsLocalhostOrigin cases.
func TestIsLocalhostOriginAdditional(t *testing.T) {
	tests := []struct {
		origin   string
		expected bool
	}{
		{"http://[::1]", true},
		{"https://[::1]", true},
		{"http://127.0.0.1", true},
		{"https://127.0.0.1", true},
		{"http://localhost", true},
		{"https://localhost", true},
		{"http://192.168.1.1", false},
		{"https://example.com", false},
		{"", false},
		{"not-a-url", false},
	}

	for _, tt := range tests {
		t.Run(tt.origin, func(t *testing.T) {
			result := cors.IsLocalhostOrigin(tt.origin)
			if result != tt.expected {
				t.Errorf("IsLocalhostOrigin(%s) = %v, expected %v", tt.origin, result, tt.expected)
			}
		})
	}
}

// TestIsSameOrigin tests the IsSameOrigin function.
func TestIsSameOrigin(t *testing.T) {
	tests := []struct {
		name        string
		origin      string
		requestHost string
		want        bool
	}{
		{"same host and port", "http://10.0.0.1:8080", "10.0.0.1:8080", true},
		{"different port", "http://10.0.0.1:8080", "10.0.0.1:9090", false},
		{"different host", "http://10.0.0.1:8080", "10.0.0.2:8080", false},
		{"invalid URL", "not-a-url", "10.0.0.1:8080", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cors.IsSameOrigin(tt.origin, tt.requestHost)
			if got != tt.want {
				t.Errorf(
					"IsSameOrigin(%s, %s) = %v, want %v",
					tt.origin,
					tt.requestHost,
					got,
					tt.want,
				)
			}
		})
	}
}

// TestIsSameOriginAdditional tests additional IsSameOrigin cases.
func TestIsSameOriginAdditional(t *testing.T) {
	tests := []struct {
		origin      string
		requestHost string
		expected    bool
	}{
		{"http://192.168.1.1:8080", "192.168.1.1:8080", true},
		{"http://192.168.1.1", "192.168.1.1", true},
		{"https://example.com:443", "example.com:443", true},
		{"http://192.168.1.1:8080", "192.168.1.2:8080", false},
		{"http://192.168.1.1:8080", "192.168.1.1:9090", false},
		{"", "192.168.1.1:8080", false},
	}

	for _, tt := range tests {
		t.Run(tt.origin+"_"+tt.requestHost, func(t *testing.T) {
			result := cors.IsSameOrigin(tt.origin, tt.requestHost)
			if result != tt.expected {
				t.Errorf(
					"IsSameOrigin(%s, %s) = %v, expected %v",
					tt.origin,
					tt.requestHost,
					result,
					tt.expected,
				)
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

		// Localhost should NOT be matched by RFC 1918 (handled by IsLocalhostOrigin).
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
			got := cors.IsRFC1918Origin(tt.origin)
			if got != tt.want {
				t.Errorf("IsRFC1918Origin(%q) = %v, want %v", tt.origin, got, tt.want)
			}
		})
	}
}
