// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/MustardSeedNetworks/stem/internal/license"
	"github.com/MustardSeedNetworks/stem/internal/logging"
	reflectorConfig "github.com/MustardSeedNetworks/stem/internal/reflector/config"
	reflectorDP "github.com/MustardSeedNetworks/stem/internal/reflector/dataplane"
	reflectorTUI "github.com/MustardSeedNetworks/stem/internal/reflector/tui"
	testmasterTUI "github.com/MustardSeedNetworks/stem/internal/services/orchestrator/tui"
	"github.com/MustardSeedNetworks/stem/internal/version"
)

// tuiReflectMode runs the reflector TUI mode.
func tuiReflectMode(iface string) error {
	if iface == "" {
		_, _ = fmt.Fprintln(os.Stdout, "Error: --interface is required for reflect mode")
		_, _ = fmt.Fprintln(os.Stdout, "Usage: stem tui --mode reflect -i eth0")
		return errors.New("missing interface")
	}

	// Check license (Tier 1 minimum).
	mgr, err := license.NewManager()
	if err != nil {
		logging.Warn("license manager initialization failed", "error", err)
	}
	if mgr != nil && !mgr.IsActivated() {
		result := mgr.StartTrial()
		if !result.Success {
			_, _ = fmt.Fprintf(os.Stdout, "Error: %s\n", result.Message)
			return fmt.Errorf("license trial failed: %s", result.Message)
		}
	}

	// Build reflector config.
	cfg := &reflectorConfig.Config{
		Interface:       iface,
		Verbose:         false,
		SignatureFilter: DefaultSignatureFilter,
		WebUI:           reflectorConfig.WebUIConfig{Enabled: false, Port: 0},
		TUI:             reflectorConfig.TUIConfig{Enabled: true},
		Filtering: reflectorConfig.FilterConfig{
			Port:      0,
			FilterOUI: false,
			OUI:       "",
			FilterMAC: false,
		},
		Reflection: reflectorConfig.ReflectConfig{
			Mode: DefaultReflectionMode,
		},
		Platform: reflectorConfig.PlatformConfig{UseDPDK: false, UseAFXDP: false, DPDKArgs: ""},
		Stats:    reflectorConfig.StatsConfig{Format: "text", Interval: 0},
	}

	// Create and start reflector dataplane.
	dp, dpErr := reflectorDP.New(cfg)
	if dpErr != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Error: Failed to create reflector: %v\n", dpErr)
		return dpErr
	}
	defer dp.Close()

	tuiStartErr := dp.Start()
	if tuiStartErr != nil {
		dp.Close() // Cleanup before exit.
		_, _ = fmt.Fprintf(os.Stdout, "Error: Failed to start reflector: %v\n", tuiStartErr)
		return tuiStartErr
	}

	// Launch reflector TUI.
	tuiApp := reflectorTUI.New(dp)
	tuiRunErr := tuiApp.Run()
	if tuiRunErr != nil {
		_, _ = fmt.Fprintf(os.Stdout, "TUI error: %v\n", tuiRunErr)
		return tuiRunErr
	}

	return nil
}

// tuiTestMode runs the testmaster TUI mode.
func tuiTestMode() error {
	// Check license (Tier 2 required).
	mgr, mgrErr := license.NewManager()
	if mgrErr != nil {
		logging.Warn("license manager initialization failed", "error", mgrErr)
	}
	if mgr != nil {
		state := mgr.GetState()
		if state == nil {
			result := mgr.StartTrial()
			if !result.Success {
				_, _ = fmt.Fprintf(os.Stdout, "Error: %s\n", result.Message)
				return fmt.Errorf("license trial failed: %s", result.Message)
			}
		} else if license.Tier(state.Tier) < license.TierProfessional && !state.IsTrialMode {
			_, _ = fmt.Fprintln(os.Stdout, "Error: Professional TUI requires a Tier 2 (Professional) license")
			return errors.New("license tier too low")
		}
	}

	// Launch testmaster TUI.
	tuiApp := testmasterTUI.New()

	// Set up callbacks.
	tuiApp.OnQuit = func() {
		tuiApp.Stop()
	}

	tuiApp.Logf("The Stem TUI started")
	tuiApp.Logf("Press F1 to start test, F2 to stop, F10 to quit")

	tuiRunErr := tuiApp.Run()
	if tuiRunErr != nil {
		_, _ = fmt.Fprintf(os.Stdout, "TUI error: %v\n", tuiRunErr)
		return tuiRunErr
	}

	return nil
}

func tuiCmd(args []string) error {
	fs := flag.NewFlagSet("tui", flag.ExitOnError)
	mode := fs.String("mode", "test", "TUI mode: test or reflect")
	iface := fs.String("interface", "", "Network interface (required for reflect mode)")
	fs.StringVar(iface, "i", "", "Network interface (shorthand)")

	err := fs.Parse(args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "%s %s - Terminal UI\n", ProductName, version.GetVersion())

	switch *mode {
	case subReflect, "reflector":
		if modeErr := tuiReflectMode(*iface); modeErr != nil {
			return modeErr
		}
	case subTest, "testmaster", "":
		if modeErr := tuiTestMode(); modeErr != nil {
			return modeErr
		}
	default:
		_, _ = fmt.Fprintf(os.Stdout, "Error: Unknown TUI mode '%s'\n", *mode)
		_, _ = fmt.Fprintln(os.Stdout, "Valid modes: test, reflect")
		return fmt.Errorf("invalid TUI mode: %s", *mode)
	}

	return nil
}
