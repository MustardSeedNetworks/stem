// SPDX-License-Identifier: BUSL-1.1

package update

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

// The lock is one byte at offset 0. The file holds no data, so nothing needs
// to read past the locked range.
func tryLock(file *os.File) (bool, error) {
	err := windows.LockFileEx(
		windows.Handle(file.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
		0, 1, 0,
		&windows.Overlapped{},
	)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, windows.ERROR_LOCK_VIOLATION):
		return false, nil
	default:
		return false, err
	}
}

func unlock(file *os.File) error {
	return windows.UnlockFileEx(windows.Handle(file.Fd()), 0, 1, 0, &windows.Overlapped{})
}
