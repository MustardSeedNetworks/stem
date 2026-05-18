// SPDX-License-Identifier: BUSL-1.1

package license_test

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Use fast bcrypt for all license package tests.
	os.Setenv("STEM_TEST_MODE", "1")
	os.Exit(m.Run())
}
