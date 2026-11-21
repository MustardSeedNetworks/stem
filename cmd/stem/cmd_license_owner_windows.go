// SPDX-License-Identifier: BUSL-1.1

//go:build windows

package main

// Windows file ownership is a security descriptor, not a user id: the data
// directory's protected DACL (SYSTEM and Administrators) already decides who
// can read what the CLI writes there, and the elevated CLI and the service are
// both in that set. There is nothing to match and no identity to take on.
type dataDirIDs struct {
	euid, uid, gid int
}

func licenseOwnership(string) (dataDirIDs, bool) { return dataDirIDs{}, false }

func becomeUser(int, int) error { return nil }
