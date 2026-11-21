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
	"time"

	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
	"github.com/MustardSeedNetworks/stem/internal/support"
)

const (
	supportBundleCommandName = "support-bundle"

	defaultBundleLogLines = 2000
)

// supportBundleCmd collects redacted diagnostics for support.
func supportBundleCmd(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet(supportBundleCommandName, flag.ContinueOnError)
	fs.SetOutput(stderr)
	output := fs.String("output", "", "Bundle to write (default: stem-support-<time>.tar.gz in the current directory)")
	fs.StringVar(output, "o", "", "Bundle to write (shorthand)")
	logFile := fs.String("log", "", "Log file to include instead of the service's own")
	lines := fs.Int("lines", defaultBundleLogLines, "Number of recent log lines to include")
	fs.Usage = func() {
		_, _ = fmt.Fprintln(stderr, `Usage: stem support-bundle [-o FILE] [--log FILE] [--lines N]

Collect what support asks for first: the build, the host, the daemon's status,
the redacted settings files, an inventory of the data directory and the end of
the log.

Every credential is removed before anything is written. Settings files are
redacted over their structure, the administrator credential file is reduced to
its field names, and the log is scrubbed of the credential values found in the
settings, this command's environment and the activation state, plus anything else shaped
like a token, a password hash, a licence key or a private key. Certificates are
listed by name and never opened.

Read the bundle before sending it. It is redacted, not anonymised: addresses,
interface names and the contents of the log are still in it.

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
		err = fmt.Errorf("unexpected argument %q; name the bundle with -o", fs.Arg(0))
	case *lines < 1:
		err = fmt.Errorf("--lines must be at least 1, not %d", *lines)
	default:
		err = runSupportBundle(*output, *logFile, *lines, stdout, stderr)
	}
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "stem support-bundle: %v\n", err)
	}
	return err
}

func runSupportBundle(output, logFile string, lines int, stdout, stderr io.Writer) error {
	now := time.Now()
	if output == "" {
		output = defaultArchiveName("stem-support", now)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	options := support.BundleOptions{
		DataDir:     daemonconn.ServiceDataDir(),
		Environment: os.Environ(),
		Now:         now,
	}
	report, _ := checkStatus(ctx, "", "")
	var statusJSON bytes.Buffer
	if err := writeStatusJSON(&statusJSON, report); err == nil {
		options.Status = statusJSON.Bytes()
	}
	// A log that cannot be read is worth saying and not worth failing for: the
	// rest of the bundle is still what support needs first.
	log, err := collectLog(ctx, hostLogEnv(), logFile, lines)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "stem support-bundle: the log is not included: %v\n", err)
	}
	options.Log = log

	if err = writeArchive(output, func(file *os.File) error {
		return support.WriteBundle(options, file)
	}); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "Support bundle written to %s\n", output)
	return nil
}

// collectLog reads the end of the log the same way `stem logs` finds it.
func collectLog(ctx context.Context, env logEnv, file string, lines int) ([]byte, error) {
	var log bytes.Buffer
	if err := runLogs(ctx, env, file, lines, false, &log, io.Discard); err != nil {
		return nil, err
	}
	return log.Bytes(), nil
}
