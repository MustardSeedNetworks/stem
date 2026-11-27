// SPDX-License-Identifier: BUSL-1.1

package auth

import (
	"os"
	"testing"

	"github.com/MustardSeedNetworks/foundation/pkg/password"
)

// TestMain lowers the Argon2id cost for the whole package: the full 64 MiB
// profile on every hash and verify makes the race job outrun its timeout.
func TestMain(m *testing.M) {
	password.UseTestCost()
	os.Exit(m.Run())
}
