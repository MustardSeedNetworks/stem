// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestResolveLogSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		goos     string
		file     string
		wantFile string
		wantErr  string
	}{
		{"an explicit file wins everywhere", "linux", "/tmp/captured.log", "/tmp/captured.log", ""},
		{"an explicit file on windows", "windows", `C:\logs\stem.log`, `C:\logs\stem.log`, ""},
		{
			"macOS is the launchd file under the data directory",
			"darwin",
			"",
			filepath.Join("/data", "logs", "stem.log"),
			"",
		},
		{"linux is the journal", "linux", "", "", ""},
		{
			"windows is the file the service writes under the data directory",
			"windows",
			"",
			filepath.Join("/data", "logs", "stem.log"),
			"",
		},
		{"other systems have none either", "freebsd", "", "", "stem writes no log file on freebsd"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			source, err := resolveLogSource(tc.goos, "/data", tc.file)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) ||
					!strings.Contains(err.Error(), "--file") {
					t.Fatalf("resolveLogSource = %v, want %q naming --file", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveLogSource: %v", err)
			}
			if source.file != tc.wantFile {
				t.Errorf("file = %q, want %q", source.file, tc.wantFile)
			}
		})
	}
}

func numberedLines(n int) string {
	var out strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&out, "line %d\n", i)
	}
	return out.String()
}

func TestShowLastLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		lines   int
		want    string
	}{
		{"fewer lines than asked", "a\nb\n", 10, "a\nb\n"},
		{"exactly as many", "a\nb\nc\n", 3, "a\nb\nc\n"},
		{"the last two", "a\nb\nc\nd\n", 2, "c\nd\n"},
		{"the last one", "a\nb\nc\n", 1, "c\n"},
		{"no trailing newline", "a\nb\nc", 2, "b\nc"},
		{"empty file", "", 5, ""},
		{"one unterminated line", "only", 3, "only"},
		{"spans several blocks", numberedLines(30000), 3, "line 29998\nline 29999\nline 30000\n"},
		{"asks for more than a block holds", numberedLines(30000), 29999, numberedLines(30000)[len("line 1\n"):]},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "stem.log")
			if err := os.WriteFile(path, []byte(tc.content), 0o600); err != nil {
				t.Fatal(err)
			}

			var out bytes.Buffer
			err := runLogs(context.Background(), logEnv{goos: "linux"}, path, tc.lines, false, &out, io.Discard)
			if err != nil {
				t.Fatalf("runLogs: %v", err)
			}
			if out.String() != tc.want {
				t.Errorf("got %q, want %q", out.String(), tc.want)
			}
		})
	}
}

func TestLogsReadsTheLaunchdFileOnMacOS(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dataDir, "logs"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "logs", "stem.log"), []byte("from launchd\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := runLogs(
		context.Background(),
		logEnv{goos: "darwin", dataDir: dataDir},
		"",
		10,
		false,
		&out,
		io.Discard,
	); err != nil {
		t.Fatalf("runLogs: %v", err)
	}
	if out.String() != "from launchd\n" {
		t.Errorf("got %q", out.String())
	}
}

func TestLogsMissingFile(t *testing.T) {
	t.Parallel()

	missing := filepath.Join(t.TempDir(), "stem.log")
	err := runLogs(context.Background(), logEnv{goos: "linux"}, missing, 10, false, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("runLogs succeeded on a missing file")
	}
	if strings.Contains(err.Error(), "launchd") {
		t.Errorf("an explicit --file was answered with the default-path hint: %v", err)
	}

	dataDir := t.TempDir()
	err = runLogs(context.Background(), logEnv{goos: "darwin", dataDir: dataDir}, "", 10, false, io.Discard, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "launchd") || !strings.Contains(err.Error(), "--file") {
		t.Errorf("a missing default log did not say where it looks: %v", err)
	}
}

