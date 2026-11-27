// SPDX-License-Identifier: BUSL-1.1

package license

import (
	"path/filepath"

	fnd "github.com/MustardSeedNetworks/foundation/pkg/license"

	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
)

// Load builds the license manager rooted in Stem's data directory, the same
// directory the daemon and the CLI both resolve with daemonconn.ServiceDataDir.
// The installed daemon runs as a service account and the operator's console as
// someone else, so a per-user location (the old ~/.config/stem) made
// `stem license` with no daemon running write a file the service never read.
// How the state on disk loaded is the manager's own answer: ask it for
// LoadStatus and, when that is not Usable, LoadError for the one reason to log.
func Load() (*Manager, error) {
	return LoadFromDir(daemonconn.ServiceDataDir())
}

// LoadFromDir is Load rooted at configDir. Tests use it to keep activation
// state in a temp directory instead of the developer's real data directory.
func LoadFromDir(configDir string) (*Manager, error) {
	return fnd.NewManagerWithDir(fnd.NewProductionVerifier(Policy()), Policy(), configDir)
}

// DefaultLicensePath returns the file Stem reads activation state from, so a
// message about an unusable license can name the file to replace.
func DefaultLicensePath() string {
	return filepath.Join(daemonconn.ServiceDataDir(), licenseFileName)
}
