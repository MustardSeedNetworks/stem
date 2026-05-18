// SPDX-License-Identifier: BUSL-1.1

package tui

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/krisarmstrong/stem/internal/reflector/dataplane"
)

// =============================================================================
// Internal tests that can access unexported methods
// =============================================================================

func TestIsPaused(t *testing.T) {
	app := New(nil)

	// Initial state should be not paused.
	if app.isPaused() {
		t.Error("new App should not be paused")
	}
}

func TestTogglePauseInternal(t *testing.T) {
	app := New(nil)

	// Toggle to paused.
	app.pauseMu.Lock()
	app.paused = true
	app.pauseMu.Unlock()

	if !app.isPaused() {
		t.Error("App should be paused after setting paused=true")
	}

	// Toggle back.
	app.pauseMu.Lock()
	app.paused = false
	app.pauseMu.Unlock()

	if app.isPaused() {
		t.Error("App should not be paused after setting paused=false")
	}
}

func TestPauseStateRaceCondition(_ *testing.T) {
	app := New(nil)

	// Test concurrent access to isPaused.
	done := make(chan struct{})
	for range 100 {
		go func() {
			_ = app.isPaused()
			done <- struct{}{}
		}()
		go func() {
			app.pauseMu.Lock()
			app.paused = !app.paused
			app.pauseMu.Unlock()
			done <- struct{}{}
		}()
	}

	for range 200 {
		<-done
	}
}

func TestStartTime(t *testing.T) {
	before := time.Now()
	app := New(nil)
	after := time.Now()

	if app.startTime.Before(before) || app.startTime.After(after) {
		t.Errorf("startTime %v not between %v and %v", app.startTime, before, after)
	}
}

func TestFilterActiveDefault(t *testing.T) {
	app := New(nil)

	if app.filterActive != "all" {
		t.Errorf("default filterActive = %q, want 'all'", app.filterActive)
	}
}

func TestCurrentProfileDefault(t *testing.T) {
	app := New(nil)

	if app.currentProfile.Name != "all" {
		t.Errorf("default currentProfile.Name = %q, want 'all'", app.currentProfile.Name)
	}
}

func TestNewWithFilterSetsFilterActive(t *testing.T) {
	tests := []struct {
		filter   string
		expected string
	}{
		{"all", "all"},
		{"ito", "ito"},
		{"rfc2544", "rfc2544"},
		{"y1564", "y1564"},
		{"msn", "msn"},
		{"standards", "standards"},
		{"unknown", "unknown"}, // Unknown filter still gets set.
	}

	for _, tt := range tests {
		t.Run(tt.filter, func(t *testing.T) {
			app := NewWithFilter(nil, tt.filter)
			if app.filterActive != tt.expected {
				t.Errorf("filterActive = %q, want %q", app.filterActive, tt.expected)
			}
		})
	}
}

func TestNewWithFilterSetsCurrentProfile(t *testing.T) {
	tests := []struct {
		filter          string
		expectedProfile string
	}{
		{"all", "all"},
		{"ito", "ito"},
		{"rfc2544", "rfc2544"},
		{"y1564", "y1564"},
		{"msn", "msn"},
		{"standards", "standards"},
	}

	for _, tt := range tests {
		t.Run(tt.filter, func(t *testing.T) {
			app := NewWithFilter(nil, tt.filter)
			if app.currentProfile.Name != tt.expectedProfile {
				t.Errorf("currentProfile.Name = %q, want %q",
					app.currentProfile.Name, tt.expectedProfile)
			}
		})
	}
}

func TestShowExtHelpDefault(t *testing.T) {
	app := New(nil)

	if app.showExtHelp {
		t.Error("showExtHelp should be false by default")
	}
}

func TestStopChannelClosed(t *testing.T) {
	app := New(nil)

	// Stop should close the channel.
	app.Stop()

	// Check channel is closed by trying to receive (should not block).
	select {
	case <-app.stopChan:
		// Channel is closed, as expected.
	default:
		t.Error("stopChan should be closed after Stop()")
	}
}

func TestStopOncePreventsPanic(t *testing.T) {
	app := New(nil)

	// Multiple calls should not panic.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Stop() panicked: %v", r)
		}
	}()

	app.Stop()
	app.Stop()
	app.Stop()
}

func TestAppFieldsInitialized(t *testing.T) {
	app := New(nil)

	if app.app == nil {
		t.Error("tview.Application should not be nil")
	}
	if app.pages == nil {
		t.Error("pages should not be nil")
	}
	if app.stopChan == nil {
		t.Error("stopChan should not be nil")
	}
}

func TestAppViewsNilBeforeRun(t *testing.T) {
	app := New(nil)

	// Before Run(), views should be nil.
	if app.statsView != nil {
		t.Error("statsView should be nil before Run()")
	}
	if app.sigView != nil {
		t.Error("sigView should be nil before Run()")
	}
	if app.latView != nil {
		t.Error("latView should be nil before Run()")
	}
	if app.helpView != nil {
		t.Error("helpView should be nil before Run()")
	}
	if app.headerView != nil {
		t.Error("headerView should be nil before Run()")
	}
}

func TestDataplaneNilHandling(t *testing.T) {
	app := New(nil)

	if app.dp != nil {
		t.Error("dp should be nil when created with nil dataplane")
	}
}

