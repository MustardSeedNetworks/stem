// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"net/http"

	"github.com/krisarmstrong/stem/internal/netif"
)

// handleInterfaces returns the list of network interfaces.
func (s *Server) handleInterfaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ifaces, err := netif.DetectInterfaces()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, ifaces)
}
