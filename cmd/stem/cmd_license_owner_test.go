// SPDX-License-Identifier: BUSL-1.1

package main

import "testing"

func TestDecideLicenseWrite(t *testing.T) {
	tests := []struct {
		name  string
		euid  int
		owner int
		want  licenseWriteAction
	}{
		{"the service account itself", 998, 998, writeAsIs},
		{"root where root owns the directory (macOS)", 0, 0, writeAsIs},
		{"root where the service account owns it (Linux)", 0, 998, writeAsOwner},
		{"another user", 1000, 998, writeRefused},
		{"another user where root owns it", 1000, 0, writeRefused},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decideLicenseWrite(tt.euid, tt.owner); got != tt.want {
				t.Errorf("decideLicenseWrite(%d, %d) = %v, want %v", tt.euid, tt.owner, got, tt.want)
			}
		})
	}
}
