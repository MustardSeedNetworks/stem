// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/daemonclient"
	reflectorConfig "github.com/MustardSeedNetworks/stem/internal/reflector/config"
	"github.com/MustardSeedNetworks/stem/internal/version"
)

func getReflectionMode(profile string) string {
	return reflectorConfig.SettingsForProfile(profile).Mode
}

func getReflectorPort(profile string, requested uint16) uint16 {
	if requested == 0 {
		return reflectorConfig.SettingsForProfile(profile).Port
	}
	return requested
}

// watchReflector prints the daemon's reflector counters until interrupted,
// then stops the daemon-owned reflector and prints its final totals.
func watchReflector(ctx context.Context, client *daemonclient.Client) error {
	// Polling outlives the interrupt: the reflector is the daemon's, and the
	// CLI has to read the final counters after asking it to stop.
	poll := context.WithoutCancel(ctx)

	var statsTick <-chan time.Time
	if stdout, err := os.Stdout.Stat(); err == nil && stdout.Mode()&os.ModeCharDevice != 0 {
		ticker := time.NewTicker(statsIntervalSeconds * time.Second)
		defer ticker.Stop()
		statsTick = ticker.C
	}

	for {
		select {
		case <-ctx.Done():
			_, _ = fmt.Fprintln(os.Stdout, "\nStopping reflector...")
			if err := client.Stop(poll); err != nil {
				return fmt.Errorf("ask the daemon to stop the reflector: %w", err)
			}
			stats, err := client.ReflectorStats(poll)
			if err != nil {
				return fmt.Errorf("read final reflector statistics: %w", err)
			}
			printReflectorTotals(stats)
			return nil
		case <-statsTick:
			stats, err := client.ReflectorStats(poll)
			if err != nil {
				return fmt.Errorf("read reflector statistics: %w", err)
			}
			printReflectorProgress(stats)
		}
	}
}

func printReflectorProgress(stats api.ReflectorStats) {
	_, _ = fmt.Fprintf(
		os.Stdout,
		"\r[Stats] RX: %d pkts | TX: %d pkts | RX bytes: %d | TX bytes: %d",
		stats.PacketsReceived, stats.PacketsReflected,
		stats.BytesReceived, stats.BytesReflected,
	)
}

func printReflectorTotals(stats api.ReflectorStats) {
	_, _ = fmt.Fprintf(os.Stdout, "\nFinal Statistics:\n")
	_, _ = fmt.Fprintf(os.Stdout, "  Packets Received:  %d\n", stats.PacketsReceived)
	_, _ = fmt.Fprintf(os.Stdout, "  Packets Reflected: %d\n", stats.PacketsReflected)
	_, _ = fmt.Fprintf(os.Stdout, "  Bytes Received:    %d\n", stats.BytesReceived)
	_, _ = fmt.Fprintf(os.Stdout, "  Bytes Reflected:   %d\n", stats.BytesReflected)
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

	// The reflector runs in the daemon, so it is one run the web UI can see
	// and stop, and one owner of the interface (#1166).
	client, err := daemonclient.Discover()
	if err != nil {
		return reportNoDaemon(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return runReflector(ctx, client, parsed)
}

// runReflector configures, starts and follows the daemon-owned reflector.
// It is separate from reflectCmd so the sequence can be driven in a test
// without flag parsing or a signal: configuring before starting is the whole
// point, since the daemon refuses a configuration change while running and
// --oui and --port would otherwise parse and do nothing.
func runReflector(ctx context.Context, client *daemonclient.Client, parsed *reflectCmdArgs) error {
	printReflectorStartup(parsed)

	port := getReflectorPort(parsed.profile, parsed.port)
	cfg := api.ReflectorConfig{
		Profile:    parsed.profile,
		OUIFilter:  parsed.oui,
		PortFilter: int(port),
	}
	if cfgErr := client.ConfigureReflector(ctx, cfg); cfgErr != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Error: %v\n", cfgErr)
		return cfgErr
	}

	runID, startErr := client.StartReflector(ctx, parsed.iface, parsed.profile, port)
	if startErr != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Error: %v\n", startErr)
		return startErr
	}

	_, _ = fmt.Fprintf(os.Stdout, "Run %s started on UDP %d. Press Ctrl+C to stop.\n", runID, port)
	return watchReflector(ctx, client)
}

type reflectCmdArgs struct {
	iface   string
	profile string
	oui     string
	port    uint16
}

func parseReflectFlags(args []string) (*reflectCmdArgs, *flag.FlagSet, error) {
	fs := flag.NewFlagSet(subReflect, flag.ExitOnError)
	iface := fs.String("interface", "", "Network interface")
	fs.StringVar(iface, "i", "", "Network interface (shorthand)")
	profile := fs.String("profile", DefaultProfile, "Preset profile")
	port := fs.Uint("port", 0, "UDP port filter")
	oui := fs.String("oui", "", "OUI filter")

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

// reportReflectorLicense tells the operator what the license state is before
// the reflector starts. Reflecting is the Free grant, so this never refuses
// and never starts a trial: spending the 14 Professional days to run a free
// capability was wrong even on a healthy install, and on a damaged one it
// overwrote the operator's license file (#1068).
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
