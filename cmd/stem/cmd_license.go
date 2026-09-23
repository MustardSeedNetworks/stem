// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	fndlicense "github.com/MustardSeedNetworks/foundation/pkg/license"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/daemonclient"
	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
	"github.com/MustardSeedNetworks/stem/internal/license"
)

// licenseCmdFlags is what the operator asked for, decided once so the daemon
// and offline paths answer the same question.
type licenseCmdFlags struct {
	activate   string
	trial      bool
	deactivate bool
}

// displayLicenseStatus renders the entitlement state. It takes the wire type
// rather than a manager so the daemon's answer and a daemon-less host's own
// answer are rendered by one piece of code (#1335).
func displayLicenseStatus(status api.LicenseStatus) {
	_, _ = fmt.Fprintf(os.Stdout, "%s - License Status\n", ProductName)
	_, _ = fmt.Fprintln(os.Stdout, strings.Repeat("=", licenseBannerWidth))

	switch {
	case status.IsTrialMode && status.DaysRemaining > 0:
		_, _ = fmt.Fprintln(os.Stdout, "Status:    Trial Mode")
		_, _ = fmt.Fprintf(os.Stdout, "Days Left: %d\n", status.DaysRemaining)
		_, _ = fmt.Fprintf(os.Stdout, "Tier:      %s (full access during trial)\n", license.Tier(status.Tier))
		if status.DaysRemaining <= trialWarningDays {
			_, _ = fmt.Fprintln(os.Stdout, "\nWarning: Trial ending soon!")
			_, _ = fmt.Fprintf(os.Stdout, "Activate a license to continue using %s\n", ProductName)
		}
	case status.IsTrialMode:
		_, _ = fmt.Fprintln(os.Stdout, "Status:    Trial Expired")
		_, _ = fmt.Fprintf(os.Stdout, "Tier:      %s\n", license.Tier(status.Tier))
		_, _ = fmt.Fprintln(os.Stdout, "\nActivate a Pro key with:")
		_, _ = fmt.Fprintln(os.Stdout, "  stem license --activate MSN1.<payload>.<signature>")
	case status.Activated:
		_, _ = fmt.Fprintln(os.Stdout, "Status:    Licensed")
		_, _ = fmt.Fprintf(os.Stdout, "Tier:      %s\n", status.TierName)
		_, _ = fmt.Fprintf(os.Stdout, "Key:       %s\n", status.LicenseKey)
		_, _ = fmt.Fprintf(os.Stdout, "Expires:   %s\n", status.ExpiresAt.Format("2006-01-02"))
	case status.LicenseKey != "":
		// An activation is on disk but is no longer in force. Naming the key
		// and the date is what the operator renews against; offering a trial
		// here would be advice the daemon refuses on an expired paid key.
		_, _ = fmt.Fprintln(os.Stdout, "Status:    Expired")
		_, _ = fmt.Fprintf(os.Stdout, "Tier:      %s\n", status.TierName)
		_, _ = fmt.Fprintf(os.Stdout, "Key:       %s\n", status.LicenseKey)
		_, _ = fmt.Fprintf(os.Stdout, "Expires:   %s\n", status.ExpiresAt.Format("2006-01-02"))
		_, _ = fmt.Fprintln(os.Stdout, "\nActivate a current key with:")
		_, _ = fmt.Fprintln(os.Stdout, "  stem license --activate MSN1.<payload>.<signature>")
	default:
		_, _ = fmt.Fprintln(os.Stdout, "Status:    Not Activated")
		_, _ = fmt.Fprintln(os.Stdout, "\nTo start a 14-day trial:")
		_, _ = fmt.Fprintln(os.Stdout, "  stem license --trial")
		_, _ = fmt.Fprintln(os.Stdout, "\nTo activate with a license key:")
		_, _ = fmt.Fprintln(os.Stdout, "  stem license --activate MSN1.<payload>.<signature>")
	}

	_, _ = fmt.Fprintf(os.Stdout, "\nDevice ID: %s\n", status.DeviceHash)
	_, _ = fmt.Fprintf(os.Stdout, "Platform:  %s\n", status.Platform)

	if len(status.Features) > 0 {
		_, _ = fmt.Fprintf(os.Stdout, "\nEnabled Features:\n")
		for _, f := range status.Features {
			_, _ = fmt.Fprintf(os.Stdout, "  - %s\n", f)
		}
	}
}

func licenseCmd(args []string) error {
	fs := flag.NewFlagSet("license", flag.ExitOnError)
	activate := fs.String("activate", "", "Activate with license key")
	trial := fs.Bool("trial", false, "Start 14-day trial")
	// --status is the default; the flag exists so the intent can be typed.
	_ = fs.Bool("status", false, "Show license status")
	deactivate := fs.Bool("deactivate", false, "Remove license")

	if err := fs.Parse(args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}
	flags := licenseCmdFlags{activate: *activate, trial: *trial, deactivate: *deactivate}

	// The daemon holds the licence manager while it runs, and its copy is the
	// one in force: it answers /api/v1/license from memory, so a file this
	// command wrote would be invisible until a restart and the two would race
	// on the file itself. Whoever owns the licence answers about it (#1335).
	client, err := daemonclient.Discover()
	switch {
	case err == nil:
		return licenseViaDaemon(context.Background(), client, flags)
	case daemonconn.IsNotFound(err):
		return licenseOffline(flags)
	default:
		return reportUnusableDaemon(err)
	}
}

