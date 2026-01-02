// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package modules

import (
	"sync"
)

// Registry provides lookup and management of all registered modules.
type Registry struct {
	mu        sync.RWMutex
	modules   map[string]Module
	testIndex map[string]Module // testType -> module
}

// NewRegistry creates a new empty module registry.
func NewRegistry() *Registry {
	return &Registry{
		mu:        sync.RWMutex{},
		modules:   make(map[string]Module),
		testIndex: make(map[string]Module),
	}
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
//
//nolint:ireturn // Returns Module interface by design for polymorphism.
func (r *Registry) Get(name string) Module {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.modules[name]
}

// ModuleForTest returns the module that can execute the given test type.
// Returns nil if no module handles this test type.
//
//nolint:ireturn // Returns Module interface by design for polymorphism.
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
