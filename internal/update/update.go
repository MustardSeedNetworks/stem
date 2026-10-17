// SPDX-License-Identifier: BUSL-1.1

package update

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/MustardSeedNetworks/foundation/pkg/release"
)

// Policy is stem's release identity: every bundle must be signed by this
// repository's release workflow at the target tag.
func Policy() release.Policy {
	return release.Policy{
		Repository:      "MustardSeedNetworks/stem",
		Workflow:        ".github/workflows/release.yml",
		ProvenanceAsset: "stem-slsa-provenance.intoto.jsonl",
	}
}

// HealthTimeout bounds how long the new version has to come up healthy.
const HealthTimeout = 60 * time.Second

const healthPoll = time.Second

// ErrNotNewer reports that the source offers nothing newer than the running
// version. Downgrades are refused: any older signed release verifies, so a
// source could otherwise roll a site back.
var ErrNotNewer = errors.New("no newer release is available")

// Fetcher downloads a release asset and every file that proves it, and
// returns its path only once it verified. *release.Verifier is the real one.
type Fetcher interface {
	Fetch(ctx context.Context, src release.Source, tag, asset, dir string) (string, error)
}

// Updater runs one in-place update (ADR-0012).
type Updater struct {
	Verifier Fetcher
	Source   release.Source
	Method   Method
	// DataDir is the daemon's state directory, which holds the run lock.
	DataDir string
	// Running is the running version, with or without the leading v.
	Running string
	GOARCH  string
	// Health returns nil once the installed version reports version
	// (without the leading v) and is fully up.
	Health func(ctx context.Context, version string) error
	// HealthTimeout overrides the package default, for tests.
	HealthTimeout time.Duration
	Log           *slog.Logger
}

// Run updates to the newest release the source offers and returns its
// version. On any failure after the service stopped, it restores the state
// copy and the previous version before returning the error.
func (u *Updater) Run(ctx context.Context) (string, error) {
	unlock, err := NewRunGuard(u.DataDir).Acquire()
	if errors.Is(err, ErrRunLockHeld) {
		return "", errors.New("a test or reflector run is active; update after it finishes")
	}
	if err != nil {
		return "", err
	}
	defer unlock()
	// The update runs as root and may have just created the lock; the daemon
	// opens it as its own account.
	if err = shareWithDataDir(filepath.Join(u.DataDir, RunLockName), u.DataDir); err != nil {
		return "", err
	}

	from := strings.TrimPrefix(u.Running, "v")
	fromParts, ok := parseVersion(from)
	if !ok {
		return "", fmt.Errorf(
			"%w: the running version %q is not a release, so there is nothing to roll back to",
			ErrUnsupportedInstall,
			u.Running,
		)
	}
	latest, err := u.Source.Latest(ctx)
	if err != nil {
		return "", fmt.Errorf("find the latest release: %w", err)
	}
	to := strings.TrimPrefix(latest, "v")
	if toParts, valid := parseVersion(to); !valid || !newer(toParts, fromParts) {
		return "", fmt.Errorf("%w: the source's latest is %s and this is %s", ErrNotNewer, latest, from)
	}
	log := u.Log.With("from", from, "to", to, "method", u.Method.Name())

	work, err := os.MkdirTemp(filepath.Dir(u.DataDir), ".stem-update-")
	if err != nil {
		return "", err
	}
	artifacts, err := u.fetch(ctx, work, to, from)
	if err != nil {
		return "", errors.Join(err, os.RemoveAll(work))
	}
	next, previous := artifacts[0], artifacts[1]
	log.InfoContext(ctx, "update verified", "event", "update.verified")

	// Past this point an interrupt must not kill a package manager halfway
	// through: the update runs to success or rollback, bounded by the health
	// deadline.
	ctx = context.WithoutCancel(ctx)

	if err = u.Method.Stop(ctx); err != nil {
		return "", errors.Join(fmt.Errorf("stop the service: %w", err), os.RemoveAll(work))
	}
	state := filepath.Join(work, "state")
	if err = copyState(u.DataDir, state); err != nil {
		return "", errors.Join(
			fmt.Errorf("copy the state directory: %w", err), os.RemoveAll(work), u.Method.Start(ctx))
	}
	log.InfoContext(ctx, "state copied", "event", "update.state_copied", "copy", state)

	if err = u.install(ctx, u.Method.Apply, next, to); err != nil {
		log.ErrorContext(ctx, "update failed; rolling back", "event", "update.failed", "error", err)
		if rbErr := u.rollback(ctx, previous, state, from); rbErr != nil {
			log.ErrorContext(ctx, "rollback failed",
				"event", "update.rollback_failed", "error", rbErr, "state_copy", state)
			return "", fmt.Errorf(
				"update to %s failed (%w) and so did the rollback to %s (%w); the state copy is kept at %s",
				to,
				err,
				from,
				rbErr,
				state,
			)
		}
		log.InfoContext(ctx, "rolled back", "event", "update.rolled_back")
		return "", errors.Join(
			fmt.Errorf("update to %s failed and %s was restored: %w", to, from, err), os.RemoveAll(work))
	}
	log.InfoContext(ctx, "update installed", "event", "update.installed")
	return to, os.RemoveAll(work)
}

