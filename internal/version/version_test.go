// SPDX-License-Identifier: BUSL-1.1

// Black-box testing of version package.
//
// Tests verify public API behavior. SetForTesting is exported specifically
// to allow tests to modify version info for testing purposes.
//
// Tests CANNOT run in parallel (no t.Parallel()) because they:
//   - Modify shared package-level variables via SetForTesting
//   - Rely on defer to restore original values
//   - Would race if run concurrently (e.g., TestInfoReflectsModifiedVersion)
package version_test

import (
	"slices"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/version"
)

// ============================================================================
// Default Values Tests
// ============================================================================

func TestDefaultVersionValue(t *testing.T) {
	if version.GetVersion() == "" {
		t.Error("GetVersion is empty, expected a default value")
	}
}

func TestDefaultCommitValue(t *testing.T) {
	if version.GetCommit() == "" {
		t.Error("GetCommit is empty, expected a default value")
	}
}

func TestDefaultBuildTimeValue(t *testing.T) {
	if version.GetBuildTime() == "" {
		t.Error("GetBuildTime is empty, expected a default value")
	}
}

func TestDefaultUIBuildHashValue(t *testing.T) {
	// UIBuildHash defaults to "unknown" when not injected via ldflags.
	if version.GetUIBuildHash() == "" {
		t.Error("GetUIBuildHash is empty, expected a default value")
	}
}

// ============================================================================
// Info Function Tests
// ============================================================================

func TestInfoReturnsMap(t *testing.T) {
	info := version.Info()

	if info == nil {
		t.Fatal("Info() returned nil")
	}
}

func TestInfoContainsVersionKey(t *testing.T) {
	info := version.Info()

	v, ok := info["version"]
	if !ok {
		t.Error("Info() map missing 'version' key")
	}
	if v == "" {
		t.Error("Info() 'version' value is empty")
	}
}

func TestInfoContainsCommitKey(t *testing.T) {
	info := version.Info()

	commit, ok := info["commit"]
	if !ok {
		t.Error("Info() map missing 'commit' key")
	}
	if commit == "" {
		t.Error("Info() 'commit' value is empty")
	}
}

func TestInfoContainsBuildTimeKey(t *testing.T) {
	info := version.Info()

	buildTime, ok := info["buildTime"]
	if !ok {
		t.Error("Info() map missing 'buildTime' key")
	}
	if buildTime == "" {
		t.Error("Info() 'buildTime' value is empty")
	}
}

func TestInfoContainsUIBuildHashKey(t *testing.T) {
	info := version.Info()

	uih, ok := info["uiBuildHash"]
	if !ok {
		t.Error("Info() map missing 'uiBuildHash' key")
	}
	if uih == "" {
		t.Error("Info() 'uiBuildHash' value is empty")
	}
}

func TestInfoMapSize(t *testing.T) {
	info := version.Info()

	expectedSize := 4
	if len(info) != expectedSize {
		t.Errorf("Info() returned map with %d keys, want %d", len(info), expectedSize)
	}
}

func TestInfoMatchesGetters(t *testing.T) {
	info := version.Info()

	if info["version"] != version.GetVersion() {
		t.Errorf("Info()['version'] = %q, want %q (GetVersion())", info["version"], version.GetVersion())
	}
	if info["commit"] != version.GetCommit() {
		t.Errorf("Info()['commit'] = %q, want %q (GetCommit())", info["commit"], version.GetCommit())
	}
	if info["buildTime"] != version.GetBuildTime() {
		t.Errorf("Info()['buildTime'] = %q, want %q (GetBuildTime())", info["buildTime"], version.GetBuildTime())
	}
	if info["uiBuildHash"] != version.GetUIBuildHash() {
		t.Errorf(
			"Info()['uiBuildHash'] = %q, want %q (GetUIBuildHash())",
			info["uiBuildHash"],
			version.GetUIBuildHash(),
		)
	}
}

// ============================================================================
// Variable Modification Tests (simulate ldflags injection)
// ============================================================================

func TestInfoReflectsModifiedVersion(t *testing.T) {
	restore := version.SetForTesting("v1.2.3", "abc123def456", "2025-01-10T12:00:00Z")
	defer restore()

	info := version.Info()

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

func TestInfoReflectsModifiedUIBuildHash(t *testing.T) {
	restore := version.SetForTesting("v1.2.3", "abc123", "2025-01-10T12:00:00Z", "d41d8cd98f00b204e9800998ecf8427e")
	defer restore()

	info := version.Info()

	if info["uiBuildHash"] != "d41d8cd98f00b204e9800998ecf8427e" {
		t.Errorf("Info()['uiBuildHash'] = %q, want md5 hash", info["uiBuildHash"])
	}
}

func TestInfoWithSemanticVersion(t *testing.T) {
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
			restore := version.SetForTesting(tc.version, "unknown", "unknown")
			defer restore()

			info := version.Info()

			if info["version"] != tc.version {
				t.Errorf("Info()['version'] = %q, want %q", info["version"], tc.version)
			}
		})
	}
}