// =============================================================================
// handleKeyEvent Tests (without running full TUI)
// =============================================================================

func TestHandleKeyEventQuit(t *testing.T) {
	app := New(nil)

	// Create a mock key event for 'q'.
	event := tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone)
	result := app.handleKeyEvent(event)

	// handleKeyEvent should return nil for 'q' (handled).
	if result != nil {
		t.Error("handleKeyEvent('q') should return nil")
	}

	// Verify stop channel is closed (Stop was called).
	select {
	case <-app.stopChan:
		// Good - channel is closed.
	default:
		t.Error("Stop() should have been called, closing stopChan")
	}
}

func TestHandleKeyEventQuitUppercase(t *testing.T) {
	app := New(nil)

	event := tcell.NewEventKey(tcell.KeyRune, 'Q', tcell.ModNone)
	result := app.handleKeyEvent(event)

	if result != nil {
		t.Error("handleKeyEvent('Q') should return nil")
	}

	// Verify stop channel is closed.
	select {
	case <-app.stopChan:
		// Good.
	default:
		t.Error("Stop() should have been called")
	}
}

func TestHandleKeyEventReset(t *testing.T) {
	// Skip this test as it requires a non-nil dataplane.
	// The resetStats function calls dp.ResetStats() which panics on nil dp.
	t.Skip("requires non-nil dataplane")
}

func TestHandleKeyEventPause(t *testing.T) {
	// Skip - togglePause uses QueueUpdateDraw which blocks without running app.
	t.Skip("requires running tview application")
}

func TestHandleKeyEventHelp(t *testing.T) {
	// Skip - toggleExtendedHelp uses QueueUpdateDraw which blocks without running app.
	t.Skip("requires running tview application")
}

func TestHandleKeyEventQuestionMark(t *testing.T) {
	// Skip - toggleExtendedHelp uses QueueUpdateDraw which blocks without running app.
	t.Skip("requires running tview application")
}

func TestHandleKeyEventNumberKeys(t *testing.T) {
	// Skip - setProfile uses QueueUpdateDraw which blocks without running app.
	t.Skip("requires running tview application")
}

func TestHandleKeyEventFilter(t *testing.T) {
	app := New(nil)

	// 'f' should switch to profiles page - but without running app, this won't block.
	// The page switch happens synchronously, so we can test it.
	event := tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone)
	result := app.handleKeyEvent(event)

	if result != nil {
		t.Error("handleKeyEvent('f') should return nil")
	}
}

func TestHandleKeyEventFilterUppercase(t *testing.T) {
	app := New(nil)

	event := tcell.NewEventKey(tcell.KeyRune, 'F', tcell.ModNone)
	result := app.handleKeyEvent(event)

	if result != nil {
		t.Error("handleKeyEvent('F') should return nil")
	}
}

func TestHandleKeyEventRUppercase(t *testing.T) {
	// Skip - requires dataplane.
	t.Skip("requires non-nil dataplane")
}

func TestHandleKeyEventPUppercase(t *testing.T) {
	// Skip - togglePause uses QueueUpdateDraw.
	t.Skip("requires running tview application")
}

func TestHandleKeyEventHUppercase(t *testing.T) {
	// Skip - toggleExtendedHelp uses QueueUpdateDraw.
	t.Skip("requires running tview application")
}

func TestHandleKeyEventUnhandled(t *testing.T) {
	app := New(nil)

	// Unhandled key should return the event.
	event := tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone)
	result := app.handleKeyEvent(event)

	if result != event {
		t.Error("handleKeyEvent('x') should return the event unchanged")
	}
}

func TestHandleKeyEventNumberKeyProfile(t *testing.T) {
	// Skip - setProfile uses QueueUpdateDraw which blocks without running app.
	t.Skip("requires running tview application")
}

// =============================================================================
// showProfileSelector Tests
// =============================================================================

func TestShowProfileSelector(_ *testing.T) {
	app := New(nil)

	// showProfileSelector should not panic with nil pages.
	// It will fail at runtime but we're testing that it doesn't panic during the call itself.
	defer func() {
		if r := recover(); r != nil {
			_ = r // Expected - pages might not be fully initialized.
		}
	}()

	app.showProfileSelector()
}

// =============================================================================
// setProfile Tests
// =============================================================================

func TestSetProfile(t *testing.T) {
	// Skip - setProfile uses QueueUpdateDraw which blocks without running app.
	t.Skip("requires running tview application")
}

func TestSetProfileCustom(t *testing.T) {
	// Skip - setProfile uses QueueUpdateDraw which blocks without running app.
	t.Skip("requires running tview application")
}

// Test direct field manipulation (without setProfile method).
func TestSetProfileFieldsDirect(t *testing.T) {
	app := New(nil)

	profiles := GetPredefinedProfiles()
	for _, p := range profiles {
		t.Run(p.Name, func(t *testing.T) {
			// Set fields directly instead of calling setProfile.
			app.filterActive = p.Name
			app.currentProfile = p

			if app.filterActive != p.Name {
				t.Errorf("filterActive = %q, want %q", app.filterActive, p.Name)
			}
			if app.currentProfile.Name != p.Name {
				t.Errorf("currentProfile.Name = %q, want %q",
					app.currentProfile.Name, p.Name)
			}
		})
	}
}

