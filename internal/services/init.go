// SPDX-License-Identifier: BUSL-1.1

package services

import (
	"github.com/krisarmstrong/stem/internal/services/benchmark"
	"github.com/krisarmstrong/stem/internal/services/certify"
	"github.com/krisarmstrong/stem/internal/services/measure"
	"github.com/krisarmstrong/stem/internal/services/reflector"
	"github.com/krisarmstrong/stem/internal/services/servicetest"
	"github.com/krisarmstrong/stem/internal/services/trafficgen"
)

func DefaultRegistry() *Registry {
	return buildDefaultRegistry()
}

func buildDefaultRegistry() *Registry {
	reg := NewRegistry()

	// Register all modules.
	// Order: Reflector (Tier 1), then active testing modules (Tier 2).
	reg.Register(reflector.New())
	reg.Register(benchmark.New())
	reg.Register(servicetest.New())
	reg.Register(trafficgen.New())
	reg.Register(measure.New())
	reg.Register(certify.New())

	return reg
}

// GetModule returns a module by name from the default registry.
func GetModule(name string) Module {
	return buildDefaultRegistry().Get(name)
}

// GetModuleForTest returns the module that handles a given test type.
func GetModuleForTest(testType string) Module {
	return buildDefaultRegistry().ModuleForTest(testType)
}

// GetAllModules returns all registered modules.
func GetAllModules() []Module {
	return buildDefaultRegistry().AllModules()
}

// GetAllModuleInfos returns all modules as API-friendly ModuleInfo structs.
func GetAllModuleInfos() []ModuleInfo {
	mods := buildDefaultRegistry().AllModules()
	infos := make([]ModuleInfo, len(mods))
	for i, m := range mods {
		infos[i] = ToInfo(m)
	}
	return infos
}
