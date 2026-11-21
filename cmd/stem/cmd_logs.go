// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
)

const (
	logsCommandName = "logs"

	defaultLogLines = 100

	// journalUnit is the systemd unit the packaged service runs as.
	journalUnit = "stem"

	// followInterval is how often a followed file is checked for growth.
	followInterval = 250 * time.Millisecond

	// tailBlock is how much of a file's end is read at a time when looking for
	// the last lines.
	tailBlock = 64 << 10
)

// logSource is where the service's log lives on this host.
type logSource struct {
	// file is the log file, empty when the log is the systemd journal.
	file string
}

// journalRunner runs journalctl with args, streaming to the given writers.
type journalRunner func(ctx context.Context, args []string, stdout, stderr io.Writer) error

// logEnv is everything `stem logs` takes from the host, so a test can supply
// each of them.
type logEnv struct {
	goos     string
	dataDir  string
	journal  journalRunner
	interval time.Duration
}

func hostLogEnv() logEnv {
	return logEnv{
		goos:     runtime.GOOS,
		dataDir:  daemonconn.ServiceDataDir(),
		journal:  runJournalctl,
		interval: followInterval,
	}
}

// resolveLogSource finds the log the packaged service writes. The daemon itself
// writes no log file: it logs to stdout, so what exists depends on who
// supervises it. launchd redirects stdout to <data dir>/logs/stem.log, and
// systemd sends it to the journal. Windows ships no service, so there is
// nothing to find and the operator names the file.
func resolveLogSource(goos, dataDir, file string) (logSource, error) {
	switch {
	case file != "":
		return logSource{file: file}, nil
	case goos == "darwin":
		return logSource{file: filepath.Join(dataDir, "logs", "stem.log")}, nil
	case goos == "linux":
		return logSource{}, nil
	default:
		return logSource{}, fmt.Errorf(
			"stem writes no log file on %s: it logs to the console of whatever runs it; name a captured log with --file",
			goos,
		)
	}
}

// logsCmd shows the service log, or follows it.
func logsCmd(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet(logsCommandName, flag.ContinueOnError)
	fs.SetOutput(stderr)
	follow := fs.Bool("follow", false, "Keep printing new lines as they are written")
	fs.BoolVar(follow, "f", false, "Keep printing new lines (shorthand)")
	lines := fs.Int("lines", defaultLogLines, "Number of recent lines to show")
	fs.IntVar(lines, "n", defaultLogLines, "Number of recent lines (shorthand)")
	file := fs.String("file", "", "Read this log file instead of the service's own")
	fs.Usage = func() {
		_, _ = fmt.Fprintln(stderr, `Usage: stem logs [-f] [-n N] [--file FILE]

Show the end of the service log.

The daemon writes its log to standard output, so where it ends up depends on
what runs it: on Linux the systemd journal (stem logs runs "journalctl -u
stem", which needs a member of the systemd-journal group or root); on macOS
<data directory>/logs/stem.log, the file the launchd service writes. On other
systems, and for a log captured some other way, name the file with --file.

Flags:`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	var err error
	switch {
	case fs.NArg() > 0:
		err = fmt.Errorf("unexpected argument %q", fs.Arg(0))
	case *lines < 1:
		err = fmt.Errorf("--lines must be at least 1, not %d", *lines)
	default:
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		err = runLogs(ctx, hostLogEnv(), *file, *lines, *follow, stdout, stderr)
	}
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "stem logs: %v\n", err)
	}
	return err
}

func runLogs(ctx context.Context, env logEnv, file string, lines int, follow bool, stdout, stderr io.Writer) error {
	source, err := resolveLogSource(env.goos, env.dataDir, file)
	if err != nil {
		return err
	}
	if source.file == "" {
		return showJournal(ctx, env, lines, follow, stdout, stderr)
	}
	err = showFile(ctx, env, source.file, lines, follow, stdout)
	if err != nil && file == "" {
		err = fmt.Errorf(
			"%w (the launchd service writes <data directory>/logs/stem.log; name another log with --file)",
			err,
		)
	}
	return err
}

