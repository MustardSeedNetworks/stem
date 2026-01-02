// Package version provides build-time version information for The Stem.
// Version variables are populated via ldflags during build.
package version

//nolint:gochecknoglobals // Version vars must be global for ldflags injection at build time.
var (
	// Version is the semantic version (set via ldflags).
	Version = "dev"

	// Commit is the git commit SHA (set via ldflags).
	Commit = "unknown"

	// BuildTime is the UTC build timestamp (set via ldflags).
	BuildTime = "unknown"
)

// Info returns version information as a map for JSON serialization.
func Info() map[string]string {
	return map[string]string{
		"version":   Version,
		"commit":    Commit,
		"buildTime": BuildTime,
	}
}
