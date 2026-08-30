// SPDX-License-Identifier: BUSL-1.1

package certify_test

import (
	"errors"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/services/certify"
	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
)

// Every branch of Execute's switch dereferences the dataplane context. A nil
// one is harmless on the stub build and a panic on cgo+linux, so the contract
// has to be stated once and asserted -- not left to the platform.
func TestExecuteRejectsNilDataplaneContext(t *testing.T) {
	exec := certify.NewExecutorWithContext(nil)
	cfg := &modtypes.TestConfig{Interface: "lo", FrameSize: 64, Duration: 60}

	for _, testType := range []string{
		certify.TestRFC2889Forwarding,
		certify.TestRFC2889Caching,
		certify.TestRFC6349Throughput,
	} {
		t.Run(testType, func(t *testing.T) {
			result, err := exec.Execute(testType, cfg)
			if !errors.Is(err, modtypes.ErrInvalidConfig) {
				t.Fatalf("Execute with nil ctx: got %v, want ErrInvalidConfig", err)
			}
			if result != nil {
				t.Errorf("result = %+v, want nil", result)
			}
		})
	}
}
