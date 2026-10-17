// SPDX-License-Identifier: BUSL-1.1

package update

import (
	"io/fs"
	"os"
)

// shareWithDataDir does nothing on Windows: the lock file inherits the data
// directory's ACL.
func shareWithDataDir(string, string) error { return nil }

// copyOwner does nothing on Windows: a new file inherits its directory's ACL,
// which is the data directory's own.
func copyOwner(*os.Root, string, fs.FileInfo) error { return nil }