func showJournal(ctx context.Context, env logEnv, lines int, follow bool, stdout, stderr io.Writer) error {
	args := []string{"--unit", journalUnit, "--lines", strconv.Itoa(lines), "--no-pager"}
	if follow {
		args = append(args, "--follow")
	}
	err := env.journal(ctx, args, stdout, stderr)
	switch {
	case errors.Is(err, exec.ErrNotFound):
		return errors.New(
			"journalctl is not installed: read the log with the tool that supervises stem, or name a file with --file",
		)
	case err != nil && ctx.Err() == nil:
		return err
	default:
		return nil
	}
}

func runJournalctl(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	command := exec.CommandContext(ctx, "journalctl", args...)
	command.Stdout = stdout
	command.Stderr = stderr
	return command.Run()
}

// openLog opens the log through an [os.Root] on its directory: the path is the
// operator's, and the base name cannot climb out of it.
func openLog(path string) (*os.File, error) {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()
	return root.Open(filepath.Base(path))
}

func showFile(ctx context.Context, env logEnv, path string, lines int, follow bool, stdout io.Writer) error {
	file, err := openLog(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	tail, end, err := lastLines(file, lines)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if _, err = stdout.Write(tail); err != nil {
		return err
	}
	if !follow {
		return nil
	}
	return followFile(ctx, env.interval, path, file, end, stdout)
}

// lastLines returns the last n lines of file and the offset they end at, which
// is where a follower carries on from.
func lastLines(file *os.File, n int) ([]byte, int64, error) {
	info, err := file.Stat()
	if err != nil {
		return nil, 0, err
	}
	end := info.Size()
	var tail []byte
	for offset := end; offset > 0 && bytes.Count(tail, []byte("\n")) <= n; {
		size := min(int64(tailBlock), offset)
		offset -= size
		block := make([]byte, size)
		if _, readErr := file.ReadAt(block, offset); readErr != nil {
			return nil, 0, readErr
		}
		tail = slices.Concat(block, tail)
	}
	// A trailing newline ends the last line rather than starting another.
	want := n
	if bytes.HasSuffix(tail, []byte("\n")) {
		want++
	}
	for found := bytes.Count(tail, []byte("\n")); found >= want; found-- {
		tail = tail[bytes.IndexByte(tail, '\n')+1:]
	}
	return tail, end, nil
}

// followFile prints what is appended to path after offset until ctx ends. The
// daemon's log rotates, so each pass checks that the path still names the file
// being read, and starts over on a new or shorter one.
func followFile(
	ctx context.Context,
	interval time.Duration,
	path string,
	file *os.File,
	offset int64,
	stdout io.Writer,
) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	current := file
	for {
		select {
		case <-ctx.Done():
			if current != file {
				_ = current.Close()
			}
			return nil
		case <-ticker.C:
		}
		next, reopened, err := rotated(current, path)
		if err != nil {
			return err
		}
		if reopened {
			if current != file {
				_ = current.Close()
			}
			current, offset = next, 0
		}
		offset, err = drain(current, offset, stdout)
		if err != nil {
			return err
		}
	}
}

// rotated reports whether path now names a different file than current, and
// opens it if so. A file that is briefly missing between rotation and the next
// write is not an error; the follower keeps the old one until it reappears.
func rotated(current *os.File, path string) (*os.File, bool, error) {
	onDisk, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return current, false, nil
		}
		return nil, false, err
	}
	held, err := current.Stat()
	if err != nil {
		return nil, false, err
	}
	if os.SameFile(held, onDisk) {
		return current, false, nil
	}
	next, err := openLog(path)
	if err != nil {
		return nil, false, err
	}
	return next, true, nil
}

// drain copies what has been appended to file since offset and returns the new
// offset. A file shorter than offset was truncated in place, so it is read
// again from the start.
func drain(file *os.File, offset int64, stdout io.Writer) (int64, error) {
	info, err := file.Stat()
	if err != nil {
		return offset, err
	}
	if info.Size() < offset {
		offset = 0
	}
	if info.Size() == offset {
		return offset, nil
	}
	copied, err := io.Copy(stdout, io.NewSectionReader(file, offset, info.Size()-offset))
	return offset + copied, err
}