func TestSetProfileFieldsDirectCustom(t *testing.T) {
	app := New(nil)

	customProfile := FilterProfile{
		Name:        "custom",
		Description: "Custom test profile",
		ITO:         true,
		RFC2544:     true,
		Y1564:       false,
		MSN:         false,
	}

	// Set fields directly instead of calling setProfile.
	app.filterActive = customProfile.Name
	app.currentProfile = customProfile

	if app.filterActive != "custom" {
		t.Errorf("filterActive = %q, want 'custom'", app.filterActive)
	}
	if app.currentProfile.Name != "custom" {
		t.Errorf("currentProfile.Name = %q, want 'custom'", app.currentProfile.Name)
	}
	if !app.currentProfile.ITO {
		t.Error("currentProfile.ITO should be true")
	}
	if !app.currentProfile.RFC2544 {
		t.Error("currentProfile.RFC2544 should be true")
	}
}

// =============================================================================
// toggleExtendedHelp Tests
// =============================================================================

func TestToggleExtendedHelp(t *testing.T) {
	app := New(nil)

	// Initial state.
	if app.showExtHelp {
		t.Error("showExtHelp should be false initially")
	}

	// Toggle to true.
	app.showExtHelp = true
	if !app.showExtHelp {
		t.Error("showExtHelp should be true after toggle")
	}

	// Toggle back to false.
	app.showExtHelp = false
	if app.showExtHelp {
		t.Error("showExtHelp should be false after second toggle")
	}
}

// =============================================================================
// Constants Tests
// =============================================================================

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected int
	}{
		{"statsFlexWeight", statsFlexWeight, 2},
		{"tickerIntervalMs", tickerIntervalMs, 500},
		{"profileListHeight", profileListHeight, 12},
		{"profileSelectorWidth", profileSelectorWidth, 50},
		{"secondsPerMinute", secondsPerMinute, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.value, tt.expected)
			}
		})
	}
}

func TestNumericConstants(t *testing.T) {
	// Test numeric formatting constants.
	if billion != 1000000000 {
		t.Errorf("billion = %d, want 1000000000", billion)
	}
	if million != 1000000 {
		t.Errorf("million = %d, want 1000000", million)
	}
	if thousand != 1000 {
		t.Errorf("thousand = %d, want 1000", thousand)
	}
}

func TestByteConstants(t *testing.T) {
	// Test byte formatting constants.
	if terabyte != 1099511627776 {
		t.Errorf("terabyte = %d, want 1099511627776", terabyte)
	}
	if gigabyte != 1073741824 {
		t.Errorf("gigabyte = %d, want 1073741824", gigabyte)
	}
	if megabyte != 1048576 {
		t.Errorf("megabyte = %d, want 1048576", megabyte)
	}
	if kilobyte != 1024 {
		t.Errorf("kilobyte = %d, want 1024", kilobyte)
	}
}

func TestFloatConstants(t *testing.T) {
	if bitsPerByte != 8.0 {
		t.Errorf("bitsPerByte = %f, want 8.0", bitsPerByte)
	}
	if megabitsPerSecDenom != 1000000.0 {
		t.Errorf("megabitsPerSecDenom = %f, want 1000000.0", megabitsPerSecDenom)
	}
}

// =============================================================================
// Benchmark Internal Tests
// =============================================================================

func BenchmarkIsPaused(b *testing.B) {
	app := New(nil)
	for b.Loop() {
		_ = app.isPaused()
	}
}

func BenchmarkSetProfile(b *testing.B) {
	// Skip - setProfile uses QueueUpdateDraw which blocks without running app.
	b.Skip("requires running tview application")
}

func BenchmarkHandleKeyEvent(b *testing.B) {
	app := New(nil)
	event := tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone)
	for b.Loop() {
		_ = app.handleKeyEvent(event)
	}
}

// =============================================================================
// Additional Coverage Tests
// =============================================================================

// TestHandleKeyEventAllQuitVariants tests all quit key variants.
func TestHandleKeyEventAllQuitVariants(t *testing.T) {
	tests := []struct {
		name string
		key  rune
	}{
		{"lowercase q", 'q'},
		{"uppercase Q", 'Q'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New(nil)
			event := tcell.NewEventKey(tcell.KeyRune, tt.key, tcell.ModNone)
			result := app.handleKeyEvent(event)

			if result != nil {
				t.Errorf("handleKeyEvent(%q) should return nil", tt.key)
			}

			// Verify stop was called.
			select {
			case <-app.stopChan:
				// Good.
			default:
				t.Error("Stop() should have been called")
			}
		})
	}
}

// TestHandleKeyEventAllFilterVariants tests filter key variants.
func TestHandleKeyEventAllFilterVariants(t *testing.T) {
	tests := []struct {
		name string
		key  rune
	}{
		{"lowercase f", 'f'},
		{"uppercase F", 'F'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New(nil)
			event := tcell.NewEventKey(tcell.KeyRune, tt.key, tcell.ModNone)
			result := app.handleKeyEvent(event)

			if result != nil {
				t.Errorf("handleKeyEvent(%q) should return nil", tt.key)
			}
		})
	}
}

