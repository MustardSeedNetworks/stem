// SPDX-License-Identifier: BUSL-1.1

package license

import (
	"fmt"
	"os"
	"path/filepath"

	fnd "github.com/MustardSeedNetworks/foundation/pkg/license"
)

// Load builds the license manager for Stem's default config directory. How the
// state on disk loaded is the manager's own answer: ask it for LoadStatus and,
// when that is not Usable, LoadError for the one reason to log.
func Load() (*Manager, error) {
	dir, dirErr := defaultConfigDir()
	if dirErr != nil {
		return nil, dirErr
	}
	return LoadFromDir(dir)
}

// LoadFromDir is Load rooted at configDir. Tests use it to keep activation
// state in a temp directory instead of the developer's ~/.config/stem.
func LoadFromDir(configDir string) (*Manager, error) {
	return fnd.NewManagerWithDir(fnd.NewProductionVerifier(Policy()), Policy(), configDir)
}

// DefaultLicensePath returns the file Stem reads activation state from, so a
// message about an unusable license can name the file to replace.
func DefaultLicensePath() string {
	dir, dirErr := defaultConfigDir()
	if dirErr != nil {
		return licenseFileName
	}
	return filepath.Join(dir, licenseFileName)
}

func defaultConfigDir() (string, error) {
	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return "", fmt.Errorf("locate home directory for license state: %w", homeErr)
	}
	return filepath.Join(home, ".config", configSubdir), nil
}
