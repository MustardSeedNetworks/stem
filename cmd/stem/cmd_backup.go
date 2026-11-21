// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
	"github.com/MustardSeedNetworks/stem/internal/support"
)

const (
	backupCommandName  = "backup"
	restoreCommandName = "restore"
)

// backupCmd archives the daemon's state.
func backupCmd(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet(backupCommandName, flag.ContinueOnError)
	fs.SetOutput(stderr)
	output := fs.String("output", "", "Archive to write (default: stem-backup-<time>.tar.gz in the current directory)")
	fs.StringVar(output, "o", "", "Archive to write (shorthand)")
	fs.Usage = func() {
		_, _ = fmt.Fprintln(stderr, `Usage: stem backup [-o FILE]

Archive the daemon's state: the administrator credential, the reflector and
update settings, the TLS certificate and key, the activation state and the
published descriptor. The data directory is STEM_DATA_DIR, else the packaged
location.

Stop the daemon first: a copy taken under a running daemon is not a backup, so
the command takes the lock the daemon holds and refuses while it is held. The
TLS pair is read from <data directory>/certs, where the packaged services keep
it; a daemon started from another working directory keeps its own.

The archive holds credentials and a private key and is written owner-only. It
is never overwritten, and the instance lock and the first-run and recovery
tokens are not in it. Two backups of unchanged state are byte-identical.

Flags:`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	err := runBackup(*output, fs.NArg(), stdout)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "stem backup: %v\n", err)
	}
	return err
}

func runBackup(output string, extraArgs int, stdout io.Writer) error {
	if extraArgs > 0 {
		return errors.New("backup takes no arguments; name the archive with -o")
	}
	if output == "" {
		output = defaultArchiveName("stem-backup", time.Now())
	}
	dataDir := daemonconn.ServiceDataDir()
	if err := writeArchive(output, func(file *os.File) error {
		return support.Backup(dataDir, file)
	}); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "State in %s backed up to %s\n", dataDir, output)
	return nil
}

// restoreCmd replaces the daemon's state with an archive written by backup.
func restoreCmd(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet(restoreCommandName, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		_, _ = fmt.Fprintln(stderr, `Usage: stem restore ARCHIVE

Replace the daemon's state with an archive written by "stem backup". Run it as
root, or as the service account, with the daemon stopped.

The restore is refused while the daemon is running: it takes the same lock the
daemon holds for its whole life, and keeps it until it is done. Before anything
is replaced the current state is copied beside the data directory
(<data directory>.pre-restore-<time>), and the archive is checked in full first,
so a refused archive changes nothing. An entry that is not part of a stem
backup, or that names a path outside the data directory, refuses the archive.

Restored files take the data directory's owner and are readable by it alone.

Flags:`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	err := runRestore(fs.Args(), stdout)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "stem restore: %v\n", err)
	}
	return err
}

func runRestore(args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return errors.New("restore takes exactly one argument: the archive")
	}
	archive, err := openArchive(args[0])
	if err != nil {
		return err
	}
	defer func() { _ = archive.Close() }()

	dataDir := daemonconn.ServiceDataDir()
	previous, err := support.Restore(archive, dataDir, time.Now())
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(
		stdout,
		"State restored from %s into %s\nThe previous state is kept at %s\n",
		args[0],
		dataDir,
		previous,
	)
	return nil
}