// licenseViaDaemon asks the running daemon, which owns the entitlement.
func licenseViaDaemon(ctx context.Context, client *daemonclient.Client, flags licenseCmdFlags) error {
	switch {
	case flags.activate != "":
		result, err := client.ActivateLicense(ctx, flags.activate)
		if err != nil {
			return reportDaemonError(err)
		}
		return reportActivation(result)

	case flags.trial:
		result, err := client.StartTrial(ctx)
		if err != nil {
			return reportDaemonError(err)
		}
		return reportActivation(result)

	case flags.deactivate:
		result, err := client.DeactivateLicense(ctx)
		if err != nil {
			return reportDaemonError(err)
		}
		if !result.Success {
			_, _ = fmt.Fprintf(os.Stdout, "Error: %s\n", result.Message)
			return errors.New(result.Message)
		}
		_, _ = fmt.Fprintln(os.Stdout, result.Message)
		return nil

	default:
		status, err := client.License(ctx)
		if err != nil {
			return reportDaemonError(err)
		}
		displayLicenseStatus(status)
		return nil
	}
}

// licenseOffline is the no-daemon host: this process is the only writer, so
// it may write the file itself. There is deliberately no path here for a
// daemon that was found but could not be reached — writing under a live
// daemon is the defect, not the fallback.
func licenseOffline(flags licenseCmdFlags) error {
	mgr, err := license.Load()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Error: Failed to initialize license manager: %v\n", err)
		return err
	}
	if loadStatus := mgr.LoadStatus(); !loadStatus.Usable() {
		// The operator asked about licensing, so say what is on disk. The
		// subcommands still run: --activate and --trial are how a damaged
		// file is replaced.
		_, _ = fmt.Fprintf(os.Stdout, "Warning: license file %s is %s; this host is entitled to Free features only\n",
			license.DefaultLicensePath(), loadStatus)
	}

	switch {
	case flags.activate != "":
		return reportActivation(*mgr.Activate(flags.activate))

	case flags.trial:
		return reportActivation(*mgr.StartTrial())

	case flags.deactivate:
		if deactErr := mgr.Deactivate(); deactErr != nil {
			_, _ = fmt.Fprintf(os.Stdout, "Error: Failed to deactivate: %v\n", deactErr)
			return deactErr
		}
		_, _ = fmt.Fprintln(os.Stdout, "License deactivated successfully")
		return nil

	default:
		displayLicenseStatus(api.LicenseStatusOf(mgr))
		return nil
	}
}

// reportActivation prints whoever decided the activation, verbatim. The CLI
// does not second-guess the verdict: with a daemon running it is not the
// CLI's manager that would have to honour the key.
func reportActivation(result fndlicense.ActivationResult) error {
	if !result.Success {
		_, _ = fmt.Fprintf(os.Stdout, "Error: %s\n", result.Message)
		return errors.New(result.Message)
	}
	_, _ = fmt.Fprintf(os.Stdout, "Success: %s\n", result.Message)
	if result.Tier != 0 {
		_, _ = fmt.Fprintf(os.Stdout, "Tier: %s\n", license.Tier(result.Tier))
	}
	if result.DaysRemaining > 0 {
		_, _ = fmt.Fprintf(os.Stdout, "Days remaining: %d\n", result.DaysRemaining)
	}
	return nil
}

// reportUnusableDaemon covers a descriptor that is present but unusable. It
// is an error rather than a fallback: the daemon is running, so a local write
// would be exactly the race #1335 removes.
func reportUnusableDaemon(err error) error {
	_, _ = fmt.Fprintf(os.Stdout, "Error: %v\n", err)
	switch {
	case errors.Is(err, daemonconn.ErrUnreadable):
		_, _ = fmt.Fprintln(os.Stdout,
			"The daemon's credential is readable only by the account it runs as. "+
				"Run this as root (sudo stem license ...) or as the stem user.")
	case errors.Is(err, daemonconn.ErrPermissions):
		_, _ = fmt.Fprintln(os.Stdout,
			"Another account can read the daemon's credential, so it will not be used. "+
				"Restore it with 'chmod 600' and restart the daemon to reissue the token.")
	default:
		_, _ = fmt.Fprintf(os.Stdout,
			"A Stem daemon published a descriptor that cannot be used. Looked in: %s\n",
			strings.Join(daemonconn.SearchDirs(), ", "))
	}
	return err
}

// reportDaemonError explains a daemon that answered with a refusal. The
// licence stays whatever the daemon says it is.
func reportDaemonError(err error) error {
	_, _ = fmt.Fprintf(os.Stdout, "Error: %v\n", err)
	return err
}
