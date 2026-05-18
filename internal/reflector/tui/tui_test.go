// SPDX-License-Identifier: BUSL-1.1

package tui_test

import (
	"testing"
	"time"

	"github.com/krisarmstrong/stem/internal/reflector/tui"
)

// =============================================================================
// FilterProfile Tests
// =============================================================================

func TestGetPredefinedProfiles(t *testing.T) {
	profiles := tui.GetPredefinedProfiles()

	if len(profiles) == 0 {
		t.Fatal("GetPredefinedProfiles() returned empty slice")
	}

	// Should have at least 6 predefined profiles.
	expectedCount := 6
	if len(profiles) != expectedCount {
		t.Errorf("expected %d profiles, got %d", expectedCount, len(profiles))
	}
}

func TestGetPredefinedProfilesContent(t *testing.T) {
	profiles := tui.GetPredefinedProfiles()

	// Expected profile names in order.
	expectedNames := []string{"all", "ito", "rfc2544", "y1564", "msn", "standards"}

	for i, expected := range expectedNames {
		if i >= len(profiles) {
			t.Errorf("missing profile at index %d: expected %q", i, expected)
			continue
		}
		if profiles[i].Name != expected {
			t.Errorf("profile[%d].Name = %q, want %q", i, profiles[i].Name, expected)
		}
	}
}

func TestFilterProfileAllProfile(t *testing.T) {
	profiles := tui.GetPredefinedProfiles()
	all := profiles[0]

	if all.Name != "all" {
		t.Errorf("first profile should be 'all', got %q", all.Name)
	}
	if !all.ITO || !all.RFC2544 || !all.Y1564 || !all.MSN {
		t.Error("'all' profile should have all signature types enabled")
	}
}

func TestFilterProfileITOOnly(t *testing.T) {
	profiles := tui.GetPredefinedProfiles()

	var itoProfile *tui.FilterProfile
	for i := range profiles {
		if profiles[i].Name == "ito" {
			itoProfile = &profiles[i]
			break
		}
	}

	if itoProfile == nil {
		t.Fatal("'ito' profile not found")
	}

	if !itoProfile.ITO {
		t.Error("'ito' profile should have ITO enabled")
	}
	if itoProfile.RFC2544 || itoProfile.Y1564 || itoProfile.MSN {
		t.Error("'ito' profile should not have RFC2544, Y1564, or MSN enabled")
	}
}

func TestFilterProfileRFC2544Only(t *testing.T) {
	profiles := tui.GetPredefinedProfiles()

	var profile *tui.FilterProfile
	for i := range profiles {
		if profiles[i].Name == "rfc2544" {
			profile = &profiles[i]
			break
		}
	}

	if profile == nil {
		t.Fatal("'rfc2544' profile not found")
	}

	if !profile.RFC2544 {
		t.Error("'rfc2544' profile should have RFC2544 enabled")
	}
	if profile.ITO || profile.Y1564 || profile.MSN {
		t.Error("'rfc2544' profile should not have ITO, Y1564, or MSN enabled")
	}
}

func TestFilterProfileY1564Only(t *testing.T) {
	profiles := tui.GetPredefinedProfiles()

	var profile *tui.FilterProfile
	for i := range profiles {
		if profiles[i].Name == "y1564" {
			profile = &profiles[i]
			break
		}
	}

	if profile == nil {
		t.Fatal("'y1564' profile not found")
	}

	if !profile.Y1564 {
		t.Error("'y1564' profile should have Y1564 enabled")
	}
	if profile.ITO || profile.RFC2544 || profile.MSN {
		t.Error("'y1564' profile should not have ITO, RFC2544, or MSN enabled")
	}
}

func TestFilterProfileMSNOnly(t *testing.T) {
	profiles := tui.GetPredefinedProfiles()

	var profile *tui.FilterProfile
	for i := range profiles {
		if profiles[i].Name == "msn" {
			profile = &profiles[i]
			break
		}
	}

	if profile == nil {
		t.Fatal("'msn' profile not found")
	}

	if !profile.MSN {
		t.Error("'msn' profile should have MSN enabled")
	}
	if profile.ITO || profile.RFC2544 || profile.Y1564 {
		t.Error("'msn' profile should not have ITO, RFC2544, or Y1564 enabled")
	}
}

func TestFilterProfileStandards(t *testing.T) {
	profiles := tui.GetPredefinedProfiles()

	var profile *tui.FilterProfile
	for i := range profiles {
		if profiles[i].Name == "standards" {
			profile = &profiles[i]
			break
		}
	}

	if profile == nil {
		t.Fatal("'standards' profile not found")
	}

	if !profile.RFC2544 || !profile.Y1564 {
		t.Error("'standards' profile should have RFC2544 and Y1564 enabled")
	}
	if profile.ITO || profile.MSN {
		t.Error("'standards' profile should not have ITO or MSN enabled")
	}
}

func TestFilterProfileDescriptions(t *testing.T) {
	profiles := tui.GetPredefinedProfiles()

	for i, p := range profiles {
		if p.Description == "" {
			t.Errorf("profile[%d] (%q) has empty description", i, p.Name)
		}
	}
}

// =============================================================================
// App Constructor Tests
// =============================================================================

func TestNewApp(t *testing.T) {
	// Test creating new App with nil dataplane (should not panic).
	app := tui.New(nil)

	if app == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNewAppDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("New() panicked: %v", r)
		}
	}()

	_ = tui.New(nil)
}

