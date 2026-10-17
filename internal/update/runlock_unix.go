// SPDX-License-Identifier: BUSL-1.1

//go:build unix

package update

import (
	"errors"
	"os"
	"syscall"
)

// flock, not fcntl(F_SETLK): a POSIX record lock belongs to the process, so
// it could not keep two holders in one process apart, and the guard would be
// untestable in-process. flock belongs to the open file description.
func tryLock(file *os.File) (bool, error) {
	err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, syscall.EWOULDBLOCK):
		return false, nil
	default:
		return false, err
	}
}

func unlock(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}