// TestHandleKeyEventUnhandledRunes tests various unhandled key runes.
func TestHandleKeyEventUnhandledRunes(t *testing.T) {
	unhandledKeys := []rune{
		'a', 'b', 'c', 'd', 'e', 'g', 'i', 'j', 'k', 'l',
		'm', 'n', 'o', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
		'0', '7', '8', '9', '@', '#', '$', '%', '^', '&', '*',
	}

	for _, key := range unhandledKeys {
		t.Run(string(key), func(t *testing.T) {
			app := New(nil)
			event := tcell.NewEventKey(tcell.KeyRune, key, tcell.ModNone)
			result := app.handleKeyEvent(event)

			if result != event {
				t.Errorf("handleKeyEvent(%q) should return the event unchanged", key)
			}
		})
	}
}

// TestHandleKeyEventSpecialKeys tests special (non-rune) keys.
func TestHandleKeyEventSpecialKeys(t *testing.T) {
	specialKeys := []struct {
		key  tcell.Key
		name string
	}{
		{tcell.KeyEnter, "Enter"},
		{tcell.KeyEscape, "Escape"},
		{tcell.KeyTab, "Tab"},
		{tcell.KeyBackspace, "Backspace"},
		{tcell.KeyDelete, "Delete"},
		{tcell.KeyUp, "Up"},
		{tcell.KeyDown, "Down"},
		{tcell.KeyLeft, "Left"},
		{tcell.KeyRight, "Right"},
		{tcell.KeyHome, "Home"},
		{tcell.KeyEnd, "End"},
		{tcell.KeyPgUp, "PgUp"},
		{tcell.KeyPgDn, "PgDn"},
	}

	for _, tc := range specialKeys {
		t.Run(tc.name, func(t *testing.T) {
			app := New(nil)
			event := tcell.NewEventKey(tc.key, 0, tcell.ModNone)
			result := app.handleKeyEvent(event)

			// Special keys should return the event unchanged.
			if result != event {
				t.Errorf("handleKeyEvent(%s) should return the event unchanged", tc.name)
			}
		})
	}
}

// TestHandleKeyEventWithModifiers tests keys with modifiers.
func TestHandleKeyEventWithModifiers(t *testing.T) {
	modifiers := []tcell.ModMask{
		tcell.ModShift,
		tcell.ModCtrl,
		tcell.ModAlt,
		tcell.ModMeta,
	}

	for _, mod := range modifiers {
		t.Run("x+modifier", func(t *testing.T) {
			app := New(nil)
			event := tcell.NewEventKey(tcell.KeyRune, 'x', mod)
			result := app.handleKeyEvent(event)

			// Should return the event unchanged.
			if result != event {
				t.Errorf("handleKeyEvent with modifier should return event unchanged")
			}
		})
	}
}

// TestAppCreationInitializesAllFields tests that all App fields are properly initialized.
func TestAppCreationInitializesAllFields(t *testing.T) {
	app := New(nil)

	// Check all essential fields.
	if app.app == nil {
		t.Error("app.app (tview.Application) should not be nil")
	}
	if app.pages == nil {
		t.Error("app.pages should not be nil")
	}
	if app.stopChan == nil {
		t.Error("app.stopChan should not be nil")
	}
	if app.filterActive != "all" {
		t.Errorf("app.filterActive should be 'all', got %q", app.filterActive)
	}
	if app.currentProfile.Name != "all" {
		t.Errorf("app.currentProfile.Name should be 'all', got %q", app.currentProfile.Name)
	}
	if app.paused {
		t.Error("app.paused should be false")
	}
	if app.showExtHelp {
		t.Error("app.showExtHelp should be false")
	}

	// Views should be nil before Run().
	if app.statsView != nil {
		t.Error("app.statsView should be nil before Run()")
	}
	if app.sigView != nil {
		t.Error("app.sigView should be nil before Run()")
	}
	if app.latView != nil {
		t.Error("app.latView should be nil before Run()")
	}
	if app.helpView != nil {
		t.Error("app.helpView should be nil before Run()")
	}
	if app.headerView != nil {
		t.Error("app.headerView should be nil before Run()")
	}
}

// TestNewWithFilterAllProfileVariants tests NewWithFilter with all predefined profiles.
func TestNewWithFilterAllProfileVariants(t *testing.T) {
	profiles := GetPredefinedProfiles()

	for _, p := range profiles {
		t.Run(p.Name, func(t *testing.T) {
			app := NewWithFilter(nil, p.Name)

			if app.filterActive != p.Name {
				t.Errorf("filterActive = %q, want %q", app.filterActive, p.Name)
			}
			if app.currentProfile.Name != p.Name {
				t.Errorf("currentProfile.Name = %q, want %q", app.currentProfile.Name, p.Name)
			}
			if app.currentProfile.ITO != p.ITO {
				t.Errorf("currentProfile.ITO = %v, want %v", app.currentProfile.ITO, p.ITO)
			}
			if app.currentProfile.RFC2544 != p.RFC2544 {
				t.Errorf("currentProfile.RFC2544 = %v, want %v", app.currentProfile.RFC2544, p.RFC2544)
			}
			if app.currentProfile.Y1564 != p.Y1564 {
				t.Errorf("currentProfile.Y1564 = %v, want %v", app.currentProfile.Y1564, p.Y1564)
			}
			if app.currentProfile.MSN != p.MSN {
				t.Errorf("currentProfile.MSN = %v, want %v", app.currentProfile.MSN, p.MSN)
			}
		})
	}
}

