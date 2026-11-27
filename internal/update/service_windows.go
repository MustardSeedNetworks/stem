// SPDX-License-Identifier: BUSL-1.1

//go:build windows

package update

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	windowsServiceName  = "Stem"
	serviceStateTimeout = 30 * time.Second
	servicePollInterval = 250 * time.Millisecond
)

// serviceHandle is the part of *mgr.Service the update uses.
type serviceHandle interface {
	Query() (svc.Status, error)
	Control(c svc.Cmd) (svc.Status, error)
	Start(args ...string) error
}

// scmService drives the Stem service through the service control manager.
// open returns a nil handle, and no error, when the service is not installed.
type scmService struct {
	open func() (serviceHandle, func(), error)
	// stopped records that Stop took a running service down, so Start brings
	// back only what the update stopped: a service the operator left stopped
	// stays stopped.
	stopped bool
}

func newServiceController() serviceController { return &scmService{open: openStemService} }

func openStemService() (serviceHandle, func(), error) {
	m, err := mgr.Connect()
	if err != nil {
		return nil, nil, fmt.Errorf("connect to the service manager: %w", err)
	}
	s, err := m.OpenService(windowsServiceName)
	if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
		_ = m.Disconnect()
		return nil, nil, nil
	}
	if err != nil {
		_ = m.Disconnect()
		return nil, nil, fmt.Errorf("open service %s: %w", windowsServiceName, err)
	}
	return s, func() { _ = s.Close(); _ = m.Disconnect() }, nil
}

func (c *scmService) Installed() bool {
	handle, closeFn, err := c.open()
	if err != nil || handle == nil {
		return false
	}
	closeFn()
	return true
}

func (c *scmService) Stop(ctx context.Context) error {
	handle, closeFn, err := c.open()
	if err != nil || handle == nil {
		return err
	}
	defer closeFn()
	st, err := handle.Query()
	if err != nil {
		return fmt.Errorf("query service: %w", err)
	}
	if st.State == svc.Stopped {
		return nil
	}
	if _, err = handle.Control(svc.Stop); err != nil && !errors.Is(err, windows.ERROR_SERVICE_NOT_ACTIVE) {
		return fmt.Errorf("stop service: %w", err)
	}
	if err = waitForState(ctx, handle, svc.Stopped); err != nil {
		return err
	}
	c.stopped = true
	return nil
}

func (c *scmService) Start(ctx context.Context) error {
	if !c.stopped {
		return nil
	}
	handle, closeFn, err := c.open()
	if err != nil || handle == nil {
		return err
	}
	defer closeFn()
	if err = handle.Start(); err != nil && !errors.Is(err, windows.ERROR_SERVICE_ALREADY_RUNNING) {
		return fmt.Errorf("start service: %w", err)
	}
	if err = waitForState(ctx, handle, svc.Running); err != nil {
		return err
	}
	c.stopped = false
	return nil
}

// waitForState polls the service until it reports want. A service that is
// Stopped while Running is wanted has failed to start.
func waitForState(ctx context.Context, handle serviceHandle, want svc.State) error {
	deadline := time.Now().Add(serviceStateTimeout)
	for {
		st, err := handle.Query()
		if err != nil {
			return fmt.Errorf("query service: %w", err)
		}
		if st.State == want {
			return nil
		}
		if want == svc.Running && st.State == svc.Stopped {
			return errors.New("the service stopped instead of starting")
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("service %s did not reach state %d within %s",
				windowsServiceName, want, serviceStateTimeout)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(servicePollInterval):
		}
	}
}
