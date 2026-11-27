// SPDX-License-Identifier: BUSL-1.1

//go:build !windows

package main

import (
	"fmt"
	"os"
	"syscall"
)

// dataDirIDs are the caller's effective user id and the data directory's owner
// and group.
type dataDirIDs struct {
	euid, uid, gid int
}

// licenseOwnership reports the caller's user id and the data directory's owner.
// known is false when the directory does not exist (no packaged install),
// where the manager creates it as the caller and there is nothing to match.
func licenseOwnership(dataDir string) (dataDirIDs, bool) {
	info, err := os.Stat(dataDir)
	if err != nil {
		return dataDirIDs{}, false
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return dataDirIDs{}, false
	}
	return dataDirIDs{euid: os.Geteuid(), uid: int(stat.Uid), gid: int(stat.Gid)}, true
}

// becomeUser drops root to uid and gid for the rest of the process: groups
// first, then the group, then the user, since the user change takes the right
// to make the others away. It is irreversible, which is the point: the license
// write that follows runs with the service account's authority only, so
// foundation's path-based write cannot be turned against a file the service
// account does not own. The CLI exits after the write.
//
// foundation's license manager reads, writes and removes its file by path and
// takes no caller-supplied store, so it cannot write through securedir; this is
// stem's side of that gap.
func becomeUser(uid, gid int) error {
	if err := syscall.Setgroups([]int{gid}); err != nil {
		return fmt.Errorf("drop to user id %d: %w", uid, err)
	}
	if err := syscall.Setgid(gid); err != nil {
		return fmt.Errorf("drop to group id %d: %w", gid, err)
	}
	if err := syscall.Setuid(uid); err != nil {
		return fmt.Errorf("drop to user id %d: %w", uid, err)
	}
	if os.Geteuid() != uid || os.Getuid() != uid {
		return fmt.Errorf("could not become user id %d", uid)
	}
	return nil
}
