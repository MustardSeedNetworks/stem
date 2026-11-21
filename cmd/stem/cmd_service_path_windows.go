// SPDX-License-Identifier: BUSL-1.1

//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// underDir reports whether file is dir or inside it, ignoring case as the
// Windows file system does.
func underDir(dir, file string) bool {
	dir = filepath.Clean(dir)
	file = filepath.Clean(file)
	return strings.EqualFold(file, dir) ||
		strings.HasPrefix(strings.ToLower(file), strings.ToLower(dir)+string(filepath.Separator))
}

// serviceExecutable resolves the binary the service would run and refuses one
// outside %ProgramFiles%\Stem. A LocalSystem service runs whatever the path
// names, so a binary a normal user can replace would be a path to SYSTEM.
func serviceExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate this binary: %w", err)
	}
	if exe, err = filepath.EvalSymlinks(exe); err != nil {
		return "", fmt.Errorf("resolve this binary: %w", err)
	}
	programFiles, err := windows.KnownFolderPath(windows.FOLDERID_ProgramFiles, 0)
	if err != nil {
		return "", fmt.Errorf("locate Program Files: %w", err)
	}
	installDir := filepath.Join(programFiles, "Stem")
	if !underDir(installDir, exe) {
		return "", fmt.Errorf(
			"refusing to register %s as a service: it runs as SYSTEM, so the binary must be in %s; run install.ps1 from the release zip",
			exe,
			installDir,
		)
	}
	return exe, nil
}
