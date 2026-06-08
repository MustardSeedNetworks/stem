// SPDX-License-Identifier: BUSL-1.1

package services

import (
	"sync"

	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
)

// Registry provides lookup and management of all registered modules and their
// executor factories — the single source of truth for both module metadata and
// (for executable modules) how to construct an executor.
type Registry struct {
	mu        sync.RWMutex
	modules   map[string]Module
	testIndex map[string]Module                   // testType -> module
	factories map[string]modtypes.ExecutorFactory // moduleName -> executor factory
}

// NewRegistry creates a new empty module registry.
func NewRegistry() *Registry {
	return &Registry{
		mu:        sync.RWMutex{},
		modules:   make(map[string]Module),
		testIndex: make(map[string]Module),
		factories: make(map[string]modtypes.ExecutorFactory),
	}
}

// RegisterExecutable registers a module together with the factory that builds
// its executor, so metadata and execution are declared in one place. Use
// Register for modules with a non-standard execution lifecycle (e.g. reflector).
func (r *Registry) RegisterExecutable(m Module, f modtypes.ExecutorFactory) {
	r.Register(m)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[m.Name()] = f
}

// Factory returns the executor factory for a module, or (nil, false) if the
// module is not registered or has no executor (e.g. reflector).
func (r *Registry) Factory(name string) (modtypes.ExecutorFactory, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.factories[name]
	return f, ok
}

// Register adds a module to the registry.
// It also indexes all test types the module can execute.
func (r *Registry) Register(m Module) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.modules[m.Name()] = m
	for _, testType := range m.TestTypes() {
		r.testIndex[testType] = m
	}
}

// Get returns a module by name, or nil if not found.
func (r *Registry) Get(name string) Module {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.modules[name]
}

// ModuleForTest returns the module that can execute the given test type.
// Returns nil if no module handles this test type.
func (r *Registry) ModuleForTest(testType string) Module {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.testIndex[testType]
}

// AllModules returns a slice of all registered modules.
func (r *Registry) AllModules() []Module {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Module, 0, len(r.modules))
	for _, m := range r.modules {
		result = append(result, m)
	}
	return result
}

// AllTestTypes returns all test types across all modules.
func (r *Registry) AllTestTypes() []TestType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]TestType, 0, len(r.testIndex))
	for testType, m := range r.testIndex {
		result = append(result, TestType{
			Name:        testType,
			Description: "",
			Standard:    m.Standard(),
			ModuleName:  m.Name(),
		})
	}
	return result
}

// ModuleCount returns the number of registered modules.
func (r *Registry) ModuleCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.modules)
}

// TestCount returns the total number of registered test types.
func (r *Registry) TestCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.testIndex)
}
