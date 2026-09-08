// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

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

	reportReflectorLicense()

	cfg := buildTUIReflectorConfig(iface)

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

func buildTUIReflectorConfig(iface string) *reflectorConfig.Config {
	return &reflectorConfig.Config{
		Interface:       iface,
		Verbose:         false,
		SignatureFilter: DefaultSignatureFilter,
		WebUI:           reflectorConfig.WebUIConfig{Enabled: false, Port: 0},
		TUI:             reflectorConfig.TUIConfig{Enabled: true},
		Filtering: reflectorConfig.FilterConfig{
			Port:      reflectorConfig.NetAllyPort,
			FilterOUI: false,
			OUI:       "",
			FilterMAC: false,
		},
		Reflection: reflectorConfig.ReflectConfig{
			Mode: DefaultReflectionMode,
		},
		Platform: reflectorConfig.PlatformConfig{UseAFXDP: true},
		Stats:    reflectorConfig.StatsConfig{Format: "text", Interval: 0},
	}
}

// tuiTestMode runs the testmaster TUI mode.
func tuiTestMode() error {
	// The TUI runs the same paid standards the CLI does, so it asks the same
	// question in the same place rather than keeping a second copy of it.
	if !checkTestLicense() {
		return errors.New("license check failed")
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
