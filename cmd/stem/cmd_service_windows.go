// SPDX-License-Identifier: BUSL-1.1

//go:build windows

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/MustardSeedNetworks/foundation/pkg/securedir"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
	"github.com/MustardSeedNetworks/stem/internal/logging"
)

const (
	windowsServiceName        = "Stem"
	windowsServiceDisplayName = "The Stem"
	windowsServiceDescription = "Network performance testing daemon by Mustard Seed Networks"

	// serviceStateTimeout bounds how long start and stop wait for the service
	// manager to report the requested state, and how long the handler gives
	// the server to drain once it is told to stop.
	serviceStateTimeout = 30 * time.Second
	servicePollInterval = 250 * time.Millisecond

	// A start the daemon refuses (wrong data directory, port in use) must stop
	// rather than respin, the policy the systemd unit states: restart twice,
	// five seconds apart, then leave the service failed.
	serviceRestartDelay       = 5 * time.Second
	serviceFailureResetPeriod = 60 // seconds, the unit of ResetPeriod
)

// serviceCommandLine is the command the service manager runs. It is composed
// with Windows quoting, so an install under "Program Files" survives.
func serviceCommandLine(exe string) string {
	return windows.ComposeCommandLine([]string{exe, serviceCommandName, "run"})
}

func runServiceAction(action string, out io.Writer) error {
	if action == "run" {
		return runAsService()
	}
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect to the service manager (run from an elevated prompt): %w", err)
	}
	defer func() { _ = m.Disconnect() }()

	switch action {
	case "install":
		return installService(m, out)
	case "start":
		return startService(m, out)
	case "stop":
		return stopService(m, out)
	case "status":
		return statusService(m, out)
	default:
		return uninstallService(m, out)
	}
}

func openService(m *mgr.Mgr) (*mgr.Service, error) {
	s, err := m.OpenService(windowsServiceName)
	if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
		return nil, fmt.Errorf("service %s is not installed: run 'stem service install'", windowsServiceName)
	}
	if err != nil {
		return nil, fmt.Errorf("open service %s: %w", windowsServiceName, err)
	}
	return s, nil
}

// installService registers the service, or repoints an existing registration
// at this binary, so a reinstall from a newer archive needs no uninstall.
func installService(m *mgr.Mgr, out io.Writer) error {
	exe, err := serviceExecutable()
	if err != nil {
		return err
	}
	if err = ensureDataDir(daemonconn.ServiceDataDir()); err != nil {
		return err
	}

	verb := "installed"
	s, err := m.OpenService(windowsServiceName)
	switch {
	case err == nil:
		verb = "updated"
		cfg, cfgErr := s.Config()
		if cfgErr != nil {
			_ = s.Close()
			return fmt.Errorf("read service configuration: %w", cfgErr)
		}
		cfg.BinaryPathName = serviceCommandLine(exe)
		cfg.DisplayName = windowsServiceDisplayName
		cfg.Description = windowsServiceDescription
		cfg.StartType = mgr.StartAutomatic
		cfg.DelayedAutoStart = true
		if updateErr := s.UpdateConfig(cfg); updateErr != nil {
			_ = s.Close()
			return fmt.Errorf("update service configuration: %w", updateErr)
		}
	case errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST):
		s, err = m.CreateService(windowsServiceName, exe, mgr.Config{
			DisplayName:      windowsServiceDisplayName,
			Description:      windowsServiceDescription,
			StartType:        mgr.StartAutomatic,
			DelayedAutoStart: true,
		}, serviceCommandName, "run")
		if err != nil {
			return fmt.Errorf("create service: %w", err)
		}
	default:
		return fmt.Errorf("open service %s: %w", windowsServiceName, err)
	}
	defer func() { _ = s.Close() }()

	actions := []mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: serviceRestartDelay},
		{Type: mgr.ServiceRestart, Delay: serviceRestartDelay},
		{Type: mgr.NoAction},
	}
	if err = s.SetRecoveryActions(actions, serviceFailureResetPeriod); err != nil {
		return fmt.Errorf("set service recovery actions: %w", err)
	}
	// The daemon exits with a service-specific code when it refuses to start;
	// without this the manager treats that as a clean stop and never restarts.
	if err = s.SetRecoveryActionsOnNonCrashFailures(true); err != nil {
		return fmt.Errorf("set service recovery actions: %w", err)
	}

	// A source already registered is not a failure; the log is a convenience.
	_ = eventlog.InstallAsEventCreate(windowsServiceName, eventlog.Error)

	_, err = fmt.Fprintf(out, "Service %s %s. It starts automatically at boot.\n\nStart it now: stem service start\n",
		windowsServiceName, verb)
	return err
}

