// SPDX-License-Identifier: BUSL-1.1

//go:build windows

package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
)

type handlerRun struct {
	states      []svc.State
	svcSpecific bool
	code        uint32
}

// drive runs the handler against fake service-manager channels and returns
// the statuses it reported in order, plus the result of Execute.
func drive(t *testing.T, run func(context.Context) error, requests ...svc.Cmd) handlerRun {
	return driveHandler(t, serviceHandler{
		prepare: func() (func(context.Context) error, error) { return run, nil },
		report:  func(error) {},
	}, requests...)
}

func driveHandler(t *testing.T, handler serviceHandler, requests ...svc.Cmd) handlerRun {
	t.Helper()

	reqCh := make(chan svc.ChangeRequest, len(requests))
	for _, cmd := range requests {
		reqCh <- svc.ChangeRequest{Cmd: cmd}
	}
	statusCh := make(chan svc.Status, 16)
	var result handlerRun
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		result.svcSpecific, result.code = handler.Execute(nil, reqCh, statusCh)
	}()
	select {
	case <-finished:
	case <-time.After(5 * time.Second):
		t.Fatal("Execute did not return")
	}
	close(statusCh)
	for st := range statusCh {
		result.states = append(result.states, st.State)
	}
	return result
}

func TestWindowsServiceHandlerStopCancelsDaemon(t *testing.T) {
	cancelled := make(chan struct{})
	run := func(ctx context.Context) error {
		<-ctx.Done()
		close(cancelled)
		return nil
	}
	got := drive(t, run, svc.Stop)
	select {
	case <-cancelled:
	default:
		t.Error("stop did not cancel the daemon's context")
	}
	want := []svc.State{svc.StartPending, svc.Running, svc.StopPending}
	if !slices.Equal(got.states, want) {
		t.Fatalf("states = %v, want %v", got.states, want)
	}
	if got.svcSpecific || got.code != 0 {
		t.Errorf("Execute = (%v, %d), want a clean stop (false, 0)", got.svcSpecific, got.code)
	}
}

func TestWindowsServiceHandlerDaemonFailureIsAFailure(t *testing.T) {
	run := func(context.Context) error { return errors.New("port in use") }
	got := drive(t, run)
	if !got.svcSpecific || got.code == 0 {
		t.Errorf("Execute = (%v, %d), want a service-specific failure code", got.svcSpecific, got.code)
	}
}

func TestWindowsServiceHandlerAnswersInterrogate(t *testing.T) {
	run := func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	}
	got := drive(t, run, svc.Interrogate, svc.Shutdown)
	if len(got.states) != 4 {
		t.Fatalf("states = %v, want StartPending, Running, the interrogate reply and StopPending", got.states)
	}
}

func TestWindowsServiceCommandLineQuotesPath(t *testing.T) {
	got := serviceCommandLine(`C:\Program Files\Stem\stem.exe`)
	want := `"C:\Program Files\Stem\stem.exe" service run`
	if got != want {
		t.Errorf("serviceCommandLine = %s, want %s", got, want)
	}
}

func TestWindowsServiceWaitForState(t *testing.T) {
	t.Run("reaches the wanted state", func(t *testing.T) {
		calls := 0
		query := func() (svc.Status, error) {
			calls++
			if calls < 3 {
				return svc.Status{State: svc.StartPending}, nil
			}
			return svc.Status{State: svc.Running}, nil
		}
		if err := waitForState(query, svc.Running, time.Minute); err != nil {
			t.Fatalf("waitForState = %v", err)
		}
	})
	t.Run("a stopped service failed to start", func(t *testing.T) {
		query := func() (svc.Status, error) { return svc.Status{State: svc.Stopped}, nil }
		err := waitForState(query, svc.Running, time.Minute)
		if err == nil || !strings.Contains(err.Error(), "stopped") {
			t.Fatalf("waitForState = %v, want a stopped error", err)
		}
	})
	t.Run("times out", func(t *testing.T) {
		query := func() (svc.Status, error) { return svc.Status{State: svc.StartPending}, nil }
		err := waitForState(query, svc.Running, time.Millisecond)
		if err == nil || !strings.Contains(err.Error(), "timed out") {
			t.Fatalf("waitForState = %v, want a timeout", err)
		}
	})
}

// requireElevated skips a test that creates a private directory, which only an
// elevated token may do.
func requireElevated(t *testing.T) {
	t.Helper()
	if !windows.GetCurrentProcessToken().IsElevated() {
		t.Skip("needs an elevated token")
	}
}

func TestWindowsServiceEnsureDataDirCreatesAPrivateDirectory(t *testing.T) {
	requireElevated(t)
	// Under ProgramData, whose parent ACL a SYSTEM service trusts; the temp
	// directory belongs to the runner's own account.
	dir := filepath.Join(os.Getenv("ProgramData"), "stem-test-"+filepath.Base(t.TempDir()))
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := ensureDataDir(dir); err != nil {
		t.Fatalf("ensureDataDir on a missing directory = %v", err)
	}
	if err := ensureDataDir(dir); err != nil {
		t.Fatalf("ensureDataDir on the directory it created = %v", err)
	}
}

func TestWindowsServiceUnderDir(t *testing.T) {
	tests := []struct {
		dir, file string
		want      bool
	}{
		{`C:\Program Files\Stem`, `C:\Program Files\Stem\stem.exe`, true},
		{`C:\Program Files\Stem`, `c:\program files\stem\STEM.EXE`, true},
		{`C:\Program Files\Stem`, `C:\Program Files\Stem2\stem.exe`, false},
		{`C:\Program Files\Stem`, `C:\Users\u\Downloads\stem.exe`, false},
		{`C:\Program Files\Stem`, `C:\Program Files\Stem\..\Evil\stem.exe`, false},
	}
	for _, tt := range tests {
		if got := underDir(tt.dir, tt.file); got != tt.want {
			t.Errorf("underDir(%s, %s) = %v, want %v", tt.dir, tt.file, got, tt.want)
		}
	}
}
