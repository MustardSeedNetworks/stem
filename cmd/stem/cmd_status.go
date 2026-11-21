// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/daemonclient"
	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
)

const statusCommandName = "status"

// Daemon states `stem status` reports.
const (
	stateRunning    = "running"
	stateNotReady   = "not_ready"
	stateNotRunning = "not_running"
	stateError      = "error"
)

var errNotReady = errors.New("the daemon is up but reports it is not ready")

// statusReport is what `stem status` prints, human or JSON. State is always
// set, so a script reading --json gets an object whatever happened.
type statusReport struct {
	State          string            `json:"state"`
	Running        bool              `json:"running"`
	URL            string            `json:"url,omitempty"`
	Version        string            `json:"version,omitempty"`
	Commit         string            `json:"commit,omitempty"`
	BuildTime      string            `json:"buildTime,omitempty"`
	UIBuildHash    string            `json:"uiBuildHash,omitempty"`
	TLSFingerprint string            `json:"tlsFingerprint,omitempty"`
	Ready          string            `json:"ready,omitempty"`
	Failing        map[string]string `json:"failing,omitempty"`
	Error          string            `json:"error,omitempty"`
}

// statusCmd reports whether the daemon is up and healthy, from what the daemon
// itself says at /__version and /health/ready. Neither needs a credential.
//
// Exit status: 0 healthy, 1 not running (nothing listening, or up but not
// ready), 2 the check could not be made.
func statusCmd(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet(statusCommandName, flag.ContinueOnError)
	fs.SetOutput(stderr)
	asJSON := fs.Bool("json", false, "Print the status as a JSON object")
	apiURL := fs.String("api", "",
		"Daemon base URL (default: the daemon on this host that published a descriptor)")
	caCert := fs.String("cacert", "",
		"Certificate to trust for the daemon (default: the one the daemon published)")
	fs.Usage = func() {
		_, _ = fmt.Fprintln(stderr, `Usage: stem status [--json] [--api URL] [--cacert FILE]

Report whether the stem daemon is up and ready. The daemon is found through
the descriptor it publishes in its data directory (STEM_DATA_DIR, else the
packaged location), and its certificate is trusted exactly as published; there
is no way to skip verification. For a daemon on another host, name it with the
flag --api and trust its certificate with the flag --cacert.

Exit status:
  0  the daemon is running and ready
  1  not running: nothing is listening, or the daemon is up but not ready
  2  the check could not be made (certificate, descriptor, usage, ...)

Flags:`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return withExitCode(statusError, err)
	}
	if fs.NArg() > 0 {
		err := fmt.Errorf("unexpected argument %q", fs.Arg(0))
		_, _ = fmt.Fprintf(stderr, "stem status: %v\n", err)
		return withExitCode(statusError, err)
	}

	report, code := checkStatus(context.Background(), *apiURL, *caCert)
	var err error
	if *asJSON {
		err = writeStatusJSON(stdout, report)
	} else {
		writeStatusText(stdout, stderr, report)
	}
	if err != nil {
		return err
	}
	if code == exitOK {
		return nil
	}
	return withExitCode(code, errors.New(report.Error))
}

// checkStatus probes the daemon and classifies the outcome. Only a failed
// connection is "not running": a certificate the caller cannot trust, a
// descriptor this account may not read or a daemon answering badly all mean
// something is there, and reporting them as a stopped daemon would send an
// operator to restart a service that is fine.
func checkStatus(ctx context.Context, apiURL, caCert string) (statusReport, int) {
	target, found, err := statusTarget(apiURL, caCert)
	switch {
	case err != nil:
		return statusReport{State: stateError, Error: err.Error()}, statusError
	case !found:
		return statusReport{
			State: stateNotRunning,
			Error: "no daemon has published a descriptor on this host (name one with --api and --cacert)",
		}, statusNotRunning
	}
	report := statusReport{URL: target.URL}

	health, err := daemonclient.Probe(ctx, target)
	switch {
	case errors.Is(err, daemonclient.ErrUnreachable):
		report.State = stateNotRunning
		report.Error = err.Error()
		return report, statusNotRunning
	case err != nil:
		report.State = stateError
		report.Error = err.Error()
		return report, statusError
	}

	report.Running = true
	report.Version = health.Build["version"]
	report.Commit = health.Build["commit"]
	report.BuildTime = health.Build["buildTime"]
	report.UIBuildHash = health.Build["uiBuildHash"]
	report.TLSFingerprint = health.Build["tlsFingerprint"]
	report.Ready = health.Ready.Status
	if !health.Healthy() {
		report.State = stateNotReady
		report.Failing = failingChecks(health.Ready.Checks)
		report.Error = errNotReady.Error()
		return report, statusNotRunning
	}
	report.State = stateRunning
	return report, exitOK
}

// statusTarget picks the daemon to ask: the one --api names, else the one this
// host's descriptor names. found is false when there is neither: a host whose
// daemon never published a descriptor has no daemon to ask, and guessing a port
// could reach a sibling product's. A descriptor that exists but cannot be used
// is an error, never "not running".
func statusTarget(apiURL, caCert string) (daemonclient.Target, bool, error) {
	if apiURL != "" {
		return daemonclient.Target{URL: apiURL, CAFile: caCert}, true, nil
	}
	descriptor, err := daemonconn.Discover()
	switch {
	case err == nil:
		target := daemonclient.Target{URL: descriptor.URL, CAFile: descriptor.CAFile}
		if caCert != "" {
			target.CAFile = caCert
		}
		return target, true, nil
	case daemonconn.IsNotFound(err):
		return daemonclient.Target{}, false, nil
	default:
		return daemonclient.Target{}, false, fmt.Errorf(
			"%w (run as the service account or root, or name the daemon with --api and --cacert)", err)
	}
}

func failingChecks(checks map[string]api.ReadinessCheck) map[string]string {
	failing := make(map[string]string)
	for name, check := range checks {
		if check.Status != "ok" {
			failing[name] = check.Error
		}
	}
	if len(failing) == 0 {
		return nil
	}
	return failing
}

func writeStatusJSON(w io.Writer, report statusReport) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// writeStatusText prints the report for a person. A daemon that is down is a
// finding and goes to stdout, where a script piping the output sees it; a check
// that could not be made is an error and goes to stderr.
func writeStatusText(stdout, stderr io.Writer, report statusReport) {
	switch report.State {
	case stateNotRunning:
		_, _ = fmt.Fprintf(stdout, "Status: not running\n%s\n", report.Error)
	case stateError:
		_, _ = fmt.Fprintf(stderr, "stem status: %s\n", report.Error)
	default:
		label := "running"
		if report.State == stateNotReady {
			label = "up, not ready"
		}
		_, _ = fmt.Fprintf(stdout, "Status:      %s\n", label)
		_, _ = fmt.Fprintf(stdout, "URL:         %s\n", report.URL)
		_, _ = fmt.Fprintf(
			stdout,
			"Version:     %s (commit %s, built %s)\n",
			report.Version,
			report.Commit,
			report.BuildTime,
		)
		_, _ = fmt.Fprintf(stdout, "Web UI:      %s\n", report.UIBuildHash)
		_, _ = fmt.Fprintf(stdout, "Certificate: %s\n", report.TLSFingerprint)
		for name, reason := range report.Failing {
			_, _ = fmt.Fprintf(stdout, "Failing:     %s: %s\n", name, reason)
		}
	}
}
