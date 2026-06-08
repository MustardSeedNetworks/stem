// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/krisarmstrong/stem/internal/services"
)

// handleModules returns the list of all modules.
func (s *Server) handleModules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	moduleInfos := services.GetAllModuleInfos()
	writeJSON(w, map[string]any{
		"modules": moduleInfos,
		"count":   len(moduleInfos),
	})
}

// handleModuleByName handles requests for specific modules.
// Supports: GET /api/v1/modules/{name} and GET /api/v1/modules/{name}/tests.
func (s *Server) handleModuleByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse path: /api/v1/modules/{name} or /api/v1/modules/{name}/tests.
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/modules/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Module name required", http.StatusBadRequest)
		return
	}

	moduleName := parts[0]
	module := services.GetModule(moduleName)

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
	writeJSON(w, services.ToInfo(module))
}