func startService(m *mgr.Mgr, out io.Writer) error {
	s, err := openService(m)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()
	if err = s.Start(); err != nil && !errors.Is(err, windows.ERROR_SERVICE_ALREADY_RUNNING) {
		return fmt.Errorf("start service: %w", err)
	}
	if err = waitForState(s.Query, svc.Running, serviceStateTimeout); err != nil {
		return fmt.Errorf("service did not reach Running (see %s): %w", serviceLogPath(), err)
	}
	_, err = fmt.Fprintln(out, "Service started.")
	return err
}

func stopService(m *mgr.Mgr, out io.Writer) error {
	s, err := openService(m)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()
	if err = stopAndWait(s); err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, "Service stopped.")
	return err
}

func stopAndWait(s *mgr.Service) error {
	_, err := s.Control(svc.Stop)
	if err != nil && !errors.Is(err, windows.ERROR_SERVICE_NOT_ACTIVE) {
		return fmt.Errorf("stop service: %w", err)
	}
	return waitForState(s.Query, svc.Stopped, serviceStateTimeout)
}

func statusService(m *mgr.Mgr, out io.Writer) error {
	s, err := openService(m)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()
	st, err := s.Query()
	if err != nil {
		return fmt.Errorf("query service: %w", err)
	}
	_, err = fmt.Fprintf(out, "Service: %s\nStatus:  %s\n", windowsServiceDisplayName, stateName(st.State))
	return err
}

func uninstallService(m *mgr.Mgr, out io.Writer) error {
	s, err := openService(m)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()
	if err = stopAndWait(s); err != nil {
		return err
	}
	if err = s.Delete(); err != nil {
		return fmt.Errorf("remove service: %w", err)
	}
	_ = eventlog.Remove(windowsServiceName)
	_, err = fmt.Fprintf(out, "Service removed. Configuration and data in %s are kept.\n", daemonconn.ServiceDataDir())
	return err
}

func stateName(state svc.State) string {
	names := map[svc.State]string{
		svc.Stopped:         "Stopped",
		svc.StartPending:    "Starting",
		svc.StopPending:     "Stopping",
		svc.Running:         "Running",
		svc.ContinuePending: "Resuming",
		svc.PausePending:    "Pausing",
		svc.Paused:          "Paused",
	}
	if name, ok := names[state]; ok {
		return name
	}
	return fmt.Sprintf("Unknown (%d)", uint32(state))
}

// waitForState polls query until the service reports want, or fails once
// timeout has passed. A service that reaches Stopped while a different state
// is wanted has failed to start, so that is reported at once.
func waitForState(query func() (svc.Status, error), want svc.State, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		st, err := query()
		if err != nil {
			return fmt.Errorf("query service: %w", err)
		}
		if st.State == want {
			return nil
		}
		if st.State == svc.Stopped {
			return errors.New("the service stopped")
		}
		if time.Now().After(deadline) {
			return fmt.Errorf(
				"timed out after %s in state %s, wanted %s",
				timeout,
				stateName(st.State),
				stateName(want),
			)
		}
		time.Sleep(servicePollInterval)
	}
}

func serviceLogPath() string {
	return filepath.Join(daemonconn.ServiceDataDir(), "logs", "stem.log")
}

