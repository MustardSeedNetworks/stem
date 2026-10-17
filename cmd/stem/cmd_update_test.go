// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"strings"
	"testing"
)

// Flag combinations the verb refuses before it looks at privilege, the
// settings or the network.
func TestUpdateCmdRefusesConflictingFlags(t *testing.T) {
	for _, tc := range []struct {
		args    []string
		wantErr string
	}{
		{args: []string{"--scheduled", "--source", "/srv/stem-releases"}, wantErr: "--scheduled takes no other flag"},
		{args: []string{"--scheduled", "--automatic", "on"}, wantErr: "--scheduled takes no other flag"},
		{args: []string{"--automatic", "yes"}, wantErr: `takes "on" or "off", not "yes"`},
		{args: []string{"now"}, wantErr: `unexpected argument "now"`},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			err := updateCmd(tc.args)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("updateCmd(%q) = %v; want %q", tc.args, err, tc.wantErr)
			}
		})
	}
}
