// SPDX-License-Identifier: BUSL-1.1

package main

// Observing a daemon-owned run (#1166). The daemon executes; the CLI starts
// the run, follows it, and reports what the daemon says — the same run ID,
// progress and terminal state the web UI shows, because it is the same run.

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/daemonclient"
)

// pollInterval is how often a watching CLI asks the daemon where it is.
const pollInterval = time.Second

// watchRun follows a run to its terminal state and returns the daemon's
// final view of it. Interrupting (Ctrl+C) asks the daemon to stop the run
// and keeps watching: the CLI must see a terminal state before it can
// honestly report one, so it never says "cancelled" for a run still holding
// the interface.
func watchRun(
	ctx context.Context,
	client *daemonclient.Client,
	runID string,
	quiet bool,
) (api.Stats, error) {
	// Polling outlives the interrupt on purpose — the run is the daemon's,
	// and abandoning it would leave the operator with no outcome at all.
	poll := context.WithoutCancel(ctx)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	interrupted := ctx.Done()
	for {
		status, err := client.Status(poll)
		if err != nil {
			return api.Stats{}, err
		}
		if status.SuiteID != "" && status.SuiteID != runID {
			return api.Stats{}, fmt.Errorf(
				"the daemon is now running %s, not %s; this run is no longer ours",
				status.SuiteID, runID)
		}
		if daemonclient.Terminal(status.TestStatus) {
			return status, nil
		}
		if !quiet {
			reportProgress(status)
		}

		select {
		case <-ticker.C:
		case <-interrupted:
			// Handled once; the channel stays closed, so drop it or the
			// select would spin instead of waiting for the next tick.
			interrupted = nil
			_, _ = fmt.Fprintln(os.Stdout, "\nCancelling run...")
			if stopErr := client.Stop(poll); stopErr != nil {
				return api.Stats{}, fmt.Errorf("ask the daemon to stop the run: %w", stopErr)
			}
		}
	}
}

// reportProgress prints one line per poll on a terminal.
func reportProgress(status api.Stats) {
	if status.StepsTotal == 0 {
		return
	}
	line := fmt.Sprintf("\r[%d/%d] %s", status.CurrentStep, status.StepsTotal, status.Phase)
	if status.ElapsedSeconds > 0 {
		line += fmt.Sprintf(" (%ds", status.ElapsedSeconds)
		if status.EstimatedRemainingSeconds != nil {
			line += fmt.Sprintf(", ~%ds left", *status.EstimatedRemainingSeconds)
		}
		line += ")"
	}
	_, _ = fmt.Fprint(os.Stdout, line)
}
