// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
)

const (
	sentinelHash  = "SENTINEL-HASH-aa11"
	sentinelToken = "SENTINEL-TOKEN-bb22"
	sentinelKey   = "SENTINEL-KEY-cc33"
	sentinelEnv   = "SENTINEL-ENV-dd44"
	sentinelLog   = "SENTINEL-LOG-ee55"
)

func readBundle(t *testing.T, path string) map[string]string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for tr := tar.NewReader(gz); ; {
		header, nextErr := tr.Next()
		if errors.Is(nextErr, io.EOF) {
			return files
		}
		if nextErr != nil {
			t.Fatal(nextErr)
		}
		body, _ := io.ReadAll(tr)
		files[header.Name] = string(body)
	}
}

func TestSupportBundleCmd(t *testing.T) {
	dir := stateDir(t)
	for name, content := range map[string]string{
		"auth/credentials.sealed": `{"username":"admin","passwordHash":"` + sentinelHash + `"}`,
		"certs/server.key":        "-----BEGIN PRIVATE KEY-----\n" + sentinelKey + "\n-----END PRIVATE KEY-----\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(name)), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	// A stale descriptor: the daemon is "not running", which is a normal
	// thing for a bundle to record, and nothing on this host is contacted.
	if err := daemonconn.Publish(dir, daemonconn.Descriptor{URL: closedURL(t), Token: sentinelToken}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("STEM_JWT_SECRET", sentinelEnv)

	logFile := filepath.Join(t.TempDir(), "stem.log")
	logText := `{"level":"info","message":"started"}` + "\n" +
		`{"level":"warn","password":"` + sentinelLog + `"}` + "\n" +
		"jwt " + sentinelEnv + " token " + sentinelToken + "\n"
	if err := os.WriteFile(logFile, []byte(logText), 0o600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "bundle.tar.gz")

	var stdout, stderr bytes.Buffer
	if err := supportBundleCmd([]string{"-o", output, "--log", logFile}, &stdout, &stderr); err != nil {
		t.Fatalf("supportBundleCmd: %v\n%s", err, stderr.String())
	}

	files := readBundle(t, output)
	var everything strings.Builder
	for name, body := range files {
		everything.WriteString(name + "\n" + body + "\n")
	}
	for _, sentinel := range []string{sentinelHash, sentinelToken, sentinelKey, sentinelEnv, sentinelLog} {
		if strings.Contains(everything.String(), sentinel) {
			t.Errorf("the bundle carries %q", sentinel)
		}
	}
	for _, want := range []string{"manifest.json", "status.json", "logs/stem.log"} {
		if _, ok := files[want]; !ok {
			t.Errorf("the bundle lacks %s", want)
		}
	}
	if !strings.Contains(files["status.json"], stateNotRunning) {
		t.Errorf("status.json = %s", files["status.json"])
	}
	if !strings.Contains(files["logs/stem.log"], `"message":"started"`) {
		t.Errorf("the log lost an ordinary line: %q", files["logs/stem.log"])
	}
	if !strings.Contains(stdout.String(), output) {
		t.Errorf("stdout = %q", stdout.String())
	}
}

func TestSupportBundleCmdWithoutALogStillWritesTheBundle(t *testing.T) {
	stateDir(t)
	output := filepath.Join(t.TempDir(), "bundle.tar.gz")

	var stderr bytes.Buffer
	missing := filepath.Join(t.TempDir(), "missing.log")
	if err := supportBundleCmd([]string{"-o", output, "--log", missing}, io.Discard, &stderr); err != nil {
		t.Fatalf("supportBundleCmd: %v", err)
	}
	if !strings.Contains(stderr.String(), "the log is not included") {
		t.Errorf("stderr = %q", stderr.String())
	}
	if _, ok := readBundle(t, output)["manifest.json"]; !ok {
		t.Error("the bundle has no manifest")
	}
}

func TestSupportBundleCmdRefusesBadInvocations(t *testing.T) {
	stateDir(t)
	existing := filepath.Join(t.TempDir(), "bundle.tar.gz")
	if err := os.WriteFile(existing, []byte("precious"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		args []string
	}{
		{"stray argument", []string{"bundle.tar.gz"}},
		{"lines out of range", []string{"--lines", "0"}},
		{"unknown flag", []string{"--nope"}},
		{"an existing output is not overwritten", []string{"-o", existing}},
		{"output directory missing", []string{"-o", filepath.Join(t.TempDir(), "missing", "b.tar.gz")}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := supportBundleCmd(tc.args, io.Discard, io.Discard)
			if got := exitCodeOf(err); got != exitError {
				t.Fatalf("exit status = %d, want %d (err %v)", got, exitError, err)
			}
		})
	}
	if content, _ := os.ReadFile(existing); string(content) != "precious" {
		t.Errorf("the existing file was changed: %q", content)
	}
}
