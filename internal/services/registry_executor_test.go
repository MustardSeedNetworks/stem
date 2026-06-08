// SPDX-License-Identifier: BUSL-1.1

package services_test

import (
	"testing"

	"github.com/krisarmstrong/stem/internal/services"
)

// TestRegistryExecutorFactories pins the single-source-of-truth invariant: every
// executable module is registered together with its executor factory (so adding
// a module no longer means editing a parallel map in the API layer), and the
// reflector is registered metadata-only because it has a distinct lifecycle.
func TestRegistryExecutorFactories(t *testing.T) {
	reg := services.DefaultRegistry()

	for _, name := range []string{"benchmark", "servicetest", "trafficgen", "measure", "certify"} {
		if reg.Get(name) == nil {
			t.Errorf("module %q is not registered", name)
		}
		if _, ok := reg.Factory(name); !ok {
			t.Errorf("module %q has no executor factory in the registry (single-source-of-truth broken)", name)
		}
	}

	if reg.Get("reflector") == nil {
		t.Error("reflector module is not registered")
	}
	if _, ok := reg.Factory("reflector"); ok {
		t.Error("reflector must NOT have an executor factory — it has a distinct lifecycle")
	}
}
