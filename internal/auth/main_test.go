// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Use fast bcrypt for all auth package tests.
	os.Setenv("STEM_TEST_MODE", "1")
	os.Exit(m.Run())
}
