// SPDX-License-Identifier: BUSL-1.1

package daemonconn

import (
	"path/filepath"
	"testing"
)

// Each answer must match where that platform's shipped service definition
// keeps state (deploy/systemd/stem.service, deploy/launchd/com.stem.plist),
// or the CLI never finds a daemon installed per docs/DEPLOYMENT.md (#1649).
func TestPackagedDataDirMatchesTheShippedServiceDefinition(t *testing.T) {
	programData := filepath.Join(t.TempDir(), "ProgramData")
	tests := []struct {
		goos        string
		programData string
		want        string
	}{
		{goos: "linux", want: "/var/lib/stem"},
		{goos: "freebsd", want: "/var/lib/stem"},
		{goos: "darwin", want: "/usr/local/stem"},
		{goos: "windows", programData: programData, want: filepath.Join(programData, "stem")},
		{goos: "windows", want: `C:\ProgramData\stem`},
	}
	for _, tt := range tests {
		t.Run(tt.goos+"/"+tt.programData, func(t *testing.T) {
			t.Setenv("ProgramData", tt.programData)
			if got := packagedDataDir(tt.goos); got != tt.want {
				t.Errorf("packagedDataDir(%q) = %q, want %q", tt.goos, got, tt.want)
			}
		})
	}
}
