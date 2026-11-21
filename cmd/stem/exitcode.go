// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"errors"
	"os"
)

// Process exit statuses. Scripts branch on these, so they are the CLI's
// contract. A command that fails for any reason is exitError. `stem status`
// has its own meaning for 1 and 2, the same as niac's: 1 is "not running" and
// 2 is "could not tell", so a script can tell a stopped daemon from a check
// that failed.
const (
	exitOK = 0

	exitError = 1

	statusNotRunning = 1
	statusError      = 2
)

// codedError carries a specific exit status out to the single exit point.
type codedError struct {
	code int
	err  error
}

func (e codedError) Error() string { return e.err.Error() }

func (e codedError) Unwrap() error { return e.err }

// withExitCode tags err with the exit status it should produce. A nil err stays
// nil, so a caller can wrap unconditionally.
func withExitCode(code int, err error) error {
	if err == nil {
		return nil
	}
	return codedError{code: code, err: err}
}

// exitCodeOf maps a command's error to its exit status: success for nil, the
// tagged status for a codedError, exitError for anything else.
func exitCodeOf(err error) int {
	if err == nil {
		return exitOK
	}
	if coded, ok := errors.AsType[codedError](err); ok {
		return coded.code
	}
	return exitError
}

// exitOnError ends the process with err's status. Every command's failure goes
// through here, so no verb picks its own number.
func exitOnError(err error) {
	if code := exitCodeOf(err); code != exitOK {
		os.Exit(code)
	}
}
