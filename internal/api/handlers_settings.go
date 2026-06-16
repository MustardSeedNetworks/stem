// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"net/http"

	"github.com/MustardSeedNetworks/stem/internal/api/sse"
	"github.com/MustardSeedNetworks/stem/internal/logging"
	"github.com/MustardSeedNetworks/stem/internal/netif"
	reflectorDP "github.com/MustardSeedNetworks/stem/internal/reflector/dataplane"
)

// reflectorAvailabilityFn is a swappable platform-capability probe so
// tests can simulate macOS / Windows builds without rebuilding with
// different tags. Production code uses [reflectorDP.Available]; tests
// override via [Server.UseReflectorAvailabilityForTest].
type reflectorAvailabilityFn func() (available bool, reason string)

func defaultReflectorAvailability() (bool, string) {
	if reflectorDP.Available() {
		return true, ""
	}
	return false, reflectorDP.UnsupportedReason()
}

// handleSettings handles settings get/update.
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.statsMu.RLock()
		mode := s.mode
		iface := s.selectedIface
		s.statsMu.RUnlock()

		writeJSON(w, SettingsResponse{
			Mode:      mode,
			Interface: iface,
			Theme:     "system",
		})

	case http.MethodPost:
		var update SettingsUpdate
		if !decodeJSONStrict(w, r, &update, maxRequestBodySize) {
			return
		}

		if update.Interface != "" {
			// Validate that the interface exists.
			if !validateInterfaceExists(w, update.Interface) {
				return
			}

			s.statsMu.Lock()
			s.selectedIface = update.Interface
			s.statsMu.Unlock()
			logging.Info("interface selected", "interface", update.Interface)
		}

		writeJSON(w, StatusResponse{Status: "updated"})

	default:
		WriteMethodNotAllowed(w)
	}
}

// handleMode handles mode switching between reflector and test_master.
//
// GET returns the current mode. POST validates the request body and
// switches modes, performing the role-change side effects in this
// order:
//
//  1. Reject unknown mode values (400).
//  2. If the binary cannot support the requested mode on this platform
//     (e.g. reflector on macOS / Windows pure-Go builds), reject with
//     403 and a reason from the dataplane availability probe — the
//     same probe /api/v1/capabilities uses.
//  3. No-op if the requested mode equals the current mode: return 200
//     with previous == new so clients can render consistently without
//     any side effects.
//  4. Tear down any running role-bound work before swapping the mode
//     flag: stop the active reflector executor (if any) and cancel any
//     running / starting test. This mirrors the user-facing semantics
//     described by the RoleChip confirm dialog ("any in-progress test
//     will be cancelled").
//  5. Update [Server.mode] under the stats lock.
//  6. Emit an api.mode.changed audit log line with previous, new, and
//     caller IP so operators can trace role flips.
//
// Returns a [ModeUpdateResponse] with the new mode plus the previous
// mode it replaced.
func (s *Server) handleMode(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.statsMu.RLock()
		mode := s.mode
		s.statsMu.RUnlock()
		writeJSON(w, ModeResponse{Mode: mode})

	case http.MethodPost:
		s.handleModeUpdate(w, r)

	default:
		WriteMethodNotAllowed(w)
	}
}