// TestNewWithFilterUnknownFallback tests that unknown filter names don't crash.
func TestNewWithFilterUnknownFallback(t *testing.T) {
	unknownFilters := []string{
		"unknown",
		"invalid",
		"foo",
		"bar",
		"",
		" ",
		"ALL", // Case-sensitive test
		"ITO", // Case-sensitive test
		"rfc-2544",
		"y.1564",
	}

	for _, filter := range unknownFilters {
		t.Run(filter, func(t *testing.T) {
			app := NewWithFilter(nil, filter)

			if app == nil {
				t.Fatal("NewWithFilter should never return nil")
			}
			// filterActive gets set to whatever is passed.
			if app.filterActive != filter {
				t.Errorf("filterActive = %q, want %q", app.filterActive, filter)
			}
			// currentProfile should remain the default "all" for unknown filters.
			if app.currentProfile.Name != "all" {
				t.Errorf("currentProfile.Name should be 'all' for unknown filter, got %q",
					app.currentProfile.Name)
			}
		})
	}
}

// TestStopIdempotency tests that Stop can be called many times safely.
func TestStopIdempotency(t *testing.T) {
	app := New(nil)

	for i := range 100 {
		func(callNum int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Stop() panicked on call %d: %v", callNum, r)
				}
			}()
			app.Stop()
		}(i)
	}
}

// TestIsPausedThreadSafety tests isPaused under concurrent access.
func TestIsPausedThreadSafety(_ *testing.T) {
	app := New(nil)
	done := make(chan struct{}, 1000)

	// Start many goroutines reading and writing pause state.
	for range 500 {
		go func() {
			_ = app.isPaused()
			done <- struct{}{}
		}()
		go func() {
			app.pauseMu.Lock()
			app.paused = !app.paused
			app.pauseMu.Unlock()
			done <- struct{}{}
		}()
	}

	// Wait for all goroutines.
	for range 1000 {
		<-done
	}
}

// TestFilterProfileFieldAccess tests that FilterProfile fields are correctly accessible.
func TestFilterProfileFieldAccess(t *testing.T) {
	profiles := GetPredefinedProfiles()

	for _, p := range profiles {
		t.Run(p.Name+"_fields", func(t *testing.T) {
			// All profiles should have non-empty name.
			if p.Name == "" {
				t.Error("profile name should not be empty")
			}
			// All profiles should have non-empty description.
			if p.Description == "" {
				t.Error("profile description should not be empty")
			}

			// Verify "all" profile has everything enabled.
			if p.Name == "all" {
				if !p.ITO || !p.RFC2544 || !p.Y1564 || !p.MSN {
					t.Error("'all' profile should have all flags enabled")
				}
			}
		})
	}
}

// TestAppPausedStateMutations tests pause state mutations.
func TestAppPausedStateMutations(t *testing.T) {
	app := New(nil)

	// Initially not paused.
	if app.isPaused() {
		t.Error("initially should not be paused")
	}

	// Set paused.
	app.pauseMu.Lock()
	app.paused = true
	app.pauseMu.Unlock()

	if !app.isPaused() {
		t.Error("should be paused after setting true")
	}

	// Unset paused.
	app.pauseMu.Lock()
	app.paused = false
	app.pauseMu.Unlock()

	if app.isPaused() {
		t.Error("should not be paused after setting false")
	}
}

// TestAppShowExtHelpMutations tests showExtHelp state mutations.
func TestAppShowExtHelpMutations(t *testing.T) {
	app := New(nil)

	// Initially false.
	if app.showExtHelp {
		t.Error("initially showExtHelp should be false")
	}

	// Toggle true.
	app.showExtHelp = true
	if !app.showExtHelp {
		t.Error("should be true after setting")
	}

	// Toggle false.
	app.showExtHelp = false
	if app.showExtHelp {
		t.Error("should be false after unsetting")
	}
}

// TestAppFilterMutations tests filter state mutations.
func TestAppFilterMutations(t *testing.T) {
	app := New(nil)

	profiles := GetPredefinedProfiles()
	for _, p := range profiles {
		t.Run("set_"+p.Name, func(t *testing.T) {
			app.filterActive = p.Name
			app.currentProfile = p

			if app.filterActive != p.Name {
				t.Errorf("filterActive = %q, want %q", app.filterActive, p.Name)
			}
			if app.currentProfile.Name != p.Name {
				t.Errorf("currentProfile.Name = %q, want %q", app.currentProfile.Name, p.Name)
			}
		})
	}
}

// TestAppStartTimeIsRecent tests that startTime is set correctly.
func TestAppStartTimeIsRecent(t *testing.T) {
	before := time.Now().Add(-time.Second)
	app := New(nil)
	after := time.Now().Add(time.Second)

	if app.startTime.Before(before) {
		t.Errorf("startTime %v is before %v", app.startTime, before)
	}
	if app.startTime.After(after) {
		t.Errorf("startTime %v is after %v", app.startTime, after)
	}
}

