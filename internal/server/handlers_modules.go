// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/krisarmstrong/stem/internal/modules"
)

// handleModules returns the list of all modules.
func (s *Server) handleModules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	moduleInfos := modules.GetAllModuleInfos()
	writeJSON(w, map[string]any{
		"modules": moduleInfos,
		"count":   len(moduleInfos),
	})
}

// handleModuleByName handles requests for specific modules.
// Supports: GET /api/modules/{name} and GET /api/modules/{name}/tests.
func (s *Server) handleModuleByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse path: /api/modules/{name} or /api/modules/{name}/tests.
	path := strings.TrimPrefix(r.URL.Path, "/api/modules/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Module name required", http.StatusBadRequest)
		return
	}

	moduleName := parts[0]
	module := modules.GetModule(moduleName)

	if module == nil {
		http.Error(w, fmt.Sprintf("Module not found: %s", moduleName), http.StatusNotFound)
		return
	}

	// Check for /tests subpath.
	if len(parts) > 1 && parts[1] == "tests" {
		// Return just the test types for this module.
		writeJSON(w, map[string]any{
			"module": moduleName,
			"tests":  module.TestTypes(),
			"count":  len(module.TestTypes()),
		})
		return
	}

	// Return full module info.
	writeJSON(w, modules.ToInfo(module))
}
