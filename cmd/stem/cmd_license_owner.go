// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"errors"
	"fmt"
	"os"
)

// licenseWriteAction is what the CLI must do before a license write it makes
// itself, with no daemon running, into the packaged data directory.
type licenseWriteAction int

const (
	// writeAsIs: the file will be owned by whoever the data directory is owned
	// by, or there is no packaged directory to match.
	writeAsIs licenseWriteAction = iota
	// writeAsOwner: root takes on the directory owner's identity first, so the
	// file is the service account's own, written with that account's authority
	// and no more.
	writeAsOwner
	// writeRefused: neither root nor the owner; a write would leave a file the
	// service cannot read, or land somewhere else, so nothing is written.
	writeRefused
)

var (
	errLicenseWriteOwner = errors.New(
		"run as root or as the stem user: the license belongs to the account that runs the service",
	)
	errLicenseLinked = errors.New("is a symbolic link or not a regular file, refusing to write through it")
)

// decideLicenseWrite is the whole ownership rule. euid is the caller's user id
// and dirOwner the data directory's. It is only consulted where file owners are
// user ids: macOS (daemon and data directory both root's) and Windows
// (SYSTEM and Administrators, which the DACL already settles) reach here with
// matching or no owner, so they never switch identity.
func decideLicenseWrite(euid, dirOwner int) licenseWriteAction {
	if euid == dirOwner {
		return writeAsIs
	}
	if euid == 0 {
		return writeAsOwner
	}
	return writeRefused
}

// prepareLicenseWrite is called before the CLI writes the license itself. It
// refuses a caller who could not leave a readable file, becomes the directory's
// owner when that caller is root, and refuses a link or irregular file at the
// license path.
func prepareLicenseWrite(path, dataDir string) error {
	ids, known := licenseOwnership(dataDir)
	if known {
		switch decideLicenseWrite(ids.euid, ids.uid) {
		case writeRefused:
			return fmt.Errorf("%w (%s is owned by user id %d)", errLicenseWriteOwner, dataDir, ids.uid)
		case writeAsOwner:
			if err := becomeUser(ids.uid, ids.gid); err != nil {
				return err
			}
		case writeAsIs:
		}
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s %w", path, errLicenseLinked)
	}
	return nil
}
