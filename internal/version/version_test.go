//nolint:testpackage,paralleltest // Tests internal state; tests modify globals so cannot run in parallel.
package version

import (
	"slices"
	"testing"
)

// ============================================================================
// Default Values Tests
// ============================================================================

func TestDefaultVersionValue(t *testing.T) {
	// Version should have a default value (either "dev" or set via ldflags)
	if Version == "" {
		t.Error("Version is empty, expected a default value")
	}
}

func TestDefaultCommitValue(t *testing.T) {
	// Commit should have a default value (either "unknown" or set via ldflags)
	if Commit == "" {
		t.Error("Commit is empty, expected a default value")
	}
}

func TestDefaultBuildTimeValue(t *testing.T) {
	// BuildTime should have a default value (either "unknown" or set via ldflags)
	if BuildTime == "" {
		t.Error("BuildTime is empty, expected a default value")
	}
}

// ============================================================================
// Info Function Tests
// ============================================================================

func TestInfoReturnsMap(t *testing.T) {
	info := Info()

	if info == nil {
		t.Fatal("Info() returned nil")
	}
}

func TestInfoContainsVersionKey(t *testing.T) {
	info := Info()

	version, ok := info["version"]
	if !ok {
		t.Error("Info() map missing 'version' key")
	}
	if version == "" {
		t.Error("Info() 'version' value is empty")
	}
}

func TestInfoContainsCommitKey(t *testing.T) {
	info := Info()

	commit, ok := info["commit"]
	if !ok {
		t.Error("Info() map missing 'commit' key")
	}
	if commit == "" {
		t.Error("Info() 'commit' value is empty")
	}
}

func TestInfoContainsBuildTimeKey(t *testing.T) {
	info := Info()

	buildTime, ok := info["buildTime"]
	if !ok {
		t.Error("Info() map missing 'buildTime' key")
	}
	if buildTime == "" {
		t.Error("Info() 'buildTime' value is empty")
	}
}

func TestInfoMapSize(t *testing.T) {
	info := Info()

	expectedSize := 3
	if len(info) != expectedSize {
		t.Errorf("Info() returned map with %d keys, want %d", len(info), expectedSize)
	}
}

func TestInfoMatchesGlobalVariables(t *testing.T) {
	info := Info()

	if info["version"] != Version {
		t.Errorf("Info()['version'] = %q, want %q (Version variable)", info["version"], Version)
	}
	if info["commit"] != Commit {
		t.Errorf("Info()['commit'] = %q, want %q (Commit variable)", info["commit"], Commit)
	}
	if info["buildTime"] != BuildTime {
		t.Errorf("Info()['buildTime'] = %q, want %q (BuildTime variable)", info["buildTime"], BuildTime)
	}
}

// ============================================================================
// Variable Modification Tests (simulate ldflags injection)
// ============================================================================

func TestInfoReflectsModifiedVersion(t *testing.T) {
	// Save original values
	origVersion := Version
	origCommit := Commit
	origBuildTime := BuildTime

	// Restore after test
	defer func() {
		Version = origVersion
		Commit = origCommit
		BuildTime = origBuildTime
	}()

	// Simulate ldflags injection
	Version = "v1.2.3"
	Commit = "abc123def456"
	BuildTime = "2025-01-10T12:00:00Z"

	info := Info()

	if info["version"] != "v1.2.3" {
		t.Errorf("Info()['version'] = %q, want %q", info["version"], "v1.2.3")
	}
	if info["commit"] != "abc123def456" {
		t.Errorf("Info()['commit'] = %q, want %q", info["commit"], "abc123def456")
	}
	if info["buildTime"] != "2025-01-10T12:00:00Z" {
		t.Errorf("Info()['buildTime'] = %q, want %q", info["buildTime"], "2025-01-10T12:00:00Z")
	}
}

func TestInfoWithSemanticVersion(t *testing.T) {
	origVersion := Version
	defer func() { Version = origVersion }()

	testCases := []struct {
		name    string
		version string
	}{
		{"major only", "v1"},
		{"major minor", "v1.2"},
		{"full semver", "v1.2.3"},
		{"prerelease", "v1.2.3-alpha"},
		{"prerelease with number", "v1.2.3-beta.1"},
		{"build metadata", "v1.2.3+build123"},
		{"prerelease and build", "v1.2.3-rc.1+build456"},
		{"without v prefix", "1.2.3"},
		{"dev version", "dev"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Version = tc.version
			info := Info()

			if info["version"] != tc.version {
				t.Errorf("Info()['version'] = %q, want %q", info["version"], tc.version)
			}
		})
	}
}

func TestInfoWithGitCommitFormats(t *testing.T) {
	origCommit := Commit
	defer func() { Commit = origCommit }()

	testCases := []struct {
		name   string
		commit string
	}{
		{"short hash", "abc123d"},
		{"full hash", "abc123def456789012345678901234567890abcd"},
		{"unknown", "unknown"},
		{"dirty commit", "abc123d-dirty"},
		{"with branch", "main-abc123d"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Commit = tc.commit
			info := Info()

			if info["commit"] != tc.commit {
				t.Errorf("Info()['commit'] = %q, want %q", info["commit"], tc.commit)
			}
		})
	}
}

