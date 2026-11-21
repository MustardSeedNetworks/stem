// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// archiveStamp is the timestamp in a default archive name.
	archiveStamp = "20060102T150405Z"

	// archiveMode is what an archive is created with: it holds credentials.
	archiveMode os.FileMode = 0o600
)

// defaultArchiveName names an archive for the current directory.
func defaultArchiveName(prefix string, now time.Time) string {
	return prefix + "-" + now.UTC().Format(archiveStamp) + ".tar.gz"
}

// createArchive creates the output file for a backup or bundle. Both carry
// credentials, so it is owner-only, and it must not exist: overwriting an
// earlier archive on a typo would destroy the only copy. The create goes
// through an [os.Root] on the file's directory, so the operator's path cannot
// name its way anywhere but that directory.
func createArchive(path string) (*os.File, error) {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return nil, fmt.Errorf("open the directory for %s: %w", path, err)
	}
	defer func() { _ = root.Close() }()

	file, err := root.OpenFile(filepath.Base(path), os.O_WRONLY|os.O_CREATE|os.O_EXCL, archiveMode)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", path, err)
	}
	return file, nil
}

// writeArchive creates path and fills it with write. A half-written archive is
// removed on failure: an operator must never keep a truncated backup believing
// it complete.
func writeArchive(path string, write func(*os.File) error) error {
	file, err := createArchive(path)
	if err != nil {
		return err
	}
	if writeErr := write(file); writeErr != nil {
		_ = file.Close()
		_ = removeArchive(path)
		return writeErr
	}
	if closeErr := file.Close(); closeErr != nil {
		_ = removeArchive(path)
		return fmt.Errorf("close %s: %w", path, closeErr)
	}
	return nil
}

func removeArchive(path string) error {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	return root.Remove(filepath.Base(path))
}

// openArchive opens an archive to read, through an [os.Root] on its directory.
func openArchive(path string) (*os.File, error) {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return nil, fmt.Errorf("open the directory for %s: %w", path, err)
	}
	defer func() { _ = root.Close() }()

	file, err := root.Open(filepath.Base(path))
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	return file, nil
}
