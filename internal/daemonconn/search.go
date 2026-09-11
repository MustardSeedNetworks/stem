// SPDX-License-Identifier: BUSL-1.1

package daemonconn

// Finding the running daemon. The CLI is normally invoked from an
// operator's shell while the daemon keeps its state under the packaged data
// directory, so "look in the working directory" — which is all the daemon
// itself defaults to — finds nothing. Discovery walks the same places the
// daemon may have been started from, override first.

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// dataDirEnv overrides where both the daemon writes and the CLI looks.
const dataDirEnv = "STEM_DATA_DIR"

// SearchDirs lists the data directories a client checks, in order: the
// explicit override, then the platform's packaged location, then the
// working directory the daemon falls back to when nothing is configured.
func SearchDirs() []string {
	var dirs []string
	if override := os.Getenv(dataDirEnv); override != "" {
		dirs = append(dirs, override)
	}
	dirs = append(dirs, packagedDataDir(), ".")
	return dirs
}

// packagedDataDir is where the shipped service unit keeps daemon state.
func packagedDataDir() string {
	if runtime.GOOS == "windows" {
		if programData := os.Getenv("ProgramData"); programData != "" {
			return filepath.Join(programData, "stem")
		}
		return `C:\ProgramData\stem`
	}
	return "/var/lib/stem"
}

// Discover returns the descriptor of the first daemon found on the search
// path. A descriptor that exists but cannot be used — wrong permissions,
// corrupt, half-written — is reported rather than skipped: silently moving
// on would hide the very thing the operator needs to fix.
func Discover() (Descriptor, error) {
	for _, dir := range SearchDirs() {
		descriptor, err := Read(dir)
		switch {
		case err == nil:
			return descriptor, nil
		case IsNotFound(err):
			continue
		default:
			return Descriptor{}, err
		}
	}
	return Descriptor{}, ErrNotFound
}

// IsNotFound reports whether err means "no descriptor here", as opposed to
// one that is present but unusable.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