// fetch downloads and verifies the target release and the running one, and
// returns their artifacts in that order. The running version's artifact is
// the rollback: without it verified nothing is touched, so a rollback never
// needs the network.
func (u *Updater) fetch(ctx context.Context, work string, versions ...string) ([]string, error) {
	paths := make([]string, 0, len(versions))
	for _, v := range versions {
		asset, err := u.Method.Asset(v, u.GOARCH)
		if err != nil {
			return nil, err
		}
		dir := filepath.Join(work, v)
		if err = os.Mkdir(dir, 0o700); err != nil {
			return nil, err
		}
		path, err := u.Verifier.Fetch(ctx, u.Source, "v"+v, asset, dir)
		if err != nil {
			return nil, fmt.Errorf("verify %s: %w", asset, err)
		}
		paths = append(paths, path)
	}
	return paths, nil
}

// install applies an artifact, starts the service and waits until version is
// healthy and the service manager has not restarted it.
func (u *Updater) install(
	ctx context.Context,
	apply func(context.Context, string) error,
	artifact, version string,
) error {
	if err := apply(ctx, artifact); err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if err := u.Method.Start(ctx); err != nil {
		return fmt.Errorf("start the service: %w", err)
	}
	restarts, err := u.Method.Restarts(ctx)
	if err != nil {
		return fmt.Errorf("read the service restart count: %w", err)
	}
	timeout := u.HealthTimeout
	if timeout == 0 {
		timeout = HealthTimeout
	}
	healthCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	tick := time.NewTicker(healthPoll)
	defer tick.Stop()
	for {
		err = u.Health(healthCtx, version)
		if err == nil {
			break
		}
		select {
		case <-healthCtx.Done():
			return fmt.Errorf("health check: %s did not come up within %s: %w", version, timeout, err)
		case <-tick.C:
		}
	}
	after, err := u.Method.Restarts(ctx)
	if err != nil {
		return fmt.Errorf("read the service restart count: %w", err)
	}
	if after != restarts {
		return fmt.Errorf("health check: the service restarted %d times after %s started", after-restarts, version)
	}
	return nil
}

// rollback restores the state copy and reinstalls the previous version.
func (u *Updater) rollback(ctx context.Context, previous, state, from string) error {
	if err := u.Method.Stop(ctx); err != nil {
		return fmt.Errorf("stop the service: %w", err)
	}
	if err := restoreState(u.DataDir, state); err != nil {
		return fmt.Errorf("restore the state copy: %w", err)
	}
	return u.install(ctx, u.Method.Reapply, previous, from)
}

// parseVersion reads MAJOR.MINOR.PATCH; release tags carry nothing else.
func parseVersion(v string) ([3]int, bool) {
	var parts [3]int
	fields := strings.Split(v, ".")
	if len(fields) != len(parts) {
		return parts, false
	}
	for i, field := range fields {
		n, err := strconv.Atoi(field)
		if err != nil || n < 0 {
			return parts, false
		}
		parts[i] = n
	}
	return parts, true
}

func newer(a, b [3]int) bool {
	for i := range a {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}