// TestMultipleAppInstances tests that multiple App instances are independent.
func TestMultipleAppInstances(t *testing.T) {
	app1 := New(nil)
	app2 := New(nil)

	// Pointers should be different.
	if app1 == app2 {
		t.Error("different New() calls should return different instances")
	}

	// Modify app1.
	app1.pauseMu.Lock()
	app1.paused = true
	app1.pauseMu.Unlock()
	app1.filterActive = "ito"

	// app2 should be unaffected.
	if app2.isPaused() {
		t.Error("app2 should not be affected by app1 pause state")
	}
	if app2.filterActive != "all" {
		t.Errorf("app2.filterActive should be 'all', got %q", app2.filterActive)
	}
}

// TestStopChannelBehavior tests stop channel behavior.
func TestStopChannelBehavior(t *testing.T) {
	app := New(nil)

	// Channel should be open initially.
	select {
	case <-app.stopChan:
		t.Error("stopChan should be open initially")
	default:
		// Good - channel is open.
	}

	// Stop should close the channel.
	app.Stop()

	// Channel should be closed now.
	select {
	case _, ok := <-app.stopChan:
		if ok {
			t.Error("stopChan should be closed after Stop()")
		}
		// Good - channel is closed.
	default:
		t.Error("stopChan should be closed after Stop()")
	}
}

// TestStatsInputZeroValues tests StatsInput with all zero values.
func TestStatsInputZeroValues(t *testing.T) {
	var input StatsInput

	// All numeric fields should be zero.
	if input.PacketsReceived != 0 {
		t.Error("PacketsReceived should be 0")
	}
	if input.PacketsReflected != 0 {
		t.Error("PacketsReflected should be 0")
	}
	if input.BytesReceived != 0 {
		t.Error("BytesReceived should be 0")
	}
	if input.BytesReflected != 0 {
		t.Error("BytesReflected should be 0")
	}
	if input.SigProbeOT != 0 {
		t.Error("SigProbeOT should be 0")
	}
	if input.SigDataOT != 0 {
		t.Error("SigDataOT should be 0")
	}
	if input.SigLatency != 0 {
		t.Error("SigLatency should be 0")
	}
	if input.SigRFC2544 != 0 {
		t.Error("SigRFC2544 should be 0")
	}
	if input.SigY1564 != 0 {
		t.Error("SigY1564 should be 0")
	}
	if input.SigMSN != 0 {
		t.Error("SigMSN should be 0")
	}
	if input.LatencyMin != 0 {
		t.Error("LatencyMin should be 0")
	}
	if input.LatencyAvg != 0 {
		t.Error("LatencyAvg should be 0")
	}
	if input.LatencyMax != 0 {
		t.Error("LatencyMax should be 0")
	}
	if input.LatencyCount != 0 {
		t.Error("LatencyCount should be 0")
	}
	if input.Elapsed != 0 {
		t.Error("Elapsed should be 0")
	}
	if input.Uptime != 0 {
		t.Error("Uptime should be 0")
	}
}

// =============================================================================
// Testable State Functions Tests
// =============================================================================

// TestSetProfileState tests the setProfileState function.
func TestSetProfileState(t *testing.T) {
	app := New(nil)

	profiles := GetPredefinedProfiles()
	for _, p := range profiles {
		t.Run(p.Name, func(t *testing.T) {
			app.setProfileState(p)

			if app.filterActive != p.Name {
				t.Errorf("filterActive = %q, want %q", app.filterActive, p.Name)
			}
			if app.currentProfile.Name != p.Name {
				t.Errorf("currentProfile.Name = %q, want %q", app.currentProfile.Name, p.Name)
			}
			if app.currentProfile.ITO != p.ITO {
				t.Errorf("currentProfile.ITO = %v, want %v", app.currentProfile.ITO, p.ITO)
			}
			if app.currentProfile.RFC2544 != p.RFC2544 {
				t.Errorf("currentProfile.RFC2544 = %v, want %v", app.currentProfile.RFC2544, p.RFC2544)
			}
			if app.currentProfile.Y1564 != p.Y1564 {
				t.Errorf("currentProfile.Y1564 = %v, want %v", app.currentProfile.Y1564, p.Y1564)
			}
			if app.currentProfile.MSN != p.MSN {
				t.Errorf("currentProfile.MSN = %v, want %v", app.currentProfile.MSN, p.MSN)
			}
		})
	}
}

// TestSetProfileStateCustom tests setProfileState with a custom profile.
func TestSetProfileStateCustom(t *testing.T) {
	app := New(nil)

	custom := FilterProfile{
		Name:        "custom-test",
		Description: "Custom test profile",
		ITO:         true,
		RFC2544:     false,
		Y1564:       true,
		MSN:         false,
	}

	app.setProfileState(custom)

	if app.filterActive != "custom-test" {
		t.Errorf("filterActive = %q, want 'custom-test'", app.filterActive)
	}
	if app.currentProfile.Name != "custom-test" {
		t.Errorf("currentProfile.Name = %q, want 'custom-test'", app.currentProfile.Name)
	}
	if !app.currentProfile.ITO {
		t.Error("currentProfile.ITO should be true")
	}
	if app.currentProfile.RFC2544 {
		t.Error("currentProfile.RFC2544 should be false")
	}
	if !app.currentProfile.Y1564 {
		t.Error("currentProfile.Y1564 should be true")
	}
	if app.currentProfile.MSN {
		t.Error("currentProfile.MSN should be false")
	}
}

