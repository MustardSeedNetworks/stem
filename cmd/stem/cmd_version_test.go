// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/version"
)

func setBuild(t *testing.T) {
	t.Helper()

	previous := [4]string{version.Version, version.Commit, version.BuildTime, version.UIBuildHash}
	version.Version, version.Commit, version.BuildTime, version.UIBuildHash = "v9.8.7", "abc1234", "2026-11-21T09:00:00Z", "cafe"
	t.Cleanup(func() {
		version.Version, version.Commit, version.BuildTime, version.UIBuildHash = previous[0], previous[1], previous[2], previous[3]
	})
}

func TestVersionCmdPlainOutputKeepsItsFirstLine(t *testing.T) {
	setBuild(t)

	var stdout bytes.Buffer
	if err := versionCmd(nil, &stdout, io.Discard); err != nil {
		t.Fatalf("versionCmd: %v", err)
	}

	// `stem update` health-checks the installed binary by reading this line.
	first, _, _ := strings.Cut(stdout.String(), "\n")
	if first != "The Stem v9.8.7" {
		t.Errorf("first line = %q", first)
	}
	if !strings.Contains(stdout.String(), "Commit: abc1234") {
		t.Errorf("output = %q", stdout.String())
	}
}

func TestVersionCmdJSONIsTheVersionDocument(t *testing.T) {
	setBuild(t)

	var stdout bytes.Buffer
	if err := versionCmd([]string{"--json"}, &stdout, io.Discard); err != nil {
		t.Fatalf("versionCmd: %v", err)
	}
	var got map[string]string
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("--json did not print an object: %v\n%s", err, stdout.String())
	}
	want := map[string]string{
		"version": "v9.8.7", "commit": "abc1234", "buildTime": "2026-11-21T09:00:00Z",
		"releaseTrain": "2026.11", "uiBuildHash": "cafe",
	}
	for key, value := range want {
		if got[key] != value {
			t.Errorf("%s = %q, want %q", key, got[key], value)
		}
	}
	if len(got) != len(want) {
		t.Errorf("keys = %v, want exactly %v", got, want)
	}
}

func TestVersionCmdRejectsBadInvocations(t *testing.T) {
	for _, args := range [][]string{{"extra"}, {"--nope"}, {"--json", "extra"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			err := versionCmd(args, io.Discard, io.Discard)
			if got := exitCodeOf(err); got != exitError {
				t.Fatalf("exit status = %d, want %d (err %v)", got, exitError, err)
			}
		})
	}
}
