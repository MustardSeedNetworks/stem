// SPDX-License-Identifier: BUSL-1.1

package update

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// The state copy makes a database migration by the new version undoable
// along with the binary (niac-go#2092). It is taken with the service stopped
// and lives beside the data directory, so restoring it is a rename on the
// same filesystem. The run lock is left out both ways: the update holds it
// on the open file, and swapping the file underneath would let the restarted
// daemon lock a different inode. Every path is resolved through an os.Root,
// so a symlink in the data directory is copied as a link and never followed
// out of it.

// copyState copies the data directory src to dst, which must not exist,
// keeping modes, ownership and symlinks.
func copyState(src, dst string) error {
	srcRoot, err := os.OpenRoot(src)
	if err != nil {
		return err
	}
	defer srcRoot.Close()
	info, err := srcRoot.Lstat(".")
	if err != nil {
		return err
	}
	if err = os.Mkdir(dst, info.Mode().Perm()); err != nil {
		return err
	}
	dstRoot, err := os.OpenRoot(dst)
	if err != nil {
		return err
	}
	defer dstRoot.Close()
	return fs.WalkDir(srcRoot.FS(), ".", func(name string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if name == RunLockName {
			return nil
		}
		entry, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		var copyErr error
		switch mode := entry.Mode(); {
		case name == ".":
		case mode.IsDir():
			copyErr = dstRoot.Mkdir(name, mode.Perm())
		case mode&fs.ModeSymlink != 0:
			var link string
			if link, copyErr = srcRoot.Readlink(name); copyErr == nil {
				copyErr = dstRoot.Symlink(link, name)
			}
		case mode.IsRegular():
			copyErr = copyFile(srcRoot, dstRoot, name, mode.Perm())
		default:
			return nil // sockets and FIFOs belong to a running daemon, which is stopped
		}
		if copyErr != nil {
			return copyErr
		}
		return copyOwner(dstRoot, name, entry)
	})
}

func copyFile(src, dst *os.Root, name string, perm fs.FileMode) error {
	in, err := src.Open(name)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := dst.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		return errors.Join(err, out.Close())
	}
	return out.Close()
}

// restoreState puts the copy taken by copyState back in dataDir and removes
// it. The copy must sit under dataDir's parent.
func restoreState(dataDir, copied string) error {
	parent, err := os.OpenRoot(filepath.Dir(dataDir))
	if err != nil {
		return err
	}
	defer parent.Close()
	data := filepath.Base(dataDir)
	saved, err := filepath.Rel(parent.Name(), copied)
	if err != nil {
		return err
	}
	current, err := fs.ReadDir(parent.FS(), data)
	if err != nil {
		return err
	}
	for _, entry := range current {
		if entry.Name() == RunLockName {
			continue
		}
		if err = parent.RemoveAll(filepath.Join(data, entry.Name())); err != nil {
			return err
		}
	}
	entries, err := fs.ReadDir(parent.FS(), filepath.ToSlash(saved))
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err = parent.Rename(filepath.Join(saved, entry.Name()), filepath.Join(data, entry.Name())); err != nil {
			return err
		}
	}
	return parent.Remove(saved)
}
