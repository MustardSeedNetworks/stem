// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"net/http"

	"github.com/MustardSeedNetworks/stem/internal/logging"
	reflectorDP "github.com/MustardSeedNetworks/stem/internal/reflector/dataplane"
)

// defaultDataplaneAvailability reports whether this binary can put packets
// on the wire. The reflector and test dataplanes are compiled under the same
// `cgo && linux` constraint, so the reflector package's probe answers for
// both: a binary without one has neither.
func defaultDataplaneAvailability() (bool, string) {
	if reflectorDP.Available() {
		return true, ""
	}
	return false, reflectorDP.UnsupportedReason()
}

func (s *Server) dataplaneAvailable() (bool, string) {
	if s.dataplaneAvailability != nil {
		return s.dataplaneAvailability()
	}
	return defaultDataplaneAvailability()
}

// refuseWithoutDataplane writes a 403 carrying the capability's reason and
// returns true when this binary has no dataplane. Requests that can only end
// in a dataplane failure are refused before they change any state, so a
// refused reflector start saves no configuration and no autostart.
func (s *Server) refuseWithoutDataplane(w http.ResponseWriter, action string) bool {
	available, reason := s.dataplaneAvailable()
	if available {
		return false
	}
	logging.Warn(action+" rejected: dataplane unavailable", "reason", reason)
	WriteError(w, &Error{
		HTTPStatus: http.StatusForbidden,
		Code:       ErrCodePermissionDenied,
		Message:    reason,
	})
	return true
}