func TestInfoWithGitCommitFormats(t *testing.T) {
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
			restore := version.SetForTesting("dev", tc.commit, "unknown")
			defer restore()

			info := version.Info()

			if info["commit"] != tc.commit {
				t.Errorf("Info()['commit'] = %q, want %q", info["commit"], tc.commit)
			}
		})
	}
}

func TestInfoWithBuildTimeFormats(t *testing.T) {
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
			restore := version.SetForTesting("dev", "unknown", tc.buildTime)
			defer restore()

			info := version.Info()

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
	info1 := version.Info()
	info2 := version.Info()

	info1["version"] = "modified"

	if info2["version"] == "modified" {
		t.Error("Info() returns same map instance, expected new map each call")
	}
}

func TestInfoMapIsNotShared(t *testing.T) {
	info1 := version.Info()
	info2 := version.Info()

	info1["extra"] = "value"

	if _, ok := info2["extra"]; ok {
		t.Error("Info() maps share underlying storage")
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestInfoWithEmptyValues(t *testing.T) {
	restore := version.SetForTesting("", "", "")
	defer restore()

	info := version.Info()

	// Empty Version falls back to debug.ReadBuildInfo / "dev". Commit and
	// BuildTime get coerced to "unknown" when empty. UIBuildHash defaults
	// to "unknown" when empty.
	if info["version"] == "" {
		t.Error("Info()['version'] should never be empty (fallback path)")
	}
	if info["commit"] == "" {
		t.Error("Info()['commit'] should never be empty (coerces to 'unknown')")
	}
	if info["buildTime"] == "" {
		t.Error("Info()['buildTime'] should never be empty (coerces to 'unknown')")
	}
	if info["uiBuildHash"] == "" {
		t.Error("Info()['uiBuildHash'] should never be empty (coerces to 'unknown')")
	}
}

func TestInfoWithSpecialCharacters(t *testing.T) {
	testVersion := "v1.0.0-beta+build.123"
	testCommit := "abc123/feature-branch"
	testBuildTime := "2025-01-10T12:00:00+00:00"

	restore := version.SetForTesting(testVersion, testCommit, testBuildTime)
	defer restore()

	info := version.Info()

	if info["version"] != testVersion {
		t.Errorf("Info()['version'] = %q, want %q", info["version"], testVersion)
	}
	if info["commit"] != testCommit {
		t.Errorf("Info()['commit'] = %q, want %q", info["commit"], testCommit)
	}
	if info["buildTime"] != testBuildTime {
		t.Errorf("Info()['buildTime'] = %q, want %q", info["buildTime"], testBuildTime)
	}
}

func TestInfoWithUnicodeCharacters(t *testing.T) {
	testVersion := "v1.0.0-测试"
	restore := version.SetForTesting(testVersion, "unknown", "unknown")
	defer restore()

	info := version.Info()

	if info["version"] != testVersion {
		t.Errorf("Info()['version'] = %q, want %q", info["version"], testVersion)
	}
}

func TestInfoWithLongValues(t *testing.T) {
	longVersion := "v1.0.0-alpha.beta.gamma.delta.epsilon.zeta.eta.theta.iota.kappa"
	longCommit := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	restore := version.SetForTesting(longVersion, longCommit, "unknown")
	defer restore()

	info := version.Info()

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
	info := version.Info()

	expectedKeys := []string{"version", "commit", "buildTime", "uiBuildHash"}

	for _, key := range expectedKeys {
		if _, ok := info[key]; !ok {
			t.Errorf("Info() map missing expected key %q", key)
		}
	}

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
	done := make(chan bool)

	for range 100 {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Info() panicked during concurrent access: %v", r)
				}
				done <- true
			}()

			info := version.Info()
			_ = info["version"]
			_ = info["commit"]
			_ = info["buildTime"]
			_ = info["uiBuildHash"]
		}()
	}

	for range 100 {
		<-done
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkInfo(b *testing.B) {
	for b.Loop() {
		_ = version.Info()
	}
}

func BenchmarkInfoAccess(b *testing.B) {
	for b.Loop() {
		info := version.Info()
		_ = info["version"]
		_ = info["commit"]
		_ = info["buildTime"]
		_ = info["uiBuildHash"]
	}
}