func TestInfoWithBuildTimeFormats(t *testing.T) {
	origBuildTime := BuildTime
	defer func() { BuildTime = origBuildTime }()

	testCases := []struct {
		name      string
		buildTime string
	}{
		{"ISO8601 UTC", "2025-01-10T12:00:00Z"},
		{"ISO8601 with offset", "2025-01-10T12:00:00+05:00"},
		{"RFC3339", "2025-01-10T12:00:00.000Z"},
		{"Unix timestamp", "1736510400"},
		{"unknown", "unknown"},
		{"date only", "2025-01-10"},
		{"human readable", "Jan 10 2025 12:00:00"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			BuildTime = tc.buildTime
			info := Info()

			if info["buildTime"] != tc.buildTime {
				t.Errorf("Info()['buildTime'] = %q, want %q", info["buildTime"], tc.buildTime)
			}
		})
	}
}

// ============================================================================
// Map Independence Tests
// ============================================================================

func TestInfoReturnsNewMapEachCall(t *testing.T) {
	info1 := Info()
	info2 := Info()

	// Modify info1
	info1["version"] = "modified"

	// info2 should not be affected
	if info2["version"] == "modified" {
		t.Error("Info() returns same map instance, expected new map each call")
	}
}

func TestInfoMapIsNotShared(t *testing.T) {
	info1 := Info()
	info2 := Info()

	// Add extra key to info1
	info1["extra"] = "value"

	// info2 should not have the extra key
	if _, ok := info2["extra"]; ok {
		t.Error("Info() maps share underlying storage")
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestInfoWithEmptyValues(t *testing.T) {
	origVersion := Version
	origCommit := Commit
	origBuildTime := BuildTime

	defer func() {
		Version = origVersion
		Commit = origCommit
		BuildTime = origBuildTime
	}()

	// Set empty values
	Version = ""
	Commit = ""
	BuildTime = ""

	info := Info()

	// Should still return a map with empty strings
	if info["version"] != "" {
		t.Errorf("Info()['version'] = %q, want empty string", info["version"])
	}
	if info["commit"] != "" {
		t.Errorf("Info()['commit'] = %q, want empty string", info["commit"])
	}
	if info["buildTime"] != "" {
		t.Errorf("Info()['buildTime'] = %q, want empty string", info["buildTime"])
	}
}

func TestInfoWithSpecialCharacters(t *testing.T) {
	origVersion := Version
	origCommit := Commit
	origBuildTime := BuildTime

	defer func() {
		Version = origVersion
		Commit = origCommit
		BuildTime = origBuildTime
	}()

	// Test with special characters that might appear in build info
	Version = "v1.0.0-beta+build.123"
	Commit = "abc123/feature-branch"
	BuildTime = "2025-01-10T12:00:00+00:00"

	info := Info()

	if info["version"] != Version {
		t.Errorf("Info()['version'] = %q, want %q", info["version"], Version)
	}
	if info["commit"] != Commit {
		t.Errorf("Info()['commit'] = %q, want %q", info["commit"], Commit)
	}
	if info["buildTime"] != BuildTime {
		t.Errorf("Info()['buildTime'] = %q, want %q", info["buildTime"], BuildTime)
	}
}

func TestInfoWithUnicodeCharacters(t *testing.T) {
	origVersion := Version
	defer func() { Version = origVersion }()

	// Edge case: unicode in version (unlikely but should not panic)
	Version = "v1.0.0-测试"
	info := Info()

	if info["version"] != "v1.0.0-测试" {
		t.Errorf("Info()['version'] = %q, want %q", info["version"], "v1.0.0-测试")
	}
}

func TestInfoWithLongValues(t *testing.T) {
	origVersion := Version
	origCommit := Commit

	defer func() {
		Version = origVersion
		Commit = origCommit
	}()

	// Test with unusually long values
	longVersion := "v1.0.0-alpha.beta.gamma.delta.epsilon.zeta.eta.theta.iota.kappa"
	longCommit := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	Version = longVersion
	Commit = longCommit

	info := Info()

	if info["version"] != longVersion {
		t.Errorf("Info()['version'] = %q, want %q", info["version"], longVersion)
	}
	if info["commit"] != longCommit {
		t.Errorf("Info()['commit'] = %q, want %q", info["commit"], longCommit)
	}
}

// ============================================================================
// JSON Serialization Compatibility Tests
// ============================================================================

func TestInfoKeysAreJSONSafe(t *testing.T) {
	info := Info()

	// Verify all keys are valid JSON keys (no special characters)
	expectedKeys := []string{"version", "commit", "buildTime"}

	for _, key := range expectedKeys {
		if _, ok := info[key]; !ok {
			t.Errorf("Info() map missing expected key %q", key)
		}
	}

	// Verify no unexpected keys
	for key := range info {
		if !slices.Contains(expectedKeys, key) {
			t.Errorf("Info() map contains unexpected key %q", key)
		}
	}
}

// ============================================================================
// Concurrency Safety Tests
// ============================================================================

func TestInfoConcurrentCalls(t *testing.T) {
	// Info() should be safe to call concurrently since it creates a new map each time
	done := make(chan bool)

	for range 100 {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Info() panicked during concurrent access: %v", r)
				}
				done <- true
			}()

			info := Info()
			_ = info["version"]
			_ = info["commit"]
			_ = info["buildTime"]
		}()
	}

	// Wait for all goroutines
	for range 100 {
		<-done
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkInfo(b *testing.B) {
	for b.Loop() {
		_ = Info()
	}
}

func BenchmarkInfoAccess(b *testing.B) {
	for b.Loop() {
		info := Info()
		_ = info["version"]
		_ = info["commit"]
		_ = info["buildTime"]
	}
}
