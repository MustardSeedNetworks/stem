// SPDX-License-Identifier: BUSL-1.1

package cors

import "testing"

// These tests live in the internal test package so they can exercise the
// unexported IP-validation helpers directly.

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
