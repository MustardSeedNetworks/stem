// SPDX-License-Identifier: BUSL-1.1

package logging

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func restoreDefaultLogger(t *testing.T) {
	t.Helper()
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })
}

// A configured file appears, in a directory only its owner can enter, and
// carries what the daemon logs.
func TestInitWritesTheConfiguredFile(t *testing.T) {
	restoreDefaultLogger(t)
	path := filepath.Join(t.TempDir(), "logs", "stem.log")

	err := Init(&Config{Level: LevelInfo, Format: FormatText, File: path, MaxSize: 1, MaxBackups: 2})
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	Get().Info("a line for the file", "event", "test")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("the log file did not appear: %v", err)
	}
	if !strings.Contains(string(data), "a line for the file") {
		t.Errorf("log file = %q", data)
	}
	if runtime.GOOS != "windows" {
		info, statErr := os.Stat(filepath.Dir(path))
		if statErr != nil {
			t.Fatal(statErr)
		}
		if info.Mode().Perm() != 0o700 {
			t.Errorf("log directory mode = %v, want 0700", info.Mode().Perm())
		}
	}
}

// Past MaxSize the file rotates: the old lines move to a backup and the
// live file starts over.
func TestInitRotatesTheFile(t *testing.T) {
	restoreDefaultLogger(t)
	dir := filepath.Join(t.TempDir(), "logs")
	path := filepath.Join(dir, "stem.log")

	if err := Init(&Config{Level: LevelInfo, Format: FormatText, File: path, MaxSize: 1, MaxBackups: 2}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	filler := strings.Repeat("x", 1000)
	for range 1200 { // about 1.2 MB, over the 1 MB limit
		Get().Info("fill", "padding", filler)
	}
	Get().Info("after rotation")

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 2 {
		t.Fatalf("log directory holds %d files, want the live file and a rotated one", len(entries))
	}
	live, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(live), "after rotation") || len(live) > 1<<20 {
		t.Errorf("live file is %d bytes and lacks the newest line", len(live))
	}
}

func TestInitRefusesAnUncreatableLogDirectory(t *testing.T) {
	restoreDefaultLogger(t)
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	err := Init(&Config{Level: LevelInfo, Format: FormatText, File: filepath.Join(blocker, "logs", "stem.log")})
	if err == nil {
		t.Fatal("Init accepted a log directory it cannot create")
	}
}
