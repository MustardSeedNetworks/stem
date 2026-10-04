// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/MustardSeedNetworks/foundation/pkg/instance"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/version"
)

func webCmd(args []string) {
	fs := flag.NewFlagSet("web", flag.ExitOnError)
	port := fs.Int("port", defaultWebPort, "HTTPS port (1-65535)")
	fs.IntVar(port, "p", defaultWebPort, "HTTPS port (shorthand)")
	host := fs.String("host", "", "Bind address (default: every address)")

	err := fs.Parse(args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Validate port range.
	if *port < 1 || *port > 65535 {
		_, _ = fmt.Fprintf(os.Stderr, "Error: port must be between 1 and 65535, got %d\n", *port)
		os.Exit(1)
	}

	// A port the operator named is the one they want: a busy one is refused
	// rather than silently swapped for a neighbour.
	explicitPort := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "port" || f.Name == "p" {
			explicitPort = true
		}
	})
	listen := api.ListenAddr{Host: *host, Port: *port, Explicit: explicitPort}

	_, _ = fmt.Fprintf(os.Stdout, "%s %s - WebUI Server\n", ProductName, version.GetVersion())
	_, _ = fmt.Fprintf(os.Stdout, "Starting on https://%s\n", net.JoinHostPort(*host, strconv.Itoa(*port)))

	srv, err := api.NewServer(listen)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	err = srv.Run()
	if err != nil {
		// A refused start is not a server error: another daemon already owns
		// this data directory, and the message already names which one.
		if held, isHeld := errors.AsType[*instance.HeldError](err); isHeld {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", held)
			os.Exit(1)
		}
		_, _ = fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