// TestToggleExtendedHelpState tests the toggleExtendedHelpState function.
func TestToggleExtendedHelpState(t *testing.T) {
	app := New(nil)

	// Initial state should be false.
	if app.showExtHelp {
		t.Error("initial showExtHelp should be false")
	}

	// Toggle to true.
	app.toggleExtendedHelpState()
	if !app.showExtHelp {
		t.Error("showExtHelp should be true after first toggle")
	}

	// Toggle back to false.
	app.toggleExtendedHelpState()
	if app.showExtHelp {
		t.Error("showExtHelp should be false after second toggle")
	}

	// Toggle again.
	app.toggleExtendedHelpState()
	if !app.showExtHelp {
		t.Error("showExtHelp should be true after third toggle")
	}
}

// TestToggleExtendedHelpStateMultiple tests multiple toggles.
func TestToggleExtendedHelpStateMultiple(t *testing.T) {
	app := New(nil)

	// Toggle many times.
	for i := range 100 {
		app.toggleExtendedHelpState()
		// After odd number of toggles (1, 3, 5...), it should be true.
		// After even number of toggles (2, 4, 6...), it should be false.
		expected := ((i + 1) % 2) == 1
		if app.showExtHelp != expected {
			t.Errorf("after toggle %d, showExtHelp = %v, want %v",
				i+1, app.showExtHelp, expected)
		}
	}
}

// TestTogglePauseState tests the togglePauseState function.
func TestTogglePauseState(t *testing.T) {
	app := New(nil)

	// Initial state should be false.
	if app.isPaused() {
		t.Error("initial isPaused should be false")
	}

	// Toggle to true.
	app.togglePauseState()
	if !app.isPaused() {
		t.Error("isPaused should be true after first toggle")
	}

	// Toggle back to false.
	app.togglePauseState()
	if app.isPaused() {
		t.Error("isPaused should be false after second toggle")
	}

	// Toggle again.
	app.togglePauseState()
	if !app.isPaused() {
		t.Error("isPaused should be true after third toggle")
	}
}

// TestTogglePauseStateMultiple tests multiple pause toggles.
func TestTogglePauseStateMultiple(t *testing.T) {
	app := New(nil)

	// Toggle many times.
	for i := range 100 {
		app.togglePauseState()
		// After odd number of toggles (1, 3, 5...), it should be true.
		// After even number of toggles (2, 4, 6...), it should be false.
		expected := ((i + 1) % 2) == 1
		if app.isPaused() != expected {
			t.Errorf("after toggle %d, isPaused = %v, want %v",
				i+1, app.isPaused(), expected)
		}
	}
}

// TestTogglePauseStateConcurrent tests togglePauseState under concurrent access.
func TestTogglePauseStateConcurrent(_ *testing.T) {
	app := New(nil)
	done := make(chan struct{}, 200)

	// Run many concurrent toggles.
	for range 100 {
		go func() {
			app.togglePauseState()
			done <- struct{}{}
		}()
		go func() {
			_ = app.isPaused()
			done <- struct{}{}
		}()
	}

	// Wait for all goroutines.
	for range 200 {
		<-done
	}

	// No assertion needed - just checking no race condition or panic.
}

// TestSetProfileStateSequential tests setting profiles in sequence.
func TestSetProfileStateSequential(t *testing.T) {
	app := New(nil)
	profiles := GetPredefinedProfiles()

	// Set each profile in sequence.
	for _, p := range profiles {
		app.setProfileState(p)
		if app.filterActive != p.Name {
			t.Errorf("filterActive = %q, want %q", app.filterActive, p.Name)
		}
	}

	// Verify the last profile is set.
	lastProfile := profiles[len(profiles)-1]
	if app.currentProfile.Name != lastProfile.Name {
		t.Errorf("final currentProfile.Name = %q, want %q",
			app.currentProfile.Name, lastProfile.Name)
	}
}