func TestLogsOnLinuxHandsOffToJournalctl(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		follow  bool
		runner  func(args *[]string) journalRunner
		want    []string
		wantErr string
	}{
		{
			name: "one-shot",
			want: []string{"--unit", "stem", "--lines", "25", "--no-pager"},
			runner: func(args *[]string) journalRunner {
				return func(_ context.Context, got []string, stdout, _ io.Writer) error {
					*args = got
					_, _ = io.WriteString(stdout, "journal output\n")
					return nil
				}
			},
		},
		{
			name:   "follow",
			follow: true,
			want:   []string{"--unit", "stem", "--lines", "25", "--no-pager", "--follow"},
			runner: func(args *[]string) journalRunner {
				return func(_ context.Context, got []string, _, _ io.Writer) error { *args = got; return nil }
			},
		},
		{
			name:    "journalctl missing",
			wantErr: "journalctl is not installed",
			runner: func(*[]string) journalRunner {
				return func(context.Context, []string, io.Writer, io.Writer) error {
					return fmt.Errorf("start: %w", exec.ErrNotFound)
				}
			},
		},
		{
			name:    "journalctl fails",
			wantErr: "no journal files were found",
			runner: func(*[]string) journalRunner {
				return func(context.Context, []string, io.Writer, io.Writer) error {
					return errors.New("no journal files were found")
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var args []string
			var out bytes.Buffer
			env := logEnv{goos: "linux", journal: tc.runner(&args)}
			err := runLogs(context.Background(), env, "", 25, tc.follow, &out, io.Discard)

			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("runLogs = %v, want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("runLogs: %v", err)
			}
			if strings.Join(args, " ") != strings.Join(tc.want, " ") {
				t.Errorf("journalctl args = %v, want %v", args, tc.want)
			}
		})
	}
}

// Interrupting a followed journal is how an operator stops it; journalctl
// dies with the context, and that is not a failure.
func TestLogsInterruptedJournalIsNotAnError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	env := logEnv{goos: "linux", journal: func(ctx context.Context, _ []string, _, _ io.Writer) error {
		cancel()
		<-ctx.Done()
		return errors.New("signal: killed")
	}}
	if err := runLogs(ctx, env, "", 10, true, io.Discard, io.Discard); err != nil {
		t.Fatalf("runLogs = %v", err)
	}
}

// safeBuffer is a buffer a follower can write while the test reads it.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func eventually(t *testing.T, what string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func appendTo(t *testing.T, path, text string) {
	t.Helper()
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = file.WriteString(text); err != nil {
		t.Fatal(err)
	}
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestLogsFollow(t *testing.T) {
	tests := []struct {
		name string
		// act changes the file after the follower has printed its first lines.
		act  func(t *testing.T, path string)
		want string
	}{
		{
			name: "appended lines",
			act:  func(t *testing.T, path string) { appendTo(t, path, "new 1\nnew 2\n") },
			want: "old 1\nold 2\nnew 1\nnew 2\n",
		},
		{
			name: "rotation to a new file",
			act: func(t *testing.T, path string) {
				if err := os.Rename(path, path+".1"); err != nil {
					t.Fatal(err)
				}
				appendTo(t, path, "after rotation\n")
			},
			want: "old 1\nold 2\nafter rotation\n",
		},
		{
			name: "truncation in place",
			act: func(t *testing.T, path string) {
				if err := os.WriteFile(path, []byte("t\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			want: "old 1\nold 2\nt\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "stem.log")
			appendTo(t, path, "old 1\nold 2\n")

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var out safeBuffer
			done := make(chan error, 1)
			go func() {
				done <- runLogs(ctx, logEnv{goos: "linux", interval: 5 * time.Millisecond}, path, 10, true, &out, io.Discard)
			}()

			eventually(t, "the first lines", func() bool { return out.String() == "old 1\nold 2\n" })
			tc.act(t, path)
			eventually(t, "the followed output", func() bool { return out.String() == tc.want })

			cancel()
			if err := <-done; err != nil {
				t.Fatalf("runLogs = %v", err)
			}
		})
	}
}

func TestLogsCmdRejectsBadInvocations(t *testing.T) {
	for _, args := range [][]string{{"--lines", "0"}, {"-n", "-3"}, {"extra"}, {"--nope"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stderr bytes.Buffer
			err := logsCmd(args, io.Discard, &stderr)
			if got := exitCodeOf(err); got != exitError {
				t.Fatalf("exit status = %d, want %d (err %v)", got, exitError, err)
			}
		})
	}
}

func TestLogsCmdReadsAFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stem.log")
	if err := os.WriteFile(path, []byte(numberedLines(5)), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	if err := logsCmd([]string{"--file", path, "-n", "2"}, &stdout, io.Discard); err != nil {
		t.Fatalf("logsCmd: %v", err)
	}
	if stdout.String() != "line 4\nline 5\n" {
		t.Errorf("got %q", stdout.String())
	}
}
