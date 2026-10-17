// SPDX-License-Identifier: BUSL-1.1

//go:build unix

package update

import (
	"io/fs"
	"os"
	"syscall"
)

// shareWithDataDir gives path the owner of the data directory.
func shareWithDataDir(path, dataDir string) error {
	info, err := os.Lstat(dataDir)
	if err != nil {
		return err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil
	}
	return os.Lchown(path, int(st.Uid), int(st.Gid))
}

// copyOwner gives the copy the original's owner: the update runs as root,
// and a restored state directory the daemon's account cannot write would
// stop it starting.
func copyOwner(root *os.Root, name string, info fs.FileInfo) error {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil
	}
	return root.Lchown(name, int(st.Uid), int(st.Gid))
}
