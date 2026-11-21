// SPDX-License-Identifier: BUSL-1.1

package support_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/foundation/pkg/instance"

	"github.com/MustardSeedNetworks/stem/internal/support"
)

func restoreTime() time.Time { return time.Date(2026, 11, 21, 9, 0, 0, 0, time.UTC) }

// managedFiles is the state a backup carries, as it sits in a data directory.
func managedFiles() map[string]string {
	return map[string]string{
		"credentials.json":      `{"username":"admin","passwordHash":"$argon2id$v=19$m=1,t=1,p=1$abc$def"}`,
		"reflector.json":        `{"interface":"eth0"}`,
		"update.json":           `{"automatic":false}`,
		".daemon.json":          `{"url":"https://127.0.0.1:8444","token":"t"}`,
		".config/stem/.license": "MSN1.payload.signature",
		"certs/server.crt":      "CERT",
		"certs/server.key":      "KEY",
	}
}

// leftAlone is in the data directory and in no backup.
func leftAlone() map[string]string {
	return map[string]string{
		instance.LockFileName: "",
		"run.lock":            "",
		".setup-token":        "setup",
		".recovery-token":     "recovery",
		"logs/stem.log":       "log line",
	}
}

func writeTree(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func readTree(t *testing.T, dir string) map[string]string {
	t.Helper()
	tree := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		rel, _ := filepath.Rel(dir, path)
		tree[filepath.ToSlash(rel)] = string(content)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return tree
}

func backupOf(t *testing.T, dir string) []byte {
	t.Helper()
	var archive bytes.Buffer
	if err := support.Backup(dir, &archive); err != nil {
		t.Fatalf("Backup: %v", err)
	}
	return archive.Bytes()
}

func entryNames(t *testing.T, archive []byte) []string {
	t.Helper()
	gz, err := gzip.NewReader(bytes.NewReader(archive))
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

func TestBackupCarriesExactlyTheManagedState(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, managedFiles())
	writeTree(t, dir, leftAlone())

	got := entryNames(t, backupOf(t, dir))
	want := []string{
		".config/stem/.license", ".daemon.json", "certs/", "certs/server.crt", "certs/server.key",
		"credentials.json", "reflector.json", "update.json",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("archive entries = %v, want %v", got, want)
	}
}

func TestBackupIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, managedFiles())

	if first, second := backupOf(t, dir), backupOf(t, dir); !bytes.Equal(first, second) {
		t.Fatal("two backups of unchanged state differ")
	}
}

func TestBackupRefusesAnEmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, leftAlone())

	err := support.Backup(dir, io.Discard)
	if !errors.Is(err, support.ErrNothingToBackUp) {
		t.Fatalf("Backup = %v, want ErrNothingToBackUp", err)
	}
}

func TestBackupSkipsALinkInCerts(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating a symlink needs a privilege Windows does not grant by default")
	}
	dir := t.TempDir()
	writeTree(t, dir, managedFiles())
	if err := os.Symlink("/etc/hostname", filepath.Join(dir, "certs", "link")); err != nil {
		t.Fatal(err)
	}

	for _, name := range entryNames(t, backupOf(t, dir)) {
		if strings.Contains(name, "link") {
			t.Fatalf("archive carries the link: %v", name)
		}
	}
}

func TestRestoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, managedFiles())
	archive := backupOf(t, dir)

	// The directory drifts after the backup: a changed file, a stale
	// certificate, a missing file, and the files a backup never carries.
	writeTree(t, dir, map[string]string{
		"credentials.json": `{"username":"changed"}`,
		"certs/stale.crt":  "STALE",
		".setup-token":     "must survive",
		"logs/stem.log":    "must survive",
	})
	if err := os.Remove(filepath.Join(dir, "reflector.json")); err != nil {
		t.Fatal(err)
	}

	previous, err := support.Restore(bytes.NewReader(archive), dir, restoreTime())
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}

	got := readTree(t, dir)
	for name, want := range managedFiles() {
		if got[name] != want {
			t.Errorf("%s = %q, want %q", name, got[name], want)
		}
	}
	if _, stale := got["certs/stale.crt"]; stale {
		t.Error("a certificate the backup does not carry survived the restore")
	}
	if got[".setup-token"] != "must survive" || got["logs/stem.log"] != "must survive" {
		t.Error("a file outside the managed state was touched")
	}
	for name := range got {
		if strings.HasPrefix(name, ".restore-") {
			t.Errorf("staging left behind: %s", name)
		}
	}

	if want := dir + ".pre-restore-20261121T090000Z"; previous != want {
		t.Errorf("previous state kept at %s, want %s", previous, want)
	}
	t.Cleanup(func() { _ = os.RemoveAll(previous) })
	kept := readTree(t, previous)
	if kept["credentials.json"] != `{"username":"changed"}` || kept["certs/stale.crt"] != "STALE" {
		t.Errorf("the pre-restore copy does not hold the state before the restore: %v", kept)
	}
}