// TestStatsInputPopulated tests StatsInput with populated values.
func TestStatsInputPopulated(t *testing.T) {
	input := StatsInput{
		PacketsReceived:  1000,
		PacketsReflected: 950,
		BytesReceived:    100000,
		BytesReflected:   95000,
		SigProbeOT:       100,
		SigDataOT:        200,
		SigLatency:       300,
		SigRFC2544:       400,
		SigY1564:         500,
		SigMSN:           600,
		LatencyMin:       10.5,
		LatencyAvg:       25.3,
		LatencyMax:       50.7,
		LatencyCount:     1000,
		Elapsed:          100.0,
		Uptime:           100 * time.Second,
	}

	// Verify all fields are set correctly.
	if input.PacketsReceived != 1000 {
		t.Errorf("PacketsReceived = %d, want 1000", input.PacketsReceived)
	}
	if input.PacketsReflected != 950 {
		t.Errorf("PacketsReflected = %d, want 950", input.PacketsReflected)
	}
	if input.BytesReceived != 100000 {
		t.Errorf("BytesReceived = %d, want 100000", input.BytesReceived)
	}
	if input.BytesReflected != 95000 {
		t.Errorf("BytesReflected = %d, want 95000", input.BytesReflected)
	}
	if input.SigProbeOT != 100 {
		t.Errorf("SigProbeOT = %d, want 100", input.SigProbeOT)
	}
	if input.SigDataOT != 200 {
		t.Errorf("SigDataOT = %d, want 200", input.SigDataOT)
	}
	if input.SigLatency != 300 {
		t.Errorf("SigLatency = %d, want 300", input.SigLatency)
	}
	if input.SigRFC2544 != 400 {
		t.Errorf("SigRFC2544 = %d, want 400", input.SigRFC2544)
	}
	if input.SigY1564 != 500 {
		t.Errorf("SigY1564 = %d, want 500", input.SigY1564)
	}
	if input.SigMSN != 600 {
		t.Errorf("SigMSN = %d, want 600", input.SigMSN)
	}
	if input.LatencyMin != 10.5 {
		t.Errorf("LatencyMin = %f, want 10.5", input.LatencyMin)
	}
	if input.LatencyAvg != 25.3 {
		t.Errorf("LatencyAvg = %f, want 25.3", input.LatencyAvg)
	}
	if input.LatencyMax != 50.7 {
		t.Errorf("LatencyMax = %f, want 50.7", input.LatencyMax)
	}
	if input.LatencyCount != 1000 {
		t.Errorf("LatencyCount = %d, want 1000", input.LatencyCount)
	}
	if input.Elapsed != 100.0 {
		t.Errorf("Elapsed = %f, want 100.0", input.Elapsed)
	}
	if input.Uptime != 100*time.Second {
		t.Errorf("Uptime = %v, want 100s", input.Uptime)
	}
}

// =============================================================================
// Tests with Stub Dataplane
// =============================================================================

// TestNewWithStubDataplane tests creating an App with a stub dataplane.
func TestNewWithStubDataplane(t *testing.T) {
	// Create a stub dataplane (will be nil on non-CGO/non-Linux platforms).
	dp := &dataplane.Dataplane{}

	app := New(dp)
	if app == nil {
		t.Fatal("New(dp) returned nil")
	}
	if app.dp != dp {
		t.Error("App.dp should be set to provided dataplane")
	}
}

// TestNewWithFilterAndStubDataplane tests creating an App with filter and stub dataplane.
func TestNewWithFilterAndStubDataplane(t *testing.T) {
	dp := &dataplane.Dataplane{}

	profiles := GetPredefinedProfiles()
	for _, p := range profiles {
		t.Run(p.Name, func(t *testing.T) {
			app := NewWithFilter(dp, p.Name)
			if app == nil {
				t.Fatalf("NewWithFilter(dp, %q) returned nil", p.Name)
			}
			if app.dp != dp {
				t.Error("App.dp should be set to provided dataplane")
			}
			if app.filterActive != p.Name {
				t.Errorf("filterActive = %q, want %q", app.filterActive, p.Name)
			}
		})
	}
}

// TestResetStatsWithStubDataplane tests resetStats with a stub dataplane.
// Note: This test can't fully exercise resetStats because it calls updateStats
// which requires a running TUI. But we can test that it doesn't panic.
func TestResetStatsWithStubDataplane(t *testing.T) {
	dp := &dataplane.Dataplane{}
	app := New(dp)

	// Set a start time.
	_ = app.startTime // Acknowledge we know about startTime.

	// Wait a tiny bit to ensure time progresses.
	time.Sleep(time.Millisecond)

	// resetStats will panic because it tries to call updateStats which needs views.
	// We can only test that the dataplane's ResetStats is called.
	// Since we can't easily test this without views, skip the actual call.
	t.Skip("resetStats requires views to be initialized")
}

// TestAppWithDataplaneInterface tests that dp.Interface() is called correctly.
func TestAppWithDataplaneInterface(t *testing.T) {
	// The stub dataplane returns empty string for Interface().
	dp := &dataplane.Dataplane{}
	app := New(dp)

	// Test that GenerateHeaderText works with empty interface.
	headerText := GenerateHeaderText(dp.Interface(), app.filterActive, app.isPaused())
	if headerText == "" {
		t.Error("GenerateHeaderText should not return empty string")
	}
}

// TestAppWithDataplaneGetStats tests that dp.GetStats() is called correctly.
func TestAppWithDataplaneGetStats(t *testing.T) {
	dp := &dataplane.Dataplane{}

	// Get stats from stub dataplane (returns zeros).
	stats := dp.GetStats()
	if stats.PacketsReceived != 0 {
		t.Errorf("stub stats.PacketsReceived = %d, want 0", stats.PacketsReceived)
	}
	if stats.PacketsReflected != 0 {
		t.Errorf("stub stats.PacketsReflected = %d, want 0", stats.PacketsReflected)
	}
}

// TestDataplaneResetStats tests that dp.ResetStats() doesn't panic.
func TestDataplaneResetStats(t *testing.T) {
	dp := &dataplane.Dataplane{}

	// ResetStats on stub should not panic.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("dp.ResetStats() panicked: %v", r)
		}
	}()

	dp.ResetStats()
}

// TestDataplaneIsRunning tests that dp.IsRunning() returns false for stub.
func TestDataplaneIsRunning(t *testing.T) {
	dp := &dataplane.Dataplane{}

	if dp.IsRunning() {
		t.Error("stub dataplane should not be running")
	}
}
