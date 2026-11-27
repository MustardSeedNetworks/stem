// SPDX-License-Identifier: BUSL-1.1

//go:build windows

package update

import (
	"context"
	"testing"

	"golang.org/x/sys/windows/svc"
)

// fakeHandle is a service that moves to the state a control or start asks for.
type fakeHandle struct {
	state  svc.State
	stops  int
	starts int
}

func (f *fakeHandle) Query() (svc.Status, error) { return svc.Status{State: f.state}, nil }

func (f *fakeHandle) Control(c svc.Cmd) (svc.Status, error) {
	if c == svc.Stop {
		f.stops++
		f.state = svc.Stopped
	}
	return svc.Status{State: f.state}, nil
}

func (f *fakeHandle) Start(...string) error {
	f.starts++
	f.state = svc.Running
	return nil
}

func controllerFor(h *fakeHandle) *scmService {
	return &scmService{open: func() (serviceHandle, func(), error) {
		if h == nil {
			return nil, nil, nil
		}
		return h, func() {}, nil
	}}
}

func TestWindowsServiceStopsAndRestartsARunningService(t *testing.T) {
	h := &fakeHandle{state: svc.Running}
	c := controllerFor(h)
	ctx := context.Background()
	if err := c.Stop(ctx); err != nil {
		t.Fatalf("Stop = %v", err)
	}
	if h.stops != 1 || h.state != svc.Stopped {
		t.Fatalf("after Stop: stops %d, state %d", h.stops, h.state)
	}
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start = %v", err)
	}
	if h.starts != 1 || h.state != svc.Running {
		t.Fatalf("after Start: starts %d, state %d", h.starts, h.state)
	}
}

func TestWindowsServiceLeavesAStoppedServiceStopped(t *testing.T) {
	h := &fakeHandle{state: svc.Stopped}
	c := controllerFor(h)
	ctx := context.Background()
	if err := c.Stop(ctx); err != nil {
		t.Fatalf("Stop = %v", err)
	}
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start = %v", err)
	}
	if h.stops != 0 || h.starts != 0 {
		t.Fatalf("a stopped service was touched: stops %d, starts %d", h.stops, h.starts)
	}
}

func TestWindowsServiceNotInstalledIsANoOp(t *testing.T) {
	c := controllerFor(nil)
	ctx := context.Background()
	if c.Installed() {
		t.Fatal("Installed = true for a missing service")
	}
	if err := c.Stop(ctx); err != nil {
		t.Fatalf("Stop = %v", err)
	}
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start = %v", err)
	}
}