func TestNewWithFilter(t *testing.T) {
	tests := []struct {
		name       string
		filterName string
	}{
		{"all filter", "all"},
		{"ito filter", "ito"},
		{"rfc2544 filter", "rfc2544"},
		{"y1564 filter", "y1564"},
		{"msn filter", "msn"},
		{"standards filter", "standards"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tui.NewWithFilter(nil, tt.filterName)
			if app == nil {
				t.Fatalf("NewWithFilter(nil, %q) returned nil", tt.filterName)
			}
		})
	}
}

func TestNewWithFilterUnknownProfile(t *testing.T) {
	// Unknown profile should still create an app (defaults to "all").
	app := tui.NewWithFilter(nil, "unknown")
	if app == nil {
		t.Fatal("NewWithFilter(nil, 'unknown') returned nil")
	}
}

func TestNewWithFilterEmpty(t *testing.T) {
	app := tui.NewWithFilter(nil, "")
	if app == nil {
		t.Fatal("NewWithFilter(nil, '') returned nil")
	}
}

// =============================================================================
// App.Stop Tests
// =============================================================================

func TestStopMethod(t *testing.T) {
	app := tui.New(nil)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Stop() panicked: %v", r)
		}
	}()

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

func TestStopMethodConcurrent(t *testing.T) {
	app := tui.New(nil)

	// Test concurrent Stop() calls don't cause issues.
	done := make(chan struct{})
	for range 10 {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Concurrent Stop() panicked: %v", r)
				}
			}()
			app.Stop()
			done <- struct{}{}
		}()
	}

	// Wait for all goroutines.
	for range 10 {
		<-done
	}
}