// handleModeUpdate is the POST branch of [Server.handleMode], split out
// to keep the dispatch handler well within the cyclomatic-complexity
// budget enforced by golangci-lint.
func (s *Server) handleModeUpdate(w http.ResponseWriter, r *http.Request) {
	var req ModeRequest
	if !decodeJSONStrict(w, r, &req, maxRequestBodySize) {
		return
	}

	if req.Mode != modeReflector && req.Mode != modeTestMaster {
		logging.Warn("mode update failed: invalid mode", "mode", req.Mode)
		WriteInvalidRequest(w, "Invalid mode (must be 'reflector' or 'test_master')")
		return
	}

	// Platform gate: reject modes the binary cannot actually run.
	// Today this is reflector on non-CGO / non-Linux builds. The
	// same probe powers /api/v1/capabilities so the UI gets a
	// consistent answer no matter which endpoint it asks.
	availability := s.reflectorAvailability
	if availability == nil {
		availability = defaultReflectorAvailability
	}
	if req.Mode == modeReflector {
		available, reason := availability()
		if !available {
			if reason == "" {
				reason = "Reflector mode is not supported on this platform"
			}
			logging.Warn("mode update rejected: reflector unavailable",
				"mode", req.Mode,
				"reason", reason,
			)
			WriteError(w, &Error{
				HTTPStatus:  http.StatusForbidden,
				Code:        ErrCodePermissionDenied,
				Message:     reason,
				InternalErr: nil,
			})
			return
		}
	}

	s.statsMu.RLock()
	oldMode := s.mode
	s.statsMu.RUnlock()

	// No-op same-mode: short-circuit before tearing anything down.
	// The audit trail still logs nothing because nothing changed.
	if oldMode == req.Mode {
		resp := ModeUpdateResponse{
			Status:   "unchanged",
			Mode:     req.Mode,
			Previous: oldMode,
		}
		// Broadcast even on "unchanged" so a browser tab that missed
		// the original change still gets a confirmation frame.
		s.sseBroadcaster.Publish(sse.Frame{Type: "mode_changed", Payload: resp})
		writeJSON(w, resp)
		return
	}

	// Side effects of a real role change: stop the active reflector
	// (if any) and cancel any running / starting test. We do this
	// before flipping the mode flag so observers (e.g. the stats
	// handler) never see a window where the mode says "test_master"
	// while the reflector executor is still alive.
	s.teardownForModeSwitch()

	s.statsMu.Lock()
	s.mode = req.Mode
	s.statsMu.Unlock()

	logging.Info("api.mode.changed",
		"previous", oldMode,
		"new", req.Mode,
		"caller_ip", logging.GetClientIP(r),
	)

	resp := ModeUpdateResponse{
		Status:   "updated",
		Mode:     req.Mode,
		Previous: oldMode,
	}
	// Push to every connected SSE subscriber so other browser tabs /
	// CLI watchers see the mode change in real time (#296).
	s.sseBroadcaster.Publish(sse.Frame{Type: "mode_changed", Payload: resp})
	writeJSON(w, resp)
}

// teardownForModeSwitch stops the running reflector (if any) and
// cancels any in-progress test. Called from the mode-update path
// before swapping s.mode so the binary never serves a request from an
// inconsistent state (e.g. mode=test_master with a live reflector
// executor still attached).
//
// This is intentionally narrower than [Server.Shutdown]: it does not
// touch the rate limiters, auth manager, or HTTP server. It only
// reverses the side effects of the previous role.
func (s *Server) teardownForModeSwitch() {
	s.statsMu.Lock()
	exec := s.reflectorExec
	wasRunning := s.testStatus == statusRunning || s.testStatus == statusStarting
	if wasRunning {
		s.testStatus = statusCancelled
		s.currentTest = ""
		s.currentModule = ""
	}
	if exec != nil {
		s.reflectorExec = nil
	}
	s.statsMu.Unlock()

	if exec != nil {
		logging.Info("stopping reflector for mode switch")
		exec.Stop()
	}
	if wasRunning {
		logging.Info("test cancelled for mode switch")
	}
}

// validateInterfaceExists checks if an interface exists and returns false if not (writing error to response).
func validateInterfaceExists(w http.ResponseWriter, ifaceName string) bool {
	ifaces, detectErr := netif.DetectInterfaces()
	if detectErr != nil {
		logging.Error("failed to detect interfaces for validation", "error", detectErr)
		WriteInternalError(w, detectErr)
		return false
	}

	for _, iface := range ifaces {
		if iface.Name == ifaceName {
			return true
		}
	}

	WriteInvalidRequest(w, "Specified network interface not found")
	return false
}
