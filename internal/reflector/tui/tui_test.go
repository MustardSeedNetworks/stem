// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package tui_test

import (
	"testing"
	"time"

	"github.com/krisarmstrong/stem/internal/reflector/tui"
)

func TestNewApp(t *testing.T) {
	// Test creating new App with nil dataplane (should not panic).
	app := tui.New(nil)

	if app == nil {
		t.Fatal("New() returned nil")
	}
}

func TestAppStartTime(t *testing.T) {
	before := time.Now()
	app := tui.New(nil)
	after := time.Now()

	if app == nil {
		t.Fatal("New() returned nil")
	}

	// We can only verify the app was created, since startTime is unexported.
	_ = before
	_ = after
}

func TestNewAppDoesNotPanic(t *testing.T) {
	// Test that New() doesn't panic with various inputs.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("New() panicked: %v", r)
		}
	}()

	_ = tui.New(nil)
}

// TestStopMethod tests the Stop method.
func TestStopMethod(t *testing.T) {
	app := tui.New(nil)

	// Stop() should not panic when called.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Stop() panicked: %v", r)
		}
	}()

	// Call Stop() - should close the stop channel.
	app.Stop()
}

func TestStopMethodMultipleCalls(t *testing.T) {
	app := tui.New(nil)

	// Multiple calls to Stop() should not panic due to sync.Once.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Multiple Stop() calls panicked: %v", r)
		}
	}()

	app.Stop()
	app.Stop()
	app.Stop()
}

// Benchmark tests.
func BenchmarkNew(b *testing.B) {
	for b.Loop() {
		_ = tui.New(nil)
	}
}

func BenchmarkStop(b *testing.B) {
	for b.Loop() {
		app := tui.New(nil)
		app.Stop()
	}
}