// =============================================================================
// FormatNumber Tests
// =============================================================================

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"zero", 0, "0"},
		{"single digit", 1, "1"},
		{"double digit", 42, "42"},
		{"triple digit", 999, "999"},
		{"thousand boundary", 1000, "1.00K"},
		{"thousands", 1500, "1.50K"},
		{"large thousands", 999999, "1000.00K"},
		{"million boundary", 1000000, "1.00M"},
		{"millions", 5500000, "5.50M"},
		{"large millions", 999999999, "1000.00M"},
		{"billion boundary", 1000000000, "1.00B"},
		{"billions", 2500000000, "2.50B"},
		{"large billions", 100000000000, "100.00B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tui.FormatNumber(tt.input)
			if result != tt.expected {
				t.Errorf("FormatNumber(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatNumberBoundaries(t *testing.T) {
	// Test exact boundary values.
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"just under thousand", 999, "999"},
		{"exactly thousand", 1000, "1.00K"},
		{"just over thousand", 1001, "1.00K"},
		{"just under million", 999999, "1000.00K"},
		{"exactly million", 1000000, "1.00M"},
		{"just over million", 1000001, "1.00M"},
		{"just under billion", 999999999, "1000.00M"},
		{"exactly billion", 1000000000, "1.00B"},
		{"just over billion", 1000000001, "1.00B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tui.FormatNumber(tt.input)
			if result != tt.expected {
				t.Errorf("FormatNumber(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatNumberMaxUint64(t *testing.T) {
	// Test with max uint64 value.
	maxVal := ^uint64(0) // 18446744073709551615
	result := tui.FormatNumber(maxVal)

	// Should format as billions (very large number).
	if result == "" {
		t.Error("FormatNumber(maxUint64) returned empty string")
	}
	// Should contain 'B' suffix.
	if len(result) < 2 || result[len(result)-1] != 'B' {
		t.Errorf("FormatNumber(maxUint64) = %q, expected to end with 'B'", result)
	}
}

// =============================================================================
// FormatBytes Tests
// =============================================================================

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"zero bytes", 0, "0 B"},
		{"single byte", 1, "1 B"},
		{"few bytes", 100, "100 B"},
		{"just under KB", 1023, "1023 B"},
		{"exactly KB", 1024, "1.00 KB"},
		{"kilobytes", 2048, "2.00 KB"},
		{"large KB", 512000, "500.00 KB"},
		{"exactly MB", 1048576, "1.00 MB"},
		{"megabytes", 5242880, "5.00 MB"},
		{"large MB", 536870912, "512.00 MB"},
		{"exactly GB", 1073741824, "1.00 GB"},
		{"gigabytes", 2147483648, "2.00 GB"},
		{"large GB", 549755813888, "512.00 GB"},
		{"exactly TB", 1099511627776, "1.00 TB"},
		{"terabytes", 2199023255552, "2.00 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tui.FormatBytes(tt.input)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatBytesBoundaries(t *testing.T) {
	// Test exact boundary values with binary units (1024-based).
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"just under KB", 1023, "1023 B"},
		{"exactly KB", 1024, "1.00 KB"},
		{"just over KB", 1025, "1.00 KB"},
		{"just under MB", 1048575, "1024.00 KB"},
		{"exactly MB", 1048576, "1.00 MB"},
		{"just over MB", 1048577, "1.00 MB"},
		{"just under GB", 1073741823, "1024.00 MB"},
		{"exactly GB", 1073741824, "1.00 GB"},
		{"just over GB", 1073741825, "1.00 GB"},
		{"just under TB", 1099511627775, "1024.00 GB"},
		{"exactly TB", 1099511627776, "1.00 TB"},
		{"just over TB", 1099511627777, "1.00 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tui.FormatBytes(tt.input)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatBytesMaxUint64(t *testing.T) {
	maxVal := ^uint64(0)
	result := tui.FormatBytes(maxVal)

	if result == "" {
		t.Error("FormatBytes(maxUint64) returned empty string")
	}
	// Should contain 'TB' suffix.
	if len(result) < 3 || result[len(result)-2:] != "TB" {
		t.Errorf("FormatBytes(maxUint64) = %q, expected to end with 'TB'", result)
	}
}

// =============================================================================
// FormatDuration Tests
// =============================================================================

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		{"zero", 0, "0s"},
		{"one second", time.Second, "1s"},
		{"few seconds", 5 * time.Second, "5s"},
		{"59 seconds", 59 * time.Second, "59s"},
		{"one minute", time.Minute, "1m 0s"},
		{"minute and seconds", time.Minute + 30*time.Second, "1m 30s"},
		{"multiple minutes", 5 * time.Minute, "5m 0s"},
		{"59 minutes", 59 * time.Minute, "59m 0s"},
		{"one hour", time.Hour, "1h 0m 0s"},
		{"hour and minutes", time.Hour + 30*time.Minute, "1h 30m 0s"},
		{"hour min sec", time.Hour + 5*time.Minute + 10*time.Second, "1h 5m 10s"},
		{"multiple hours", 5 * time.Hour, "5h 0m 0s"},
		{"large hours", 100 * time.Hour, "100h 0m 0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tui.FormatDuration(tt.input)
			if result != tt.expected {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatDurationBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		{"just under minute", 59 * time.Second, "59s"},
		{"exactly minute", 60 * time.Second, "1m 0s"},
		{"just over minute", 61 * time.Second, "1m 1s"},
		{"just under hour", 59*time.Minute + 59*time.Second, "59m 59s"},
		{"exactly hour", 60 * time.Minute, "1h 0m 0s"},
		{"just over hour", 60*time.Minute + 1*time.Second, "1h 0m 1s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tui.FormatDuration(tt.input)
			if result != tt.expected {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatDurationSubSecond(t *testing.T) {
	// Sub-second durations should show as 0s.
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		{"milliseconds", 500 * time.Millisecond, "0s"},
		{"microseconds", 500 * time.Microsecond, "0s"},
		{"nanoseconds", 500 * time.Nanosecond, "0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tui.FormatDuration(tt.input)
			if result != tt.expected {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatDurationNegative(t *testing.T) {
	// Negative duration should still work (showing 0s or negative).
	result := tui.FormatDuration(-5 * time.Second)
	// The result depends on implementation - just ensure no panic.
	if result == "" {
		t.Error("FormatDuration(-5s) returned empty string")
	}
}

func TestFormatDurationVeryLarge(t *testing.T) {
	// Very large duration (days worth of hours).
	duration := 1000 * time.Hour
	result := tui.FormatDuration(duration)

	if result == "" {
		t.Error("FormatDuration(1000h) returned empty string")
	}
	// Should contain 'h'.
	found := false
	for _, c := range result {
		if c == 'h' {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("FormatDuration(1000h) = %q, expected to contain 'h'", result)
	}
}

// =============================================================================
// FilterProfile Struct Tests
// =============================================================================

func TestFilterProfileStructFields(t *testing.T) {
	// Test that FilterProfile struct has all expected fields.
	profile := tui.FilterProfile{
		Name:        "test",
		Description: "Test profile",
		ITO:         true,
		RFC2544:     false,
		Y1564:       true,
		MSN:         false,
	}

	if profile.Name != "test" {
		t.Errorf("FilterProfile.Name = %q, want 'test'", profile.Name)
	}
	if profile.Description != "Test profile" {
		t.Errorf("FilterProfile.Description = %q, want 'Test profile'", profile.Description)
	}
	if !profile.ITO {
		t.Error("FilterProfile.ITO should be true")
	}
	if profile.RFC2544 {
		t.Error("FilterProfile.RFC2544 should be false")
	}
	if !profile.Y1564 {
		t.Error("FilterProfile.Y1564 should be true")
	}
	if profile.MSN {
		t.Error("FilterProfile.MSN should be false")
	}
}

func TestFilterProfileZeroValue(t *testing.T) {
	var profile tui.FilterProfile

	if profile.Name != "" {
		t.Errorf("zero FilterProfile.Name = %q, want empty", profile.Name)
	}
	if profile.Description != "" {
		t.Errorf("zero FilterProfile.Description = %q, want empty", profile.Description)
	}
	if profile.ITO || profile.RFC2544 || profile.Y1564 || profile.MSN {
		t.Error("zero FilterProfile should have all flags false")
	}
}

// =============================================================================
// Profile Search Tests
// =============================================================================

func TestFindProfileByName(t *testing.T) {
	profiles := tui.GetPredefinedProfiles()
	searchNames := []string{"all", "ito", "rfc2544", "y1564", "msn", "standards"}

	for _, name := range searchNames {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, p := range profiles {
				if p.Name == name {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("profile %q not found in predefined profiles", name)
			}
		})
	}
}

func TestProfileNamesAreUnique(t *testing.T) {
	profiles := tui.GetPredefinedProfiles()
	seen := make(map[string]bool)

	for _, p := range profiles {
		if seen[p.Name] {
			t.Errorf("duplicate profile name: %q", p.Name)
		}
		seen[p.Name] = true
	}
}

func TestProfileNamesAreNonEmpty(t *testing.T) {
	profiles := tui.GetPredefinedProfiles()

	for i, p := range profiles {
		if p.Name == "" {
			t.Errorf("profile[%d] has empty name", i)
		}
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkNew(b *testing.B) {
	for b.Loop() {
		_ = tui.New(nil)
	}
}

func BenchmarkNewWithFilter(b *testing.B) {
	for b.Loop() {
		_ = tui.NewWithFilter(nil, "rfc2544")
	}
}

func BenchmarkStop(b *testing.B) {
	for b.Loop() {
		app := tui.New(nil)
		app.Stop()
	}
}

func BenchmarkGetPredefinedProfiles(b *testing.B) {
	for b.Loop() {
		_ = tui.GetPredefinedProfiles()
	}
}

func BenchmarkFormatNumber(b *testing.B) {
	values := []uint64{0, 100, 1000, 1000000, 1000000000}
	for b.Loop() {
		for _, v := range values {
			_ = tui.FormatNumber(v)
		}
	}
}

func BenchmarkFormatBytes(b *testing.B) {
	values := []uint64{0, 1024, 1048576, 1073741824, 1099511627776}
	for b.Loop() {
		for _, v := range values {
			_ = tui.FormatBytes(v)
		}
	}
}

func BenchmarkFormatDuration(b *testing.B) {
	durations := []time.Duration{
		0, time.Second, time.Minute, time.Hour, 100 * time.Hour,
	}
	for b.Loop() {
		for _, d := range durations {
			_ = tui.FormatDuration(d)
		}
	}
}

func BenchmarkGenerateHeaderText(b *testing.B) {
	for b.Loop() {
		_ = tui.GenerateHeaderText("eth0", "rfc2544", false)
	}
}

func BenchmarkGenerateHelpText(b *testing.B) {
	for b.Loop() {
		_ = tui.GenerateHelpText(false, true)
	}
}

func BenchmarkGenerateStatsText(b *testing.B) {
	input := tui.StatsInput{
		PacketsReceived:  1000000,
		PacketsReflected: 950000,
		BytesReceived:    1073741824,
		BytesReflected:   1000000000,
		SigProbeOT:       10000,
		SigDataOT:        0,
		SigLatency:       0,
		SigRFC2544:       0,
		SigY1564:         0,
		SigMSN:           0,
		LatencyMin:       0,
		LatencyAvg:       0,
		LatencyMax:       0,
		LatencyCount:     0,
		Elapsed:          100.0,
		Uptime:           100 * time.Second,
	}
	for b.Loop() {
		_ = tui.GenerateStatsText(input)
	}
}

func BenchmarkGenerateSignatureText(b *testing.B) {
	input := tui.StatsInput{
		PacketsReceived:  0,
		PacketsReflected: 0,
		BytesReceived:    0,
		BytesReflected:   0,
		SigProbeOT:       10000,
		SigDataOT:        20000,
		SigLatency:       30000,
		SigRFC2544:       40000,
		SigY1564:         50000,
		SigMSN:           60000,
		LatencyMin:       0,
		LatencyAvg:       0,
		LatencyMax:       0,
		LatencyCount:     0,
		Elapsed:          0,
		Uptime:           0,
	}
	for b.Loop() {
		_ = tui.GenerateSignatureText(input)
	}
}

func BenchmarkGenerateLatencyText(b *testing.B) {
	input := tui.StatsInput{
		PacketsReceived:  0,
		PacketsReflected: 0,
		BytesReceived:    0,
		BytesReflected:   0,
		SigProbeOT:       0,
		SigDataOT:        0,
		SigLatency:       0,
		SigRFC2544:       0,
		SigY1564:         0,
		SigMSN:           0,
		LatencyMin:       10.5,
		LatencyAvg:       25.3,
		LatencyMax:       50.7,
		LatencyCount:     1000,
		Elapsed:          0,
		Uptime:           0,
	}
	for b.Loop() {
		_ = tui.GenerateLatencyText(input)
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestFormatNumberSpecialValues(t *testing.T) {
	// Test specific edge cases.
	tests := []struct {
		name  string
		input uint64
	}{
		{"power of 10: 10", 10},
		{"power of 10: 100", 100},
		{"power of 10: 1000", 1000},
		{"power of 10: 10000", 10000},
		{"power of 10: 100000", 100000},
		{"power of 10: 1000000", 1000000},
		{"power of 10: 10000000", 10000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tui.FormatNumber(tt.input)
			if result == "" {
				t.Errorf("FormatNumber(%d) returned empty string", tt.input)
			}
		})
	}
}

func TestFormatBytesSpecialValues(t *testing.T) {
	// Test powers of 2.
	tests := []struct {
		name  string
		input uint64
	}{
		{"2^10 (1 KB)", 1024},
		{"2^20 (1 MB)", 1048576},
		{"2^30 (1 GB)", 1073741824},
		{"2^40 (1 TB)", 1099511627776},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tui.FormatBytes(tt.input)
			if result == "" {
				t.Errorf("FormatBytes(%d) returned empty string", tt.input)
			}
		})
	}
}

// =============================================================================
// GenerateHeaderText Tests
// =============================================================================

func TestGenerateHeaderText(t *testing.T) {
	tests := []struct {
		name          string
		interfaceName string
		filterActive  string
		paused        bool
		wantContains  []string
	}{
		{
			name:          "running with all filter",
			interfaceName: "eth0",
			filterActive:  "all",
			paused:        false,
			wantContains:  []string{"eth0", "RUNNING", "MSN Reflector"},
		},
		{
			name:          "paused with all filter",
			interfaceName: "eth0",
			filterActive:  "all",
			paused:        true,
			wantContains:  []string{"eth0", "PAUSED", "MSN Reflector"},
		},
		{
			name:          "running with ito filter",
			interfaceName: "eth1",
			filterActive:  "ito",
			paused:        false,
			wantContains:  []string{"eth1", "RUNNING", "Filter:", "ito"},
		},
		{
			name:          "empty interface",
			interfaceName: "",
			filterActive:  "all",
			paused:        false,
			wantContains:  []string{"RUNNING"},
		},
		{
			name:          "rfc2544 filter",
			interfaceName: "enp0s3",
			filterActive:  "rfc2544",
			paused:        false,
			wantContains:  []string{"enp0s3", "Filter:", "rfc2544"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tui.GenerateHeaderText(tt.interfaceName, tt.filterActive, tt.paused)

			for _, want := range tt.wantContains {
				if !containsString(result, want) {
					t.Errorf("GenerateHeaderText() = %q, want to contain %q", result, want)
				}
			}
		})
	}
}

func TestGenerateHeaderTextFilterVisibility(t *testing.T) {
	// "all" and empty filter should not show filter text.
	allResult := tui.GenerateHeaderText("eth0", "all", false)
	if containsString(allResult, "Filter:") {
		t.Error("'all' filter should not show 'Filter:' text")
	}

	emptyResult := tui.GenerateHeaderText("eth0", "", false)
	if containsString(emptyResult, "Filter:") {
		t.Error("empty filter should not show 'Filter:' text")
	}

	// Other filters should show filter text.
	itoResult := tui.GenerateHeaderText("eth0", "ito", false)
	if !containsString(itoResult, "Filter:") {
		t.Error("'ito' filter should show 'Filter:' text")
	}
}

// =============================================================================
// GenerateHelpText Tests
// =============================================================================

func TestGenerateHelpText(t *testing.T) {
	tests := []struct {
		name         string
		paused       bool
		extendedHelp bool
		wantContains []string
	}{
		{
			name:         "compact not paused",
			paused:       false,
			extendedHelp: false,
			wantContains: []string{"quit", "reset", "pause", "filter"},
		},
		{
			name:         "compact paused",
			paused:       true,
			extendedHelp: false,
			wantContains: []string{"quit", "reset", "resume", "filter"},
		},
		{
			name:         "extended not paused",
			paused:       false,
			extendedHelp: true,
			wantContains: []string{"quit", "reset", "pause", "filter", "1-6", "toggle help"},
		},
		{
			name:         "extended paused",
			paused:       true,
			extendedHelp: true,
			wantContains: []string{"quit", "reset", "resume", "filter", "1-6", "toggle help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tui.GenerateHelpText(tt.paused, tt.extendedHelp)

			for _, want := range tt.wantContains {
				if !containsString(result, want) {
					t.Errorf("GenerateHelpText(%v, %v) = %q, want to contain %q",
						tt.paused, tt.extendedHelp, result, want)
				}
			}
		})
	}
}

func TestGenerateHelpTextPauseResumeToggle(t *testing.T) {
	notPaused := tui.GenerateHelpText(false, false)
	if !containsString(notPaused, "pause") || containsString(notPaused, "resume") {
		t.Error("not paused should show 'pause' not 'resume'")
	}

	paused := tui.GenerateHelpText(true, false)
	// Note: "pause" might appear in other contexts (like "Pause/Resume"), but "resume" should be present when paused.
	if !containsString(paused, "resume") {
		t.Error("paused should show 'resume'")
	}
}

// =============================================================================
// StatsInput and Generate*Text Tests
// =============================================================================

func TestStatsInputZeroValue(t *testing.T) {
	var input tui.StatsInput

	// All fields should be zero.
	if input.PacketsReceived != 0 {
		t.Error("zero StatsInput.PacketsReceived should be 0")
	}
	if input.Elapsed != 0 {
		t.Error("zero StatsInput.Elapsed should be 0")
	}
	if input.Uptime != 0 {
		t.Error("zero StatsInput.Uptime should be 0")
	}
}

func TestGenerateStatsText(t *testing.T) {
	input := tui.StatsInput{
		PacketsReceived:  1000,
		PacketsReflected: 950,
		BytesReceived:    1048576,
		BytesReflected:   1000000,
		SigProbeOT:       0,
		SigDataOT:        0,
		SigLatency:       0,
		SigRFC2544:       0,
		SigY1564:         0,
		SigMSN:           0,
		LatencyMin:       0,
		LatencyAvg:       0,
		LatencyMax:       0,
		LatencyCount:     0,
		Elapsed:          10.0,
		Uptime:           10 * time.Second,
	}

	result := tui.GenerateStatsText(input)

	// Should contain packet stats.
	if !containsString(result, "RX Packets") {
		t.Error("should contain 'RX Packets'")
	}
	if !containsString(result, "TX Packets") {
		t.Error("should contain 'TX Packets'")
	}
	if !containsString(result, "RX Bytes") {
		t.Error("should contain 'RX Bytes'")
	}
	if !containsString(result, "TX Bytes") {
		t.Error("should contain 'TX Bytes'")
	}
	if !containsString(result, "Rate") {
		t.Error("should contain 'Rate'")
	}
	if !containsString(result, "Throughput") {
		t.Error("should contain 'Throughput'")
	}
	if !containsString(result, "Uptime") {
		t.Error("should contain 'Uptime'")
	}
}

func TestGenerateStatsTextZeroElapsed(t *testing.T) {
	input := tui.StatsInput{
		PacketsReceived:  1000,
		PacketsReflected: 950,
		BytesReceived:    1048576,
		BytesReflected:   1000000,
		SigProbeOT:       0,
		SigDataOT:        0,
		SigLatency:       0,
		SigRFC2544:       0,
		SigY1564:         0,
		SigMSN:           0,
		LatencyMin:       0,
		LatencyAvg:       0,
		LatencyMax:       0,
		LatencyCount:     0,
		Elapsed:          0, // Zero elapsed time.
		Uptime:           0,
	}

	result := tui.GenerateStatsText(input)

	// Should still generate text without division by zero.
	if result == "" {
		t.Error("GenerateStatsText with zero elapsed should not return empty string")
	}
	if !containsString(result, "0 pps") {
		t.Error("zero elapsed should show 0 pps")
	}
}

func TestGenerateStatsTextRateCalculation(t *testing.T) {
	input := tui.StatsInput{
		PacketsReceived:  0,
		PacketsReflected: 1000,
		BytesReceived:    0,
		BytesReflected:   1000000,
		SigProbeOT:       0,
		SigDataOT:        0,
		SigLatency:       0,
		SigRFC2544:       0,
		SigY1564:         0,
		SigMSN:           0,
		LatencyMin:       0,
		LatencyAvg:       0,
		LatencyMax:       0,
		LatencyCount:     0,
		Elapsed:          1.0, // 1 second.
		Uptime:           time.Second,
	}

	result := tui.GenerateStatsText(input)

	// Should show 1000 pps.
	if !containsString(result, "1000 pps") {
		t.Errorf("should show 1000 pps, got: %s", result)
	}
}

func TestGenerateSignatureText(t *testing.T) {
	input := tui.StatsInput{
		PacketsReceived:  0,
		PacketsReflected: 0,
		BytesReceived:    0,
		BytesReflected:   0,
		SigProbeOT:       100,
		SigDataOT:        200,
		SigLatency:       300,
		SigRFC2544:       400,
		SigY1564:         500,
		SigMSN:           600,
		LatencyMin:       0,
		LatencyAvg:       0,
		LatencyMax:       0,
		LatencyCount:     0,
		Elapsed:          0,
		Uptime:           0,
	}

	result := tui.GenerateSignatureText(input)

	// Should contain all signature types.
	if !containsString(result, "PROBEOT") {
		t.Error("should contain 'PROBEOT'")
	}
	if !containsString(result, "DATA:OT") {
		t.Error("should contain 'DATA:OT'")
	}
	if !containsString(result, "LATENCY") {
		t.Error("should contain 'LATENCY'")
	}
	if !containsString(result, "RFC2544") {
		t.Error("should contain 'RFC2544'")
	}
	if !containsString(result, "Y.1564") {
		t.Error("should contain 'Y.1564'")
	}
	if !containsString(result, "MSN") {
		t.Error("should contain 'MSN'")
	}
}

func TestGenerateSignatureTextZeroValues(t *testing.T) {
	input := tui.StatsInput{
		PacketsReceived:  0,
		PacketsReflected: 0,
		BytesReceived:    0,
		BytesReflected:   0,
		SigProbeOT:       0,
		SigDataOT:        0,
		SigLatency:       0,
		SigRFC2544:       0,
		SigY1564:         0,
		SigMSN:           0,
		LatencyMin:       0,
		LatencyAvg:       0,
		LatencyMax:       0,
		LatencyCount:     0,
		Elapsed:          0,
		Uptime:           0,
	}

	result := tui.GenerateSignatureText(input)

	// Should still generate text with zero values.
	if result == "" {
		t.Error("GenerateSignatureText with zero values should not return empty string")
	}
	// All counts should show "0".
	if !containsString(result, "0") {
		t.Error("zero values should show '0'")
	}
}

func TestGenerateLatencyText(t *testing.T) {
	tests := []struct {
		name         string
		input        tui.StatsInput
		wantContains []string
	}{
		{
			name: "with latency data",
			input: tui.StatsInput{
				PacketsReceived:  0,
				PacketsReflected: 0,
				BytesReceived:    0,
				BytesReflected:   0,
				SigProbeOT:       0,
				SigDataOT:        0,
				SigLatency:       0,
				SigRFC2544:       0,
				SigY1564:         0,
				SigMSN:           0,
				LatencyMin:       10.5,
				LatencyAvg:       25.3,
				LatencyMax:       50.7,
				LatencyCount:     1000,
				Elapsed:          0,
				Uptime:           0,
			},
			wantContains: []string{"Min:", "Avg:", "Max:", "Count:", "10.50", "25.30", "50.70"},
		},
		{
			name: "no latency data",
			input: tui.StatsInput{
				PacketsReceived:  0,
				PacketsReflected: 0,
				BytesReceived:    0,
				BytesReflected:   0,
				SigProbeOT:       0,
				SigDataOT:        0,
				SigLatency:       0,
				SigRFC2544:       0,
				SigY1564:         0,
				SigMSN:           0,
				LatencyMin:       0,
				LatencyAvg:       0,
				LatencyMax:       0,
				LatencyCount:     0,
				Elapsed:          0,
				Uptime:           0,
			},
			wantContains: []string{"No latency data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tui.GenerateLatencyText(tt.input)

			for _, want := range tt.wantContains {
				if !containsString(result, want) {
					t.Errorf("GenerateLatencyText() = %q, want to contain %q", result, want)
				}
			}
		})
	}
}

func TestGenerateLatencyTextZeroLatency(t *testing.T) {
	input := tui.StatsInput{
		PacketsReceived:  0,
		PacketsReflected: 0,
		BytesReceived:    0,
		BytesReflected:   0,
		SigProbeOT:       0,
		SigDataOT:        0,
		SigLatency:       0,
		SigRFC2544:       0,
		SigY1564:         0,
		SigMSN:           0,
		LatencyMin:       0,
		LatencyAvg:       0,
		LatencyMax:       0,
		LatencyCount:     0,
		Elapsed:          0,
		Uptime:           0,
	}

	result := tui.GenerateLatencyText(input)

	// With zero count, should show "No latency data".
	if !containsString(result, "No latency data") {
		t.Errorf("zero latency count should show 'No latency data', got: %s", result)
	}
}

func TestGenerateLatencyTextWithData(t *testing.T) {
	input := tui.StatsInput{
		PacketsReceived:  0,
		PacketsReflected: 0,
		BytesReceived:    0,
		BytesReflected:   0,
		SigProbeOT:       0,
		SigDataOT:        0,
		SigLatency:       0,
		SigRFC2544:       0,
		SigY1564:         0,
		SigMSN:           0,
		LatencyMin:       1.23,
		LatencyAvg:       4.56,
		LatencyMax:       7.89,
		LatencyCount:     100,
		Elapsed:          0,
		Uptime:           0,
	}

	result := tui.GenerateLatencyText(input)

	// Should NOT show "No latency data".
	if containsString(result, "No latency data") {
		t.Error("with latency count > 0, should not show 'No latency data'")
	}
	// Should contain Min, Avg, Max, Count.
	if !containsString(result, "Min:") {
		t.Error("should contain 'Min:'")
	}
	if !containsString(result, "Avg:") {
		t.Error("should contain 'Avg:'")
	}
	if !containsString(result, "Max:") {
		t.Error("should contain 'Max:'")
	}
	if !containsString(result, "Count:") {
		t.Error("should contain 'Count:'")
	}
}

// =============================================================================
// ParseKeyAction Tests
// =============================================================================

func TestParseKeyAction(t *testing.T) {
	tests := []struct {
		name     string
		input    rune
		expected tui.KeyAction
	}{
		// Quit keys
		{"lowercase q", 'q', tui.KeyActionQuit},
		{"uppercase Q", 'Q', tui.KeyActionQuit},

		// Reset keys
		{"lowercase r", 'r', tui.KeyActionReset},
		{"uppercase R", 'R', tui.KeyActionReset},

		// Pause keys
		{"lowercase p", 'p', tui.KeyActionTogglePause},
		{"uppercase P", 'P', tui.KeyActionTogglePause},

		// Profile selector keys
		{"lowercase f", 'f', tui.KeyActionShowProfiles},
		{"uppercase F", 'F', tui.KeyActionShowProfiles},

		// Help keys
		{"lowercase h", 'h', tui.KeyActionToggleHelp},
		{"uppercase H", 'H', tui.KeyActionToggleHelp},
		{"question mark", '?', tui.KeyActionToggleHelp},

		// Profile number keys
		{"profile 1", '1', tui.KeyActionSetProfile1},
		{"profile 2", '2', tui.KeyActionSetProfile2},
		{"profile 3", '3', tui.KeyActionSetProfile3},
		{"profile 4", '4', tui.KeyActionSetProfile4},
		{"profile 5", '5', tui.KeyActionSetProfile5},
		{"profile 6", '6', tui.KeyActionSetProfile6},

		// Unhandled keys
		{"unhandled a", 'a', tui.KeyActionNone},
		{"unhandled b", 'b', tui.KeyActionNone},
		{"unhandled z", 'z', tui.KeyActionNone},
		{"unhandled 0", '0', tui.KeyActionNone},
		{"unhandled 7", '7', tui.KeyActionNone},
		{"unhandled 8", '8', tui.KeyActionNone},
		{"unhandled 9", '9', tui.KeyActionNone},
		{"unhandled space", ' ', tui.KeyActionNone},
		{"unhandled tab", '\t', tui.KeyActionNone},
		{"unhandled newline", '\n', tui.KeyActionNone},
		{"unhandled @", '@', tui.KeyActionNone},
		{"unhandled #", '#', tui.KeyActionNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tui.ParseKeyAction(tt.input)
			if result != tt.expected {
				t.Errorf("ParseKeyAction(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseKeyActionAllActions(t *testing.T) {
	// Verify all KeyAction values can be produced by ParseKeyAction.
	actions := map[tui.KeyAction]bool{
		tui.KeyActionQuit:         false,
		tui.KeyActionReset:        false,
		tui.KeyActionTogglePause:  false,
		tui.KeyActionShowProfiles: false,
		tui.KeyActionToggleHelp:   false,
		tui.KeyActionSetProfile1:  false,
		tui.KeyActionSetProfile2:  false,
		tui.KeyActionSetProfile3:  false,
		tui.KeyActionSetProfile4:  false,
		tui.KeyActionSetProfile5:  false,
		tui.KeyActionSetProfile6:  false,
		tui.KeyActionNone:         false,
	}

	// Test all printable ASCII characters.
	for r := range 128 {
		action := tui.ParseKeyAction(rune(r))
		actions[action] = true
	}

	// Verify all actions were found.
	for action, found := range actions {
		if !found {
			t.Errorf("KeyAction %d was never produced by ParseKeyAction", action)
		}
	}
}

func TestParseKeyActionCompleteness(t *testing.T) {
	// Test that key mapping is complete and correct.
	keyMappings := map[rune]tui.KeyAction{
		'q': tui.KeyActionQuit,
		'Q': tui.KeyActionQuit,
		'r': tui.KeyActionReset,
		'R': tui.KeyActionReset,
		'p': tui.KeyActionTogglePause,
		'P': tui.KeyActionTogglePause,
		'f': tui.KeyActionShowProfiles,
		'F': tui.KeyActionShowProfiles,
		'h': tui.KeyActionToggleHelp,
		'H': tui.KeyActionToggleHelp,
		'?': tui.KeyActionToggleHelp,
		'1': tui.KeyActionSetProfile1,
		'2': tui.KeyActionSetProfile2,
		'3': tui.KeyActionSetProfile3,
		'4': tui.KeyActionSetProfile4,
		'5': tui.KeyActionSetProfile5,
		'6': tui.KeyActionSetProfile6,
	}

	for r, expected := range keyMappings {
		result := tui.ParseKeyAction(r)
		if result != expected {
			t.Errorf("ParseKeyAction(%q) = %d, want %d", r, result, expected)
		}
	}
}

func TestKeyActionConstants(t *testing.T) {
	// Verify KeyAction constants are unique.
	actions := []tui.KeyAction{
		tui.KeyActionNone,
		tui.KeyActionQuit,
		tui.KeyActionReset,
		tui.KeyActionTogglePause,
		tui.KeyActionShowProfiles,
		tui.KeyActionToggleHelp,
		tui.KeyActionSetProfile1,
		tui.KeyActionSetProfile2,
		tui.KeyActionSetProfile3,
		tui.KeyActionSetProfile4,
		tui.KeyActionSetProfile5,
		tui.KeyActionSetProfile6,
	}

	seen := make(map[tui.KeyAction]bool)
	for _, a := range actions {
		if seen[a] {
			t.Errorf("duplicate KeyAction value: %d", a)
		}
		seen[a] = true
	}
}

func TestKeyActionNoneIsZero(t *testing.T) {
	// KeyActionNone should be zero value for easy checking.
	if tui.KeyActionNone != 0 {
		t.Errorf("KeyActionNone = %d, want 0", tui.KeyActionNone)
	}
}

func TestParseKeyActionProfileRange(t *testing.T) {
	// Verify profile keys 1-6 map to sequential action values.
	profileKeys := []rune{'1', '2', '3', '4', '5', '6'}
	expectedActions := []tui.KeyAction{
		tui.KeyActionSetProfile1,
		tui.KeyActionSetProfile2,
		tui.KeyActionSetProfile3,
		tui.KeyActionSetProfile4,
		tui.KeyActionSetProfile5,
		tui.KeyActionSetProfile6,
	}

	for i, key := range profileKeys {
		result := tui.ParseKeyAction(key)
		if result != expectedActions[i] {
			t.Errorf("ParseKeyAction('%c') = %d, want %d", key, result, expectedActions[i])
		}

		// Verify sequential ordering.
		if i > 0 {
			prevAction := tui.ParseKeyAction(profileKeys[i-1])
			if int(result) != int(prevAction)+1 {
				t.Errorf("profile actions should be sequential: %d should follow %d",
					result, prevAction)
			}
		}
	}
}

func BenchmarkParseKeyAction(b *testing.B) {
	keys := []rune{
		'q', 'Q', 'r', 'R', 'p', 'P', 'f', 'F', 'h', 'H', '?',
		'1', '2', '3', '4', '5', '6', 'a', 'x', '0',
	}
	for b.Loop() {
		for _, k := range keys {
			_ = tui.ParseKeyAction(k)
		}
	}
}

// Helper function to check if string contains substring.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// =============================================================================
// Consistency Tests
// =============================================================================

func TestGetPredefinedProfilesConsistency(t *testing.T) {
	// Multiple calls should return the same data.
	profiles1 := tui.GetPredefinedProfiles()
	profiles2 := tui.GetPredefinedProfiles()

	if len(profiles1) != len(profiles2) {
		t.Errorf("inconsistent profile counts: %d vs %d", len(profiles1), len(profiles2))
	}

	for i := range profiles1 {
		if profiles1[i].Name != profiles2[i].Name {
			t.Errorf("inconsistent profile[%d] name: %q vs %q",
				i, profiles1[i].Name, profiles2[i].Name)
		}
	}
}

func TestNewConsistency(t *testing.T) {
	// Multiple calls should return different instances.
	app1 := tui.New(nil)
	app2 := tui.New(nil)

	if app1 == app2 {
		t.Error("New() should return different instances")
	}
}

// =============================================================================
// Table-Driven Tests for Complete Coverage
// =============================================================================

func TestFormatNumberTableDriven(t *testing.T) {
	// Comprehensive table-driven tests.
	testCases := []struct {
		input    uint64
		contains string
	}{
		{0, "0"},
		{1, "1"},
		{999, "999"},
		{1000, "K"},
		{999999, "K"},
		{1000000, "M"},
		{999999999, "M"},
		{1000000000, "B"},
	}

	for _, tc := range testCases {
		result := tui.FormatNumber(tc.input)
		found := false
		for i := 0; i <= len(result)-len(tc.contains); i++ {
			if result[i:i+len(tc.contains)] == tc.contains {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("FormatNumber(%d) = %q, expected to contain %q",
				tc.input, result, tc.contains)
		}
	}
}

func TestFormatBytesTableDriven(t *testing.T) {
	testCases := []struct {
		input    uint64
		contains string
	}{
		{0, "B"},
		{1023, "B"},
		{1024, "KB"},
		{1048575, "KB"},
		{1048576, "MB"},
		{1073741823, "MB"},
		{1073741824, "GB"},
		{1099511627775, "GB"},
		{1099511627776, "TB"},
	}

	for _, tc := range testCases {
		result := tui.FormatBytes(tc.input)
		found := false
		for i := 0; i <= len(result)-len(tc.contains); i++ {
			if result[i:i+len(tc.contains)] == tc.contains {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("FormatBytes(%d) = %q, expected to contain %q",
				tc.input, result, tc.contains)
		}
	}
}

func TestFormatDurationTableDriven(t *testing.T) {
	testCases := []struct {
		input    time.Duration
		contains string
	}{
		{0, "s"},
		{time.Second, "s"},
		{time.Minute, "m"},
		{time.Hour, "h"},
		{2*time.Hour + 30*time.Minute + 45*time.Second, "h"},
	}

	for _, tc := range testCases {
		result := tui.FormatDuration(tc.input)
		found := false
		for _, c := range result {
			if string(c) == tc.contains {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("FormatDuration(%v) = %q, expected to contain %q",
				tc.input, result, tc.contains)
		}
	}
}
