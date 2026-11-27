// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

const serviceCommandName = "service"

// isServiceAction reports whether action is a verb of `stem service`. "run" is
// what the Windows service manager invokes; operators use the others.
func isServiceAction(action string) bool {
	switch action {
	case "install", "start", "stop", "status", "uninstall", "run":
		return true
	default:
		return false
	}
}

var errServiceWindowsOnly = errors.New("stem service is Windows only: on Linux use systemd, on macOS launchd")

// serviceCmd manages the Windows service that runs `stem web`.
func serviceCmd(args []string) error {
	fs := flag.NewFlagSet(serviceCommandName, flag.ContinueOnError)
	fs.Usage = func() {
		_, _ = fmt.Fprintln(os.Stderr, `Usage: stem service install|start|stop|status|uninstall

Manage the Stem Windows service, which runs the web daemon (stem web) on
port 8444 at boot with its state under %ProgramData%\stem. Run it from an
elevated prompt.

  install    register the service, or update an existing registration
  start      start it and wait until it is running
  stop       stop it and wait until it has stopped
  status     show whether it is running
  uninstall  stop and remove the service; configuration and data stay

Windows only.`)
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	var err error
	switch {
	case fs.NArg() != 1:
		err = errors.New("expected one action: install, start, stop, status or uninstall")
	case !isServiceAction(fs.Arg(0)):
		err = fmt.Errorf("unknown action %q", fs.Arg(0))
	default:
		err = runServiceAction(fs.Arg(0), os.Stdout)
	}
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	return err
}
