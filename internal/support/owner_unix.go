// SPDX-License-Identifier: BUSL-1.1

//go:build unix

package support

import (
	"io/fs"
	"os"
	"syscall"
)

// ownerOf returns the numeric user and group that own a file, and whether the
// platform reports them.
func ownerOf(info fs.FileInfo) (int, int, bool) {
	st, isStat := info.Sys().(*syscall.Stat_t)
	if !isStat {
		return 0, 0, false
	}
	return int(st.Uid), int(st.Gid), true
}

// giveOwner gives name the owner of ref, through root so a link is never
// followed. The restore runs as root and the daemon as its own account, and a
// state file the daemon cannot read stops it starting.
func giveOwner(root *os.Root, name string, ref fs.FileInfo) error {
	uid, gid, ok := ownerOf(ref)
	if !ok {
		return nil
	}
	return root.Lchown(name, uid, gid)
}
