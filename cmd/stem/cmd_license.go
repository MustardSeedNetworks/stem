// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/MustardSeedNetworks/stem/internal/license"
)

// displayLicenseStatus displays the license status.
func displayLicenseStatus(mgr *license.Manager) {
	state := mgr.GetState()
	fp := mgr.GetFingerprint()

	_, _ = fmt.Fprintf(os.Stdout, "%s - License Status\n", ProductName)
	_, _ = fmt.Fprintln(os.Stdout, strings.Repeat("=", licenseBannerWidth))

	switch {
	case state == nil:
		_, _ = fmt.Fprintln(os.Stdout, "Status:    Not Activated")
		_, _ = fmt.Fprintln(os.Stdout, "\nTo start a 14-day trial:")
		_, _ = fmt.Fprintln(os.Stdout, "  stem license --trial")
		_, _ = fmt.Fprintln(os.Stdout, "\nTo activate with a license key:")
		_, _ = fmt.Fprintln(os.Stdout, "  stem license --activate XXXX-XXXX-XXXX-XXXX")
	case state.IsTrialMode:
		remaining := mgr.TrialDaysRemaining()
		_, _ = fmt.Fprintln(os.Stdout, "Status:    Trial Mode")
		_, _ = fmt.Fprintf(os.Stdout, "Days Left: %d\n", remaining)
		_, _ = fmt.Fprintf(os.Stdout, "Tier:      %s (full access during trial)\n", license.Tier(state.Tier))
		if remaining <= trialWarningDays {
			_, _ = fmt.Fprintln(os.Stdout, "\nWarning: Trial ending soon!")
			_, _ = fmt.Fprintln(os.Stdout, "Activate a license to continue using The Stem")
		}
	default:
		_, _ = fmt.Fprintln(os.Stdout, "Status:    Licensed")
		_, _ = fmt.Fprintf(os.Stdout, "Tier:      %s\n", license.Tier(state.Tier))
		_, _ = fmt.Fprintf(os.Stdout, "Key:       %s\n", license.FormatKey(state.LicenseKey))
		_, _ = fmt.Fprintf(os.Stdout, "Expires:   %s\n", state.ExpiresAt.Format("2006-01-02"))
	}

	_, _ = fmt.Fprintf(os.Stdout, "\nDevice ID: %s\n", fp.Hash())
	_, _ = fmt.Fprintf(os.Stdout, "Platform:  %s\n", fp.Platform)

	if state != nil && len(state.Features) > 0 {
		_, _ = fmt.Fprintf(os.Stdout, "\nEnabled Features:\n")
		for _, f := range state.Features {
			_, _ = fmt.Fprintf(os.Stdout, "  - %s\n", f)
		}
	}
}

func licenseCmd(args []string) {
	fs := flag.NewFlagSet("license", flag.ExitOnError)
	activate := fs.String("activate", "", "Activate with license key")
	trial := fs.Bool("trial", false, "Start 14-day trial")
	status := fs.Bool("status", false, "Show license status")
	deactivate := fs.Bool("deactivate", false, "Remove license")

	err := fs.Parse(args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	mgr, err := license.NewManager()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Error: Failed to initialize license manager: %v\n", err)
		os.Exit(1)
	}

	switch {
	case *activate != "":
		result := mgr.Activate(*activate)
		if result.Success {
			_, _ = fmt.Fprintf(os.Stdout, "Success: %s\n", result.Message)
			_, _ = fmt.Fprintf(os.Stdout, "Tier: %s\n", license.Tier(result.Tier))
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "Error: %s\n", result.Message)
			os.Exit(1)
		}

	case *trial:
		result := mgr.StartTrial()
		if result.Success {
			_, _ = fmt.Fprintf(os.Stdout, "Success: %s\n", result.Message)
			if result.DaysRemaining > 0 {
				_, _ = fmt.Fprintf(os.Stdout, "Days remaining: %d\n", result.DaysRemaining)
			}
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "Error: %s\n", result.Message)
			os.Exit(1)
		}

	case *deactivate:
		deactErr := mgr.Deactivate()
		if deactErr != nil {
			_, _ = fmt.Fprintf(os.Stdout, "Error: Failed to deactivate: %v\n", deactErr)
			os.Exit(1)
		}
		_, _ = fmt.Fprintln(os.Stdout, "License deactivated successfully")

	case *status:
		fallthrough
	default:
		displayLicenseStatus(mgr)
	}
}