func TestRestoredFilesAreOwnerOnly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows carries no POSIX mode bits")
	}
	dir := t.TempDir()
	writeTree(t, dir, managedFiles())
	archive := backupOf(t, dir)
	for name := range managedFiles() {
		if err := os.Chmod(filepath.Join(dir, filepath.FromSlash(name)), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	previous, err := support.Restore(bytes.NewReader(archive), dir, restoreTime())
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(previous) })

	for name := range managedFiles() {
		info, statErr := os.Stat(filepath.Join(dir, filepath.FromSlash(name)))
		if statErr != nil {
			t.Fatal(statErr)
		}
		if info.Mode().Perm() != 0o600 {
			t.Errorf("%s restored with mode %v, want 0600", name, info.Mode().Perm())
		}
	}
}

func TestRestoreRefusesWhileTheDaemonRuns(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, managedFiles())
	archive := backupOf(t, dir)
	writeTree(t, dir, map[string]string{"credentials.json": "current"})

	lock, err := instance.Acquire(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Release() }()

	_, err = support.Restore(bytes.NewReader(archive), dir, restoreTime())
	if !errors.Is(err, support.ErrDaemonRunning) {
		t.Fatalf("Restore = %v, want ErrDaemonRunning", err)
	}
	if got := readTree(t, dir)["credentials.json"]; got != "current" {
		t.Errorf("credentials.json = %q, a refused restore changed it", got)
	}
	if _, statErr := os.Stat(dir + ".pre-restore-20261121T090000Z"); statErr == nil {
		t.Error("a refused restore left a pre-restore copy")
	}
}

func TestRestoreNeedsTheDataDirectory(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")

	if _, err := support.Restore(bytes.NewReader(nil), missing, restoreTime()); err == nil {
		t.Fatal("Restore created a data directory that was not there")
	}
	if _, err := os.Stat(missing); err == nil {
		t.Fatal("the missing data directory now exists")
	}
}

type tarEntry struct {
	name     string
	typeflag byte
	body     string
	size     int64
}

