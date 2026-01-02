// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package modules

import (
	"github.com/krisarmstrong/stem/internal/modules/benchmark"
	"github.com/krisarmstrong/stem/internal/modules/certify"
	"github.com/krisarmstrong/stem/internal/modules/measure"
	"github.com/krisarmstrong/stem/internal/modules/reflector"
	"github.com/krisarmstrong/stem/internal/modules/servicetest"
	"github.com/krisarmstrong/stem/internal/modules/trafficgen"
)

// DefaultRegistry is the global module registry with all modules pre-registered.
//
//nolint:gochecknoglobals // Singleton registry pattern for module system.
var DefaultRegistry *Registry

//nolint:gochecknoinits // Required for module registration at package load.
func init() {
	DefaultRegistry = NewRegistry()

	// Register all modules.
	// Order: Reflector (Tier 1), then active testing modules (Tier 2).
	DefaultRegistry.Register(reflector.New())
	DefaultRegistry.Register(benchmark.New())
	DefaultRegistry.Register(servicetest.New())
	DefaultRegistry.Register(trafficgen.New())
	DefaultRegistry.Register(measure.New())
	DefaultRegistry.Register(certify.New())
}

// GetModule returns a module by name from the default registry.
//
//nolint:ireturn // Returns Module interface by design for polymorphism.
func GetModule(name string) Module {
	return DefaultRegistry.Get(name)
}

// GetModuleForTest returns the module that handles a given test type.
//
//nolint:ireturn // Returns Module interface by design for polymorphism.
func GetModuleForTest(testType string) Module {
	return DefaultRegistry.ModuleForTest(testType)
}

// GetAllModules returns all registered modules.
func GetAllModules() []Module {
	return DefaultRegistry.AllModules()
}

// GetAllModuleInfos returns all modules as API-friendly ModuleInfo structs.
func GetAllModuleInfos() []ModuleInfo {
	mods := DefaultRegistry.AllModules()
	infos := make([]ModuleInfo, len(mods))
	for i, m := range mods {
		infos[i] = ToInfo(m)
	}
	return infos
}
