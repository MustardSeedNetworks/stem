// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/MustardSeedNetworks/foundation/pkg/release"

	"github.com/MustardSeedNetworks/stem/internal/daemonclient"
	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
	"github.com/MustardSeedNetworks/stem/internal/update"
	"github.com/MustardSeedNetworks/stem/internal/version"
)

const updateCommandName = "update"

// updateCmd installs the newest verified release in place, or runs or
// configures the automatic update (ADR-0012).
func updateCmd(args []string) error {
	fs := flag.NewFlagSet(updateCommandName, flag.ContinueOnError)
	source := fs.String("source", update.DefaultSource,
		`Where releases come from: "github", an https:// mirror, or an absolute directory`)
	automatic := fs.String("automatic", "",
		`Turn the daily automatic update "on" or "off"; --source sets where it updates from`)
	scheduled := fs.Bool("scheduled", false,
		"Run as the update timer does: only when automatic updates are on, from their saved source")
	fs.Usage = func() {
		_, _ = fmt.Fprintln(os.Stderr, `Usage: stem update [--source SOURCE]
       stem update --automatic on|off [--source SOURCE]
       stem update --scheduled

Install the newest stem release in place. Run it as root (Administrator on
Windows). It refuses while a test or reflector run is active.

Every release is verified offline against its signature, checksum and build
provenance before anything is touched; there is no way to skip a check. The
running version's release is verified too, so a failed update rolls back
without the network: the state directory is copied before the install and
restored with the previous version if the new one is not healthy within 60
seconds. Only a newer version is ever installed.

Automatic updates are off until an operator turns them on. A timer runs
"stem update --scheduled" daily; while they are off it stops before any
network call, and a timer that finds a run active waits for the next day.

Sources:
  github                  the GitHub releases of this product (default)
  https://mirror/stem     a mirror laid out as <base>/vX.Y.Z/<asset>, naming
                          its newest tag in <base>/latest
  /srv/stem-releases      an air-gapped directory with the same layout

Flags:`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })

	var err error
	switch {
	case fs.NArg() > 0:
		err = fmt.Errorf("unexpected argument %q", fs.Arg(0))
	case *scheduled && (set["source"] || set["automatic"]):
		err = errors.New("--scheduled takes no other flag: it updates from the saved source")
	case set["automatic"]:
		err = setAutomatic(*automatic, *source, set["source"])
	default:
		err = runUpdate(*source, *scheduled)
	}
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "stem update: %v\n", err)
	}
	return err
}

// setAutomatic saves the automatic update opt-in, and its source when one
// is given.
func setAutomatic(value, sourceSpec string, sourceSet bool) error {
	var on bool
	switch value {
	case "on":
		on = true
	case "off":
	default:
		return fmt.Errorf(`--automatic takes "on" or "off", not %q`, value)
	}
	if !update.Privileged() {
		return errors.New("run this as root (Administrator on Windows): the update timer runs as root")
	}
	if sourceSet {
		if _, err := release.ParseSource(sourceSpec, update.Policy().Repository, http.DefaultClient); err != nil {
			return err
		}
	}
	dataDir := daemonconn.ServiceDataDir()
	settings, err := update.LoadSettings(dataDir)
	if err != nil {
		return err
	}
	settings.Automatic = on
	if sourceSet {
		settings.Source = sourceSpec
	}
	if err = update.SaveSettings(dataDir, settings); err != nil {
		return err
	}
	if on {
		_, _ = fmt.Fprintf(os.Stdout, "Automatic updates are on, daily from %s.\n", settings.SourceSpec())
	} else {
		_, _ = fmt.Fprintln(os.Stdout, "Automatic updates are off.")
	}
	return nil
}

// runUpdate installs the newest release from sourceSpec. A scheduled run
// takes its source from the saved opt-in instead, and does nothing at all,
// network included, while automatic updates are off.
func runUpdate(sourceSpec string, scheduled bool) error {
	if !update.Privileged() {
		return errors.New(
			"run this as root (Administrator on Windows): it replaces the installed binary and restarts the service",
		)
	}
	log := slog.Default()
	dataDir := daemonconn.ServiceDataDir()
	if scheduled {
		settings, err := update.LoadSettings(dataDir)
		if err != nil {
			return err
		}
		if !settings.Automatic {
			log.Info("automatic updates are off", "event", "update.skipped", "reason", "off")
			return nil
		}
		sourceSpec = settings.SourceSpec()
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	exe, err := os.Executable()
	if err == nil {
		exe, err = filepath.EvalSymlinks(exe)
	}
	if err != nil {
		return fmt.Errorf("locate this binary: %w", err)
	}
	method, err := update.Detect(ctx, update.ExecRunner, runtime.GOOS, exe)
	if err != nil {
		return err
	}
	src, err := release.ParseSource(sourceSpec, update.Policy().Repository, http.DefaultClient)
	if err != nil {
		return err
	}
	verifier, err := release.NewVerifier(update.Policy())
	if err != nil {
		return err
	}
	health := daemonHealth(dataDir)
	if runtime.GOOS == "windows" {
		health = binaryHealth(exe)
	}

	updater := &update.Updater{
		Verifier: verifier,
		Source:   src,
		Method:   method,
		DataDir:  dataDir,
		Running:  version.GetVersion(),
		GOARCH:   runtime.GOARCH,
		Health:   health,
		Log:      log,
	}
	installed, err := updater.Run(ctx)
	switch {
	case errors.Is(err, update.ErrNotNewer):
		_, _ = fmt.Fprintf(os.Stdout, "Already up to date: %v\n", err)
		return nil
	case scheduled && errors.Is(err, update.ErrRunActive):
		log.Info("a run is active; the next scheduled check tries again",
			"event", "update.skipped", "reason", "run_active")
		return nil
	case err != nil:
		return err
	}
	_, _ = fmt.Fprintf(os.Stdout, "Updated to %s.\n", installed)
	return nil
}

// daemonHealth checks the restarted service through its own API, trusting
// the certificate its descriptor names: it must report the expected version
// with the web UI embedded.
func daemonHealth(dataDir string) func(context.Context, string) error {
	return func(ctx context.Context, want string) error {
		client, err := daemonclient.Open(dataDir)
		if err != nil {
			return err
		}
		build, err := client.Build(ctx)
		if err != nil {
			return err
		}
		if got := strings.TrimPrefix(build["version"], "v"); got != want {
			return fmt.Errorf("the service reports version %q", got)
		}
		if build["uiBuildHash"] == "" {
			return errors.New("the service reports no uiBuildHash: its web UI is not embedded")
		}
		return nil
	}
}

// binaryHealth runs the installed binary's version command. Windows ships no
// service, so the binary itself is all there is to check.
func binaryHealth(exe string) func(context.Context, string) error {
	return func(ctx context.Context, want string) error {
		out, err := update.ExecRunner(ctx, exe, "version")
		if err != nil {
			return err
		}
		first, _, _ := strings.Cut(string(out), "\n")
		if !strings.HasSuffix(strings.TrimSpace(first), " "+want) {
			return fmt.Errorf("the installed binary reports %q", strings.TrimSpace(first))
		}
		return nil
	}
}
