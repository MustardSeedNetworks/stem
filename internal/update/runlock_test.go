// SPDX-License-Identifier: BUSL-1.1

package update_test

import (
	"errors"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/update"
)

// Two guards on one directory stand in for the daemon and `stem update`: the
// lock belongs to the open file, so they exclude each other exactly as two
// processes do.
func TestRunGuardExcludesASecondHolder(t *testing.T) {
	dir := t.TempDir()
	daemon := update.NewRunGuard(dir)
	updater := update.NewRunGuard(dir)

	release, err := daemon.Acquire()
	if err != nil {
		t.Fatalf("daemon Acquire: %v", err)
	}
	if _, heldErr := updater.Acquire(); !errors.Is(heldErr, update.ErrRunLockHeld) {
		t.Fatalf("updater Acquire during a run = %v; want ErrRunLockHeld", heldErr)
	}

	release()
	releaseUpdate, err := updater.Acquire()
	if err != nil {
		t.Fatalf("updater Acquire after the run = %v; want the lock", err)
	}
	if _, heldErr := daemon.Acquire(); !errors.Is(heldErr, update.ErrRunLockHeld) {
		t.Fatalf("daemon Acquire during an update = %v; want ErrRunLockHeld", heldErr)
	}
	releaseUpdate()
}

// The lock stays held until the last holder in the process lets go, and a
// release called twice counts once: a cancelled plan's late release must not
// drop the hold of the run that replaced it.
func TestRunGuardHoldsUntilTheLastHolder(t *testing.T) {
	dir := t.TempDir()
	daemon := update.NewRunGuard(dir)
	updater := update.NewRunGuard(dir)

	first, err := daemon.Acquire()
	if err != nil {
		t.Fatalf("first Acquire: %v", err)
	}
	second, err := daemon.Acquire()
	if err != nil {
		t.Fatalf("second Acquire: %v", err)
	}

	first()
	first()
	if _, heldErr := updater.Acquire(); !errors.Is(heldErr, update.ErrRunLockHeld) {
		t.Fatalf("updater Acquire with one holder left = %v; want ErrRunLockHeld", heldErr)
	}

	second()
	release, err := updater.Acquire()
	if err != nil {
		t.Fatalf("updater Acquire after the last holder = %v; want the lock", err)
	}
	release()
}

func TestRunGuardReportsAMissingDirectory(t *testing.T) {
	guard := update.NewRunGuard(t.TempDir() + "/missing")
	_, err := guard.Acquire()
	if err == nil || errors.Is(err, update.ErrRunLockHeld) {
		t.Fatalf("Acquire in a missing directory = %v; want an open error", err)
	}
}
