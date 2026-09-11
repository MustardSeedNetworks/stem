// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"os"
	"testing"
)

// TestMain gives the package an isolated data directory. STEM_DATA_DIR
// defaults to the working directory, so a server built by a test writes its
// daemon state — the reflector configuration, the connection descriptor —
// into internal/api/ itself: one test then inherits another's state, and the
// files land in the developer's working tree.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "stem-api-test-data-*")
	if err != nil {
		panic("create test data directory: " + err.Error())
	}
	if setErr := os.Setenv("STEM_DATA_DIR", dir); setErr != nil {
		panic("set STEM_DATA_DIR: " + setErr.Error())
	}

	code := m.Run()

	_ = os.RemoveAll(dir)
	os.Exit(code)
}
