// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/version"
)

func webCmd(args []string) {
	fs := flag.NewFlagSet("web", flag.ExitOnError)
	port := fs.Int("port", defaultWebPort, "HTTPS port (1-65535)")
	fs.IntVar(port, "p", defaultWebPort, "HTTPS port (shorthand)")
	host := fs.String("host", "0.0.0.0", "Bind address")

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

	scheme := "https"

	_, _ = fmt.Fprintf(os.Stdout, "%s %s - WebUI Server\n", ProductName, version.GetVersion())
	_, _ = fmt.Fprintf(os.Stdout, "Starting on %s://%s:%d\n", scheme, *host, *port)

	srv, err := api.NewServer(*port)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		_, _ = fmt.Fprintf(
			os.Stderr,
			"Hint: Set STEM_AUTH_USERNAME and STEM_AUTH_PASSWORD environment variables\n",
		)
		os.Exit(1)
	}
	err = srv.Run()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
