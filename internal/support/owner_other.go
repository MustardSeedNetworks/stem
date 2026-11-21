// SPDX-License-Identifier: BUSL-1.1

//go:build !unix

package support

import (
	"io/fs"
	"os"
)

// ownerOf reports no owner: Windows files take their directory's ACL, which is
// the data directory's own.
func ownerOf(fs.FileInfo) (int, int, bool) { return 0, 0, false }

// giveOwner does nothing where there is no numeric owner.
func giveOwner(*os.Root, string, fs.FileInfo) error { return nil }