// runAsService hands the process to the service manager, which calls the
// handler until the service stops.
func runAsService() error {
	if err := svc.Run(
		windowsServiceName,
		serviceHandler{prepare: prepareDaemon, report: reportToEventLog},
	); err != nil {
		return fmt.Errorf(
			"%w: 'stem service run' is started by the Windows service manager; use 'stem web' to run in the foreground",
			err,
		)
	}
	return nil
}

// serviceHandler adapts the daemon to the service manager's control loop.
type serviceHandler struct {
	// prepare checks everything that can be checked before the service claims
	// to be running (the data directory, logging, the server's credentials),
	// and returns the function that serves until its context is cancelled,
	// shuts down and returns. A prepare error means the service never reports
	// Running, so `stem service start` and the installer see the refusal.
	prepare func() (run func(ctx context.Context) error, err error)
	// report records a refusal where an operator can find it, as the file log
	// may not exist yet.
	report func(error)
}

// Execute implements svc.Handler.
func (h serviceHandler) Execute(
	_ []string,
	requests <-chan svc.ChangeRequest,
	status chan<- svc.Status,
) (bool, uint32) {
	const accepted = svc.AcceptStop | svc.AcceptShutdown
	status <- svc.Status{State: svc.StartPending}

	run, prepareErr := h.prepare()
	if prepareErr != nil {
		h.report(prepareErr)
		return true, 1
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- run(ctx) }()
	status <- svc.Status{State: svc.Running, Accepts: accepted}

	for {
		select {
		case runErr := <-done:
			// Nothing asked it to stop, so the daemon died. A specific exit
			// code lets the recovery actions treat it as a failure.
			if runErr != nil {
				logging.Error("service daemon exited", "error", runErr.Error())
			}
			return true, 1
		case req := <-requests:
			if req.Cmd == svc.Interrogate {
				status <- req.CurrentStatus
				continue
			}
			// Only Stop and Shutdown are accepted, so nothing else arrives.
			if req.Cmd != svc.Stop && req.Cmd != svc.Shutdown {
				continue
			}
			status <- svc.Status{State: svc.StopPending, WaitHint: uint32(serviceStateTimeout / time.Millisecond)}
			cancel()
			awaitDaemon(done)
			return false, 0
		}
	}
}

// awaitDaemon waits, up to serviceStateTimeout, for the daemon to finish
// shutting down after its context was cancelled.
func awaitDaemon(done <-chan error) {
	select {
	case err := <-done:
		if err != nil {
			logging.Error("service daemon shutdown", "error", err.Error())
		}
	case <-time.After(serviceStateTimeout):
		logging.Error("service daemon did not stop in time", "timeout", serviceStateTimeout.String())
	}
}

// dataDirOptions is what foundation's securedir needs to judge the data
// directory: SYSTEM, Administrators and TrustedInstaller only, no extras.
func dataDirOptions() securedir.Options { return securedir.Options{} }

// ensureDataDir is `stem service install`'s half, run elevated. A missing
// directory is created by securedir.Ensure with a protected descriptor. An
// existing one is left alone: install.ps1 judged it (and refuses a hostile
// one), and the service verifies it at every start. It is not verified here
// because the daemon's private key is readable by SYSTEM alone, so an
// Administrators token cannot open it to check, and securedir fails closed on
// what it cannot read.
func ensureDataDir(dir string) error {
	canonical, err := securedir.Canonical(dir, dataDirOptions())
	if err != nil {
		return fmt.Errorf("data directory %s: %w", dir, err)
	}
	parent, err := os.OpenRoot(filepath.Dir(canonical))
	if err != nil {
		return fmt.Errorf("data directory %s: %w", canonical, err)
	}
	defer func() { _ = parent.Close() }()
	if _, err = parent.Lstat(filepath.Base(canonical)); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("data directory %s: %w", canonical, err)
	}
	if err = useAdministratorsAsDefaultOwner(); err != nil {
		return fmt.Errorf("create data directory %s: %w", canonical, err)
	}
	if err = securedir.Ensure(canonical, dataDirOptions()); err != nil {
		return fmt.Errorf("create data directory %s: %w", canonical, err)
	}
	return nil
}

