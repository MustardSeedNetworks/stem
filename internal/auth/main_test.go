// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package auth

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Use fast bcrypt for all auth package tests.
	os.Setenv("STEM_TEST_MODE", "1")
	os.Exit(m.Run())
}
