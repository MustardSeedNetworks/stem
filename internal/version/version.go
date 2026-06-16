// Package version provides build-time version information for The Stem.
//
// Values are populated in this order of precedence:
//  1. -ldflags injected at build time (canonical for releases)
//  2. [runtime/debug.ReadBuildInfo] VCS settings (modules built locally)
//  3. Compile-time defaults ("dev" / "unknown")
//
// ldflags example:
//
//	go build -ldflags "-X github.com/MustardSeedNetworks/stem/internal/version.Version=v1.2.3"
package version

import (
	"runtime/debug"
	"strings"
)

// These variables are set via ldflags at build time. Names match the
// canonical contract used by sibling projects (seed, niac).
//
//nolint:gochecknoglobals // Build metadata injected via ldflags.
var (
	Version     string // -X .../internal/version.Version=v1.2.3
	Commit      string // -X .../internal/version.Commit=abc1234
	BuildTime   string // -X .../internal/version.BuildTime=2025-01-10T12:00:00Z
	UIBuildHash string // -X .../internal/version.UIBuildHash=<md5 of internal/api/ui>
)

const (
	shortCommitLen = 7
	defaultVersion = "dev"
	unknownValue   = "unknown"
)

// extractVersionFromBuildInfo pulls version/commit/buildTime out of a
// [debug.BuildInfo]. Separated for testability with mock build info.
func extractVersionFromBuildInfo(info *debug.BuildInfo) (string, string, string) {
	ver := defaultVersion
	commit := unknownValue
	buildTime := unknownValue

	if info == nil {
		return ver, commit, buildTime
	}

	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		ver = info.Main.Version
	}

	var modified bool
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			commit = setting.Value
			if len(commit) > shortCommitLen {
				commit = commit[:shortCommitLen]
			}
		case "vcs.time":
			buildTime = setting.Value
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}

	if modified && ver != defaultVersion {
		ver += "-dirty"
	}

	return ver, commit, buildTime
}

// getVersionInfo returns version/commit/buildTime, preferring ldflags over
// [debug.ReadBuildInfo] over defaults.
func getVersionInfo() (string, string, string) {
	if Version != "" {
		ver := Version
		commit := Commit
		buildTime := BuildTime
		if commit == "" {
			commit = unknownValue
		}
		if buildTime == "" {
			buildTime = unknownValue
		}
		return ver, commit, buildTime
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return defaultVersion, unknownValue, unknownValue
	}
	return extractVersionFromBuildInfo(info)
}

// GetVersion returns the semantic version.
func GetVersion() string {
	ver, _, _ := getVersionInfo()
	return ver
}

// GetCommit returns the git commit hash (short form).
func GetCommit() string {
	_, commit, _ := getVersionInfo()
	return commit
}

// GetBuildTime returns the build timestamp.
func GetBuildTime() string {
	_, _, buildTime := getVersionInfo()
	return buildTime
}

// GetReleaseTrain returns the calendar release train in YYYY.MM format.
func GetReleaseTrain() string {
	return releaseTrainFromBuildTime(GetBuildTime())
}

// GetUIBuildHash returns the md5 hash of the embedded UI assets.
// Returns unknownValue when not injected via -ldflags at build time.
func GetUIBuildHash() string {
	if UIBuildHash == "" {
		return unknownValue
	}
	return UIBuildHash
}

func releaseTrainFromBuildTime(buildTime string) string {
	if len(buildTime) < len("2006-01") || buildTime == unknownValue {
		return unknownValue
	}
	return strings.ReplaceAll(buildTime[:len("2006-01")], "-", ".")
}

// Info returns all version information as a JSON-ready map. Keys are
// lowercase camelCase to match the /__version endpoint contract shared
// across seed, stem, and niac.
func Info() map[string]string {
	ver, commit, buildTime := getVersionInfo()
	return map[string]string{
		"version":      ver,
		"commit":       commit,
		"buildTime":    buildTime,
		"releaseTrain": releaseTrainFromBuildTime(buildTime),
		"uiBuildHash":  GetUIBuildHash(),
	}
}

// SetForTesting overrides version info for the duration of a test. The
// returned function restores the previous values; defer it immediately.
// uiBuildHash is variadic for backwards compatibility with three-arg
// callers; pass at most one value.
func SetForTesting(ver, cmt, bt string, uih ...string) func() {
	origV, origC, origB, origU := Version, Commit, BuildTime, UIBuildHash
	Version = ver
	Commit = cmt
	BuildTime = bt
	if len(uih) > 0 {
		UIBuildHash = uih[0]
	}
	return func() {
		Version, Commit, BuildTime, UIBuildHash = origV, origC, origB, origU
	}
}
