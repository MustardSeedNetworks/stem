// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestExitCodeOf(t *testing.T) {
	t.Parallel()

	base := errors.New("boom")
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil is success", nil, exitOK},
		{"a plain error is exitError", base, exitError},
		{"tagged could not tell", withExitCode(statusError, base), statusError},
		{"tagged error", withExitCode(exitError, base), exitError},
		{"a tag survives wrapping", fmt.Errorf("context: %w", withExitCode(statusError, base)), statusError},
		{"tagging nil stays nil", withExitCode(statusError, nil), exitOK},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := exitCodeOf(tc.err); got != tc.want {
				t.Errorf("exitCodeOf = %d, want %d", got, tc.want)
			}
		})
	}
}

// The numbers are the contract scripts branch on, so they are pinned.
func TestExitStatusContract(t *testing.T) {
	t.Parallel()

	if exitOK != 0 || exitError != 1 || statusNotRunning != 1 || statusError != 2 {
		t.Fatalf("exit statuses = %d/%d/%d/%d, want 0/1/1/2", exitOK, exitError, statusNotRunning, statusError)
	}
}

// TestExitStatusHelperProcess runs main with the arguments in the environment,
// so the tests below see the status the process really exits with.
func TestExitStatusHelperProcess(_ *testing.T) {
	raw, ok := os.LookupEnv("STEM_CLI_EXIT_ARGS")
	if !ok {
		return
	}
	os.Args = append([]string{"stem"}, strings.Split(raw, "\x1f")...)
	main()
}

func runStem(t *testing.T, env []string, args ...string) (int, string) {
	t.Helper()

	process := exec.Command(os.Args[0], "-test.run=TestExitStatusHelperProcess")
	process.Env = append(os.Environ(), env...)
	process.Env = append(process.Env, "STEM_CLI_EXIT_ARGS="+strings.Join(args, "\x1f"))
	output, err := process.CombinedOutput()
	if err == nil {
		return 0, string(output)
	}
	exitErr, ok := errors.AsType[*exec.ExitError](err)
	if !ok {
		t.Fatalf("run stem %v: %v", args, err)
	}
	return exitErr.ExitCode(), string(output)
}

// TestProcessExitStatus drives each verb through main and reads the status the
// operating system sees.
func TestProcessExitStatus(t *testing.T) {
	healthy := startStubDaemon(t, healthyDaemon())
	notReady := startStubDaemon(t, notReadyDaemon())
	closed := closedURL(t)
	emptyDir := t.TempDir()
	env := []string{"STEM_DATA_DIR=" + emptyDir}

	tests := []struct {
		name string
		args []string
		want int
	}{
		{"status healthy", []string{"status", "--api", healthy.url, "--cacert", healthy.caFile}, exitOK},
		{"status not ready", []string{"status", "--api", notReady.url, "--cacert", notReady.caFile}, statusNotRunning},
		{"status not running", []string{"status", "--api", closed}, statusNotRunning},
		{"status json not running", []string{"status", "--json", "--api", closed}, statusNotRunning},
		{"status untrusted certificate", []string{"status", "--api", healthy.url}, statusError},
		{"status bad flag", []string{"status", "--nope"}, statusError},
		{"status stray argument", []string{"status", "extra"}, statusError},
		{"version", []string{"version"}, exitOK},
		{"version json", []string{"version", "--json"}, exitOK},
		{"version stray argument", []string{"version", "extra"}, exitError},
		{"backup with no state", []string{"backup", "-o", t.TempDir() + "/b.tar.gz"}, exitError},
		{"restore without an archive", []string{"restore"}, exitError},
		{"logs lines out of range", []string{"logs", "--lines", "0"}, exitError},
		{"unknown command", []string{"frobnicate"}, exitError},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, output := runStem(t, env, tc.args...)
			if got != tc.want {
				t.Errorf("stem %v exited %d, want %d\n%s", tc.args, got, tc.want, output)
			}
		})
	}
}
