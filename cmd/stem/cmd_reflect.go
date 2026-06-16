// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/license"
	reflectorConfig "github.com/MustardSeedNetworks/stem/internal/reflector/config"
	reflectorDP "github.com/MustardSeedNetworks/stem/internal/reflector/dataplane"
	reflectorTUI "github.com/MustardSeedNetworks/stem/internal/reflector/tui"
	"github.com/MustardSeedNetworks/stem/internal/version"
)

// getSignatureFilter maps profile name to signature filter.
func getSignatureFilter(profile string) string {
	switch profile {
	case "netally", "ito":
		return "ito"
	case "msn":
		return "msn"
	case "custom":
		return "custom"
	default:
		return DefaultSignatureFilter
	}
}

// reflectorStatsLoop displays reflector stats periodically until interrupted.
func reflectorStatsLoop(dp *reflectorDP.Dataplane) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(statsIntervalSeconds * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			_, _ = fmt.Fprintln(os.Stdout, "\nShutting down reflector...")
			go dp.Stop()
			time.Sleep(shutdownDelayMs * time.Millisecond)
			stats := dp.GetStats()
			_, _ = fmt.Fprintf(os.Stdout, "\nFinal Statistics:\n")
			_, _ = fmt.Fprintf(os.Stdout, "  Packets Received:  %d\n", stats.PacketsReceived)
			_, _ = fmt.Fprintf(os.Stdout, "  Packets Reflected: %d\n", stats.PacketsReflected)
			_, _ = fmt.Fprintf(os.Stdout, "  Bytes Received:    %d\n", stats.BytesReceived)
			_, _ = fmt.Fprintf(os.Stdout, "  Bytes Reflected:   %d\n", stats.BytesReflected)
			return
		case <-ticker.C:
			stats := dp.GetStats()
			_, _ = fmt.Fprintf(
				os.Stdout,
				"\r[Stats] RX: %d pkts | TX: %d pkts | Signatures: ITO=%d RFC2544=%d Y.1564=%d MSN=%d",
				stats.PacketsReceived,
				stats.PacketsReflected,
				stats.SigProbeOT+stats.SigDataOT+stats.SigLatency,
				stats.SigRFC2544,
				stats.SigY1564,
				stats.SigMSN,
			)
		}
	}
}

func reflectCmd(args []string) error {
	parsed, fs, parseErr := parseReflectFlags(args)
	if parseErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", parseErr)
		return parseErr
	}

	if err := requireReflectInterface(parsed.iface, fs); err != nil {
		return err
	}

	if err := checkReflectorLicense(); err != nil {
		return err
	}

	sigFilter := getSignatureFilter(parsed.profile)

	cfg := buildReflectorConfig(parsed, sigFilter)

	// Create reflector dataplane.
	dp, err := reflectorDP.New(cfg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Error: Failed to create reflector: %v\n", err)
		return err
	}
	defer dp.Close()

	// Start reflector.
	startErr := dp.Start()
	if startErr != nil {
		dp.Close() // Cleanup before exit.
		_, _ = fmt.Fprintf(os.Stdout, "Error: Failed to start reflector: %v\n", startErr)
		return startErr
	}

	printReflectorStartup(parsed)
	_, _ = fmt.Fprintln(os.Stdout, "\nReflector started. Press Ctrl+C to stop.")

	if parsed.useTUI {
		tuiApp := reflectorTUI.New(dp)
		tuiErr := tuiApp.Run()
		if tuiErr != nil {
			_, _ = fmt.Fprintf(os.Stdout, "TUI error: %v\n", tuiErr)
		}
	} else {
		reflectorStatsLoop(dp)
	}

	return nil
}

type reflectCmdArgs struct {
	iface   string
	profile string
	oui     string
	port    uint16
	useTUI  bool
}

func parseReflectFlags(args []string) (*reflectCmdArgs, *flag.FlagSet, error) {
	fs := flag.NewFlagSet(subReflect, flag.ExitOnError)
	iface := fs.String("interface", "", "Network interface")
	fs.StringVar(iface, "i", "", "Network interface (shorthand)")
	profile := fs.String("profile", DefaultProfile, "Preset profile")
	port := fs.Uint("port", 0, "UDP port filter")
	oui := fs.String("oui", "", "OUI filter")
	useTUI := fs.Bool("tui", false, "Launch TUI dashboard")

	if err := fs.Parse(args); err != nil {
		return nil, fs, err
	}

	if *port > math.MaxUint16 {
		return nil, fs, fmt.Errorf("port %d out of valid range (0-%d)", *port, math.MaxUint16)
	}

	return &reflectCmdArgs{
		iface:   *iface,
		profile: *profile,
		port:    uint16(*port),
		oui:     *oui,
		useTUI:  *useTUI,
	}, fs, nil
}

func requireReflectInterface(iface string, fs *flag.FlagSet) error {
	if iface == "" {
		_, _ = fmt.Fprintln(os.Stdout, "Error: --interface is required")
		fs.Usage()
		return errors.New("missing interface")
	}
	return nil
}

func checkReflectorLicense() error {
	// Check license (Tier 1 minimum).
	mgr, err := license.NewManager()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Warning: License check failed: %v\n", err)
		return nil
	}
	if mgr.IsActivated() {
		return nil
	}

	_, _ = fmt.Fprintln(os.Stdout, "No active license. Starting 14-day trial...")
	result := mgr.StartTrial()
	if !result.Success {
		_, _ = fmt.Fprintf(os.Stdout, "Error: %s\n", result.Message)
		return fmt.Errorf("license trial failed: %s", result.Message)
	}
	_, _ = fmt.Fprintf(os.Stdout, "%s\n", result.Message)
	return nil
}

func buildReflectorConfig(parsed *reflectCmdArgs, sigFilter string) *reflectorConfig.Config {
	cfg := &reflectorConfig.Config{
		Interface:       parsed.iface,
		Verbose:         false,
		SignatureFilter: sigFilter,
		WebUI:           reflectorConfig.WebUIConfig{Enabled: false, Port: 0},
		TUI:             reflectorConfig.TUIConfig{Enabled: parsed.useTUI},
		Filtering: reflectorConfig.FilterConfig{
			Port:      parsed.port,
			FilterOUI: false,
			OUI:       "00:c0:17", // Default NetAlly OUI.
			FilterMAC: false,
		},
		Reflection: reflectorConfig.ReflectConfig{
			Mode: DefaultReflectionMode,
		},
		Platform: reflectorConfig.PlatformConfig{UseDPDK: false, UseAFXDP: false, DPDKArgs: ""},
		Stats:    reflectorConfig.StatsConfig{Format: "text", Interval: 0},
	}

	if parsed.oui != "" {
		cfg.Filtering.FilterOUI = true
		cfg.Filtering.OUI = parsed.oui
	}

	return cfg
}

func printReflectorStartup(parsed *reflectCmdArgs) {
	_, _ = fmt.Fprintf(os.Stdout, "%s %s - Reflector\n", ProductName, version.GetVersion())
	_, _ = fmt.Fprintf(os.Stdout, "Interface:  %s\n", parsed.iface)
	_, _ = fmt.Fprintf(os.Stdout, "Profile:    %s\n", parsed.profile)
	if parsed.port > 0 {
		_, _ = fmt.Fprintf(os.Stdout, "Port:       %d\n", parsed.port)
	}
	if parsed.oui != "" {
		_, _ = fmt.Fprintf(os.Stdout, "OUI:        %s\n", parsed.oui)
	}
}
