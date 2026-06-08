// SPDX-License-Identifier: BUSL-1.1

package services

import (
	"github.com/krisarmstrong/stem/internal/services/benchmark"
	"github.com/krisarmstrong/stem/internal/services/certify"
	"github.com/krisarmstrong/stem/internal/services/measure"
	"github.com/krisarmstrong/stem/internal/services/modtypes"
	"github.com/krisarmstrong/stem/internal/services/reflector"
	"github.com/krisarmstrong/stem/internal/services/servicetest"
	"github.com/krisarmstrong/stem/internal/services/trafficgen"
)

func DefaultRegistry() *Registry {
	return buildDefaultRegistry()
}

func buildDefaultRegistry() *Registry {
	reg := NewRegistry()

	// Register all modules with their executor factories in one place — adding a
	// module no longer means editing a separate factory map in the API layer.
	// Order: Reflector (Tier 1), then active testing modules (Tier 2).
	// Reflector has a distinct lifecycle (no Executor) so it registers metadata-only.
	reg.Register(reflector.New())
	reg.RegisterExecutable(benchmark.New(),
		func(iface string) (modtypes.Executor, error) { return benchmark.NewExecutor(iface) })
	reg.RegisterExecutable(servicetest.New(),
		func(iface string) (modtypes.Executor, error) { return servicetest.NewExecutor(iface) })
	reg.RegisterExecutable(trafficgen.New(),
		func(iface string) (modtypes.Executor, error) { return trafficgen.NewExecutor(iface) })
	reg.RegisterExecutable(measure.New(),
		func(iface string) (modtypes.Executor, error) { return measure.NewExecutor(iface) })
	reg.RegisterExecutable(certify.New(),
		func(iface string) (modtypes.Executor, error) { return certify.NewExecutor(iface) })

	return reg
}

// Factory returns the executor factory for a module from the default registry,
// or (nil, false) if the module is unknown or has no executor (e.g. reflector).
func Factory(name string) (modtypes.ExecutorFactory, bool) {
	return buildDefaultRegistry().Factory(name)
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
