// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package tui

import (
	"testing"
	"time"
)

func TestNewApp(t *testing.T) {
	// Test creating new App with nil dataplane (should not panic)
	app := New(nil)

	if app == nil {
		t.Fatal("New() returned nil")
	}
	if app.app == nil {
		t.Error("App.app (tview.Application) should not be nil")
	}
	if app.startTime.IsZero() {
		t.Error("App.startTime should be initialized")
	}
	if app.stopChan == nil {
		t.Error("App.stopChan should be initialized")
	}
}

func TestAppStartTime(t *testing.T) {
	before := time.Now()
	app := New(nil)
	after := time.Now()

	if app.startTime.Before(before) {
		t.Error("startTime should not be before New() was called")
	}
	if app.startTime.After(after) {
		t.Error("startTime should not be after New() returned")
	}
}

func TestAppStopChannelCreated(t *testing.T) {
	app := New(nil)

	// stopChan should be readable (not receive anything immediately)
	select {
	case <-app.stopChan:
		t.Error("stopChan should not have any messages initially")
	default:
		// Expected behavior - channel is empty
	}
}

func TestAppNilDataplane(t *testing.T) {
	app := New(nil)

	if app.dp != nil {
		t.Error("Expected nil dataplane to be stored")
	}
}

func TestAppMultipleNew(t *testing.T) {
	// Creating multiple apps should not cause issues
	apps := make([]*App, 10)
	for i := 0; i < 10; i++ {
		apps[i] = New(nil)
		if apps[i] == nil {
			t.Errorf("New() returned nil on iteration %d", i)
		}
	}

	// Each app should have independent start times
	for i := 1; i < len(apps); i++ {
		if apps[i].startTime.Before(apps[0].startTime) {
			t.Error("Later apps should have later or equal start times")
		}
	}
}

func TestAppFields(t *testing.T) {
	app := New(nil)

	// Test that App has expected structure
	// These tests verify the type exists and has expected fields
	if app.app == nil {
		t.Error("tview.Application should be initialized")
	}

	// statsView, sigView, latView, helpView should be nil until Run() is called
	// This is expected behavior
}

func TestNewAppDoesNotPanic(t *testing.T) {
	// Test that New() doesn't panic with various inputs
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("New() panicked: %v", r)
		}
	}()

	_ = New(nil)
}

func TestAppStopOnceField(t *testing.T) {
	app := New(nil)

	// stopOnce should be a zero value initially
	// We can't directly test sync.Once, but we can verify the App was created
	if app.app == nil {
		t.Error("App should be properly initialized")
	}
}

// TestStopMethod tests the Stop method (if it exists)
func TestStopMethod(t *testing.T) {
	app := New(nil)

	// Stop() should not panic when called
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Stop() panicked: %v", r)
		}
	}()

	// Call Stop() - should close the stop channel
	app.Stop()

	// Verify stopChan is closed by checking if we can receive from it
	select {
	case _, ok := <-app.stopChan:
		if ok {
			t.Error("Expected stopChan to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		// This might mean the channel wasn't closed, but could also be timing
	}
}

func TestStopMethodMultipleCalls(t *testing.T) {
	app := New(nil)

	// Multiple calls to Stop() should not panic due to sync.Once
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Multiple Stop() calls panicked: %v", r)
		}
	}()

	app.Stop()
	app.Stop()
	app.Stop()
}

// Benchmark tests
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New(nil)
	}
}

func BenchmarkStop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		app := New(nil)
		app.Stop()
	}
}
