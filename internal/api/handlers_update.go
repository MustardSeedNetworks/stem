// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"net/http"

	"github.com/MustardSeedNetworks/stem/internal/logging"
	"github.com/MustardSeedNetworks/stem/internal/update"
)

// handleUpdateSettings reads and changes the automatic update opt-in, the
// Settings toggle of ADR-0012. The daemon only writes the persisted setting;
// the root-owned timer reads it and runs `stem update`, so turning it on
// grants the service nothing it could not already do.
func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		settings, err := update.LoadSettings(s.dataDir)
		if err != nil {
			WriteError(w, InternalError(err))
			return
		}
		writeJSON(w, updateSettingsOf(settings))

	case http.MethodPost:
		var req UpdateSettingsRequest
		if !decodeJSONStrict(w, r, &req, maxRequestBodySize) || !validateStruct(w, &req) {
			return
		}
		s.updateSettingsMu.Lock()
		defer s.updateSettingsMu.Unlock()
		settings, err := update.LoadSettings(s.dataDir)
		if err != nil {
			WriteError(w, InternalError(err))
			return
		}
		previous := settings.Automatic
		settings.Automatic = *req.Automatic
		if err = update.SaveSettings(s.dataDir, settings); err != nil {
			WriteError(w, InternalError(err))
			return
		}
		logging.Info("api.update.automatic.changed",
			"previous", previous,
			"new", settings.Automatic,
			"source", settings.SourceSpec(),
			"caller_ip", logging.GetClientIP(r),
		)
		writeJSON(w, updateSettingsOf(settings))

	default:
		WriteMethodNotAllowed(w)
	}
}

func updateSettingsOf(settings update.Settings) UpdateSettings {
	return UpdateSettings{Automatic: settings.Automatic, Source: settings.SourceSpec()}
}