func archiveOf(t *testing.T, entries ...tarEntry) []byte {
	t.Helper()
	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)
	tw := tar.NewWriter(gz)
	for _, entry := range entries {
		typeflag := entry.typeflag
		if typeflag == 0 {
			typeflag = tar.TypeReg
		}
		size := int64(len(entry.body))
		if entry.size != 0 {
			size = entry.size
		}
		header := &tar.Header{Name: entry.name, Typeflag: typeflag, Mode: 0o600, Size: size, Linkname: "target"}
		if typeflag == tar.TypeDir || typeflag == tar.TypeSymlink || typeflag == tar.TypeLink {
			header.Size = 0
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if header.Size > 0 {
			body := entry.body
			if int64(len(body)) < header.Size {
				body += strings.Repeat("x", int(header.Size)-len(body))
			}
			if _, err := tw.Write([]byte(body)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestRestoreRefusesAnUnsafeArchiveAndChangesNothing(t *testing.T) {
	good := tarEntry{name: "credentials.json", body: "from the archive"}
	many := make([]tarEntry, 0, 300)
	for i := range 300 {
		many = append(
			many,
			tarEntry{name: "certs/f" + strings.Repeat("x", i%5) + string(rune('a'+i%26)) + string(rune('a'+i/26))},
		)
	}

	tests := []struct {
		name       string
		archive    []byte
		wantUnsafe bool
	}{
		{"parent traversal", archiveOf(t, good, tarEntry{name: "../evil", body: "x"}), true},
		{"traversal through a managed prefix", archiveOf(t, good, tarEntry{name: "certs/../../evil", body: "x"}), true},
		{
			"traversal that cleans to an unmanaged name",
			archiveOf(t, good, tarEntry{name: "certs/../evil", body: "x"}),
			true,
		},
		{"absolute path", archiveOf(t, good, tarEntry{name: "/etc/evil", body: "x"}), true},
		{"backslash path", archiveOf(t, good, tarEntry{name: `certs\..\evil`, body: "x"}), true},
		{"symlink", archiveOf(t, good, tarEntry{name: "certs/link", typeflag: tar.TypeSymlink}), true},
		{"hard link", archiveOf(t, good, tarEntry{name: "certs/link", typeflag: tar.TypeLink}), true},
		{"device", archiveOf(t, good, tarEntry{name: "certs/dev", typeflag: tar.TypeChar}), true},
		{"first-run token", archiveOf(t, good, tarEntry{name: ".setup-token", body: "x"}), true},
		{"recovery token", archiveOf(t, good, tarEntry{name: ".recovery-token", body: "x"}), true},
		{"run lock", archiveOf(t, good, tarEntry{name: "run.lock", body: "x"}), true},
		{"instance lock", archiveOf(t, good, tarEntry{name: instance.LockFileName, body: "x"}), true},
		{"an unmanaged directory", archiveOf(t, good, tarEntry{name: "logs/", typeflag: tar.TypeDir}), true},
		{"oversized entry", archiveOf(t, good, tarEntry{name: "certs/big", size: 8<<20 + 1}), true},
		{"too many entries", archiveOf(t, many...), true},
		{"duplicate entry", archiveOf(t, good, good), false},
		{"no state at all", archiveOf(t), false},
		{"not gzip", []byte("this is not an archive"), false},
		{"truncated", archiveOf(t, good, tarEntry{name: "certs/server.crt", body: "x"})[:30], false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parent := t.TempDir()
			dir := filepath.Join(parent, "stem")
			if err := os.Mkdir(dir, 0o700); err != nil {
				t.Fatal(err)
			}
			writeTree(t, dir, map[string]string{"credentials.json": "current"})

			_, err := support.Restore(bytes.NewReader(tc.archive), dir, restoreTime())
			if err == nil {
				t.Fatal("Restore accepted the archive")
			}
			if tc.wantUnsafe && !errors.Is(err, support.ErrUnsafeEntry) {
				t.Errorf("Restore = %v, want ErrUnsafeEntry", err)
			}

			if got := readTree(t, dir); len(got) != 2 || got["credentials.json"] != "current" {
				t.Errorf("a refused archive changed the data directory: %v", got)
			}
			leftovers, _ := os.ReadDir(parent)
			if len(leftovers) != 1 {
				t.Errorf("a refused archive left files beside the data directory: %v", leftovers)
			}
		})
	}
}

func TestBackupRefusesWhileTheDaemonRuns(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, managedFiles())
	lock, err := instance.Acquire(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Release() }()

	var archive bytes.Buffer
	if err = support.Backup(dir, &archive); !errors.Is(err, support.ErrDaemonRunning) {
		t.Fatalf("Backup = %v, want ErrDaemonRunning", err)
	}
	if archive.Len() != 0 {
		t.Error("a refused backup wrote output")
	}
}

func linkTestArchive(t *testing.T) []byte {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("creating a symlink needs a privilege Windows does not grant by default")
	}
	dir := t.TempDir()
	writeTree(t, dir, managedFiles())
	return backupOf(t, dir)
}

func TestRestoreRefusesADataDirectoryThatIsALink(t *testing.T) {
	archive := linkTestArchive(t)
	target := t.TempDir()
	writeTree(t, target, map[string]string{"credentials.json": "current"})
	link := filepath.Join(t.TempDir(), "stem")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	if _, err := support.Restore(bytes.NewReader(archive), link, restoreTime()); err == nil {
		t.Fatal("Restore followed a symlinked data directory")
	}
	if got := readTree(t, target)["credentials.json"]; got != "current" {
		t.Errorf("credentials.json = %q, the link's target was changed", got)
	}
}

func TestRestoreRefusesAManagedFileThatIsALink(t *testing.T) {
	archive := linkTestArchive(t)
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "credentials.json")); err != nil {
		t.Fatal(err)
	}

	if _, err := support.Restore(bytes.NewReader(archive), dir, restoreTime()); err == nil {
		t.Fatal("Restore replaced a link")
	}
	if got, _ := os.ReadFile(outside); string(got) != "outside" {
		t.Errorf("the link's target was written: %q", got)
	}
}

func TestRestoreRefusesALinkedCertsDirectory(t *testing.T) {
	archive := linkTestArchive(t)
	dir := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(dir, "certs")); err != nil {
		t.Fatal(err)
	}

	if _, err := support.Restore(bytes.NewReader(archive), dir, restoreTime()); err == nil {
		t.Fatal("Restore wrote through a linked certs directory")
	}
	if entries, _ := os.ReadDir(outside); len(entries) != 0 {
		t.Errorf("files were written outside the data directory: %v", entries)
	}
}
