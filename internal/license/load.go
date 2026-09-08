// SPDX-License-Identifier: BUSL-1.1

package license

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	fnd "github.com/MustardSeedNetworks/foundation/pkg/license"
)

// Status reports what the activation state on disk was when a manager was
// built. foundation loads that state best-effort and discards the difference
// between "no file" and "a file we could not use"; Stem has to keep the
// difference, because a missing file may start a trial and a damaged one must
// not. Handing out a Professional trial for corrupting a file is a paid tier
// given away (#1068).
type Status int

const (
	// StatusLoaded means a license file was read and parsed. Whether it
	// still entitles anything is then the manager's answer, not this one's.
	StatusLoaded Status = iota
	// StatusMissing means no license file exists: a fresh install, which is
	// the only state a trial may start from.
	StatusMissing
	// StatusUnreadable means a license file exists but could not be opened.
	StatusUnreadable
	// StatusMalformed means the file was read but did not decrypt or parse.
	StatusMalformed
)

// String names the status for an operator-facing message.
func (s Status) String() string {
	switch s {
	case StatusLoaded:
		return "loaded"
	case StatusMissing:
		return "missing"
	case StatusUnreadable:
		return "unreadable"
	case StatusMalformed:
		return "malformed"
	}
	return "unknown"
}

// Usable reports whether the state on disk can be acted on. An unusable state
// entitles nothing beyond the Free grant and must never be overwritten by an
// automatic trial: the operator has a license file whose contents this build
// cannot read, and destroying it is not this program's call.
func (s Status) Usable() bool {
	return s == StatusLoaded || s == StatusMissing
}

// Load builds the license manager for Stem's default config directory and
// reports how the state on disk loaded.
func Load() (*Manager, Status, error) {
	dir, dirErr := defaultConfigDir()
	if dirErr != nil {
		return nil, StatusUnreadable, dirErr
	}
	return LoadFromDir(dir)
}

// LoadFromDir is Load rooted at configDir. Tests use it to keep activation
// state in a temp directory instead of the developer's ~/.config/stem.
func LoadFromDir(configDir string) (*Manager, Status, error) {
	mgr, mgrErr := fnd.NewManagerWithDir(fnd.NewProductionVerifier(Policy()), Policy(), configDir)
	if mgrErr != nil {
		return nil, StatusUnreadable, mgrErr
	}
	if mgr.GetState() != nil {
		return mgr, StatusLoaded, nil
	}
	return mgr, classifyAbsentState(filepath.Join(configDir, licenseFileName)), nil
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

// classifyAbsentState says why a manager came up with no state. foundation has
// already tried to read the file, so this only separates "not there" from
// "there and unusable"; it deliberately does not re-parse.
func classifyAbsentState(path string) Status {
	f, openErr := os.Open(filepath.Clean(path))
	if openErr != nil {
		if errors.Is(openErr, fs.ErrNotExist) {
			return StatusMissing
		}
		return StatusUnreadable
	}
	_ = f.Close()
	return StatusMalformed
}
