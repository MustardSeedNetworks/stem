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
	"runtime"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/foundation/pkg/instance"
)

// stateDir is a data directory holding the state a backup carries and some it
// does not, with STEM_DATA_DIR pointing at it.
func stateDir(t *testing.T) string {
	t.Helper()

	dir := filepath.Join(t.TempDir(), "stem")
	for name, content := range map[string]string{
		"auth/credentials.sealed": `{"username":"admin","passwordHash":"hash"}`,
		"certs/server.crt":        "CERT",
		"certs/server.key":        "KEY",
		".setup-token":            "never archived",
	} {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("STEM_DATA_DIR", dir)
	return dir
}

func archiveEntries(t *testing.T, path string) []string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for tr := tar.NewReader(gz); ; {
		header, nextErr := tr.Next()
		if errors.Is(nextErr, io.EOF) {
			return names
		}
		if nextErr != nil {
			t.Fatal(nextErr)
		}
		names = append(names, header.Name)
	}
}

func TestBackupCmdWritesAnOwnerOnlyArchive(t *testing.T) {
	stateDir(t)
	archive := filepath.Join(t.TempDir(), "stem.tar.gz")

	var stdout bytes.Buffer
	if err := backupCmd([]string{"-o", archive}, &stdout, io.Discard); err != nil {
		t.Fatalf("backupCmd: %v", err)
	}

	got := strings.Join(archiveEntries(t, archive), ",")
	if got != "auth/credentials.sealed,certs/,certs/server.crt,certs/server.key" {
		t.Errorf("archive entries = %s", got)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(archive)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Errorf("archive mode = %v, want 0600", info.Mode().Perm())
		}
	}
	if !strings.Contains(stdout.String(), archive) {
		t.Errorf("stdout = %q", stdout.String())
	}
}

func TestBackupCmdNeverOverwrites(t *testing.T) {
	stateDir(t)
	archive := filepath.Join(t.TempDir(), "stem.tar.gz")
	if err := os.WriteFile(archive, []byte("precious"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := backupCmd([]string{"-o", archive}, io.Discard, io.Discard)
	if got := exitCodeOf(err); got != exitError {
		t.Fatalf("exit status = %d (err %v)", got, err)
	}
	if content, _ := os.ReadFile(archive); string(content) != "precious" {
		t.Errorf("the existing archive was changed: %q", content)
	}
}

func TestBackupCmdLeavesNoHalfWrittenArchive(t *testing.T) {
	t.Setenv("STEM_DATA_DIR", t.TempDir())
	archive := filepath.Join(t.TempDir(), "stem.tar.gz")

	var stderr bytes.Buffer
	err := backupCmd([]string{"-o", archive}, io.Discard, &stderr)
	if got := exitCodeOf(err); got != exitError {
		t.Fatalf("exit status = %d (err %v)", got, err)
	}
	if _, statErr := os.Stat(archive); statErr == nil {
		t.Error("a failed backup left its archive behind")
	}
	if !strings.Contains(stderr.String(), "no stem state found") {
		t.Errorf("stderr = %q", stderr.String())
	}
}

func TestBackupCmdNeedsTheOutputDirectory(t *testing.T) {
	stateDir(t)
	err := backupCmd([]string{"-o", filepath.Join(t.TempDir(), "missing", "b.tar.gz")}, io.Discard, io.Discard)
	if got := exitCodeOf(err); got != exitError {
		t.Fatalf("exit status = %d (err %v)", got, err)
	}
}

func TestBackupCmdTakesNoPositionalArgument(t *testing.T) {
	stateDir(t)
	err := backupCmd([]string{"stem.tar.gz"}, io.Discard, io.Discard)
	if got := exitCodeOf(err); got != exitError {
		t.Fatalf("exit status = %d (err %v)", got, err)
	}
}

func TestRestoreCmdRoundTripKeepsThePreviousState(t *testing.T) {
	dir := stateDir(t)
	archive := filepath.Join(t.TempDir(), "stem.tar.gz")
	if err := backupCmd([]string{"-o", archive}, io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "auth/credentials.sealed"), []byte("drifted"), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	if err := restoreCmd([]string{archive}, &stdout, io.Discard); err != nil {
		t.Fatalf("restoreCmd: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(dir, "auth/credentials.sealed"))
	if string(got) != `{"username":"admin","passwordHash":"hash"}` {
		t.Errorf("auth/credentials.sealed = %q", got)
	}
	matches, _ := filepath.Glob(dir + ".pre-restore-*")
	if len(matches) != 1 {
		t.Fatalf("pre-restore copies = %v", matches)
	}
	kept, _ := os.ReadFile(filepath.Join(matches[0], "auth/credentials.sealed"))
	if string(kept) != "drifted" {
		t.Errorf("the pre-restore copy holds %q", kept)
	}
	if !strings.Contains(stdout.String(), matches[0]) {
		t.Errorf("stdout does not name the pre-restore copy: %q", stdout.String())
	}
}

func TestRestoreCmdRefusedWhileTheDaemonRuns(t *testing.T) {
	dir := stateDir(t)
	archive := filepath.Join(t.TempDir(), "stem.tar.gz")
	if err := backupCmd([]string{"-o", archive}, io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	lock, err := instance.Acquire(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Release() }()

	var stderr bytes.Buffer
	err = restoreCmd([]string{archive}, io.Discard, &stderr)
	if got := exitCodeOf(err); got != exitError {
		t.Fatalf("exit status = %d (err %v)", got, err)
	}
	if !strings.Contains(stderr.String(), "the daemon is running") {
		t.Errorf("stderr = %q", stderr.String())
	}
}

func TestRestoreCmdRefusesATraversingArchive(t *testing.T) {
	dir := stateDir(t)
	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: "../evil", Typeflag: tar.TypeReg, Mode: 0o600, Size: 1}); err != nil {
		t.Fatal(err)
	}
	_, _ = tw.Write([]byte("x"))
	_ = tw.Close()
	_ = gz.Close()
	archive := filepath.Join(t.TempDir(), "evil.tar.gz")
	if err := os.WriteFile(archive, buffer.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer
	err := restoreCmd([]string{archive}, io.Discard, &stderr)
	if got := exitCodeOf(err); got != exitError {
		t.Fatalf("exit status = %d (err %v)", got, err)
	}
	if !strings.Contains(stderr.String(), "unsafe archive entry") {
		t.Errorf("stderr = %q", stderr.String())
	}
	if _, statErr := os.Stat(filepath.Join(filepath.Dir(dir), "evil")); statErr == nil {
		t.Error("the traversing entry was written")
	}
}

func TestRestoreCmdBadInvocations(t *testing.T) {
	stateDir(t)
	for _, args := range [][]string{nil, {"a", "b"}, {filepath.Join(t.TempDir(), "missing.tar.gz")}, {"--nope"}} {
		err := restoreCmd(args, io.Discard, io.Discard)
		if got := exitCodeOf(err); got != exitError {
			t.Errorf("restore %v: exit status = %d (err %v)", args, got, err)
		}
	}
}