// tokenOwner is the TOKEN_OWNER structure SetTokenInformation reads.
type tokenOwner struct{ owner *windows.SID }

// useAdministratorsAsDefaultOwner makes BUILTIN\Administrators the owner of
// what this process creates next. securedir.Ensure leaves the owner to the
// creating principal and then refuses an untrusted one, and an elevated token
// whose default owner is the user (the GitHub-hosted runner's, and any machine
// that changed the "default owner" policy) would create a directory its own
// install then rejects. Administrators is an owner-eligible group of an
// elevated token, so naming it as the default owner needs no extra privilege.
func useAdministratorsAsDefaultOwner() error {
	administrators, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		return fmt.Errorf("resolve the Administrators SID: %w", err)
	}
	var token windows.Token
	err = windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_ADJUST_DEFAULT|windows.TOKEN_QUERY, &token)
	if err != nil {
		return fmt.Errorf("open the process token: %w", err)
	}
	defer func() { _ = token.Close() }()
	owner := tokenOwner{owner: administrators}
	setTokenInformation := windows.NewLazySystemDLL("advapi32.dll").NewProc("SetTokenInformation")
	r1, _, callErr := setTokenInformation.Call(uintptr(token), uintptr(windows.TokenOwner),
		uintptr(unsafe.Pointer(&owner)), unsafe.Sizeof(owner))
	if r1 == 0 {
		return fmt.Errorf("set the token's default owner to Administrators: %w", callErr)
	}
	return nil
}

// prepareDaemon is the start of stem web as the service runs it. The SCM starts
// a service in System32 with no console and no environment of its own, so the
// working directory and data directory are pinned here the way the systemd unit
// pins WorkingDirectory and StateDirectory. The data directory is resolved once
// and verified, never repaired: a tree that is not private is a refusal to
// start.
func prepareDaemon() (func(context.Context) error, error) {
	dir, err := securedir.Canonical(daemonconn.ServiceDataDir(), dataDirOptions())
	if err != nil {
		return nil, fmt.Errorf("data directory: %w", err)
	}
	if err = securedir.Verify(dir, dataDirOptions()); err != nil {
		return nil, fmt.Errorf(
			"data directory %s is not private (%w); inspect it, then run install.ps1 as Administrator",
			dir,
			err,
		)
	}
	if err = os.Setenv("STEM_DATA_DIR", dir); err != nil {
		return nil, fmt.Errorf("set data directory: %w", err)
	}
	// An absolute certificate directory, so nothing depends on the working
	// directory the daemon is started in.
	if os.Getenv("STEM_TLS_CERTS_DIR") == "" {
		if err = os.Setenv("STEM_TLS_CERTS_DIR", filepath.Join(dir, "certs")); err != nil {
			return nil, fmt.Errorf("set certificate directory: %w", err)
		}
	}
	if err = os.Chdir(dir); err != nil {
		return nil, fmt.Errorf("enter data directory: %w", err)
	}

	logLevel := os.Getenv("STEM_LOG_LEVEL")
	if logLevel == "" {
		logLevel = logging.LevelInfo
	}
	cfg := logging.DefaultConfig()
	cfg.Level = logLevel
	cfg.File = filepath.Join(dir, "logs", "stem.log")
	cfg.Component = "stem"
	if err = logging.Init(cfg); err != nil {
		return nil, fmt.Errorf("start logging: %w", err)
	}

	// The port is explicit so a busy 8444 refuses to start instead of walking
	// to a neighbour the firewall rule does not open.
	srv, err := api.NewServer(api.ListenAddr{Port: defaultWebPort, Explicit: true})
	if err != nil {
		return nil, err
	}
	return srv.RunContext, nil
}

// reportToEventLog writes a refusal to the Windows Application log under the
// service's name; install registers the source.
func reportToEventLog(cause error) {
	log, err := eventlog.Open(windowsServiceName)
	if err != nil {
		return
	}
	defer func() { _ = log.Close() }()
	_ = log.Error(1, "Stem service refused to start: "+cause.Error())
}
