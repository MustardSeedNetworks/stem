// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/krisarmstrong/stem/internal/logging"
	reflectorDP "github.com/krisarmstrong/stem/internal/reflector/dataplane"
	"github.com/krisarmstrong/stem/internal/services/reflector"
)

// Rate calculation constants.
const (
	bitsPerByte      = 8.0
	bitsPerMegabit   = 1000000.0
	defaultZeroRatio = 0
)

// handleReflectorConfig handles reflector configuration.
func (s *Server) handleReflectorConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.statsMu.RLock()
		cfg := s.reflectorConfig
		s.statsMu.RUnlock()
		writeJSON(w, cfg)
	case http.MethodPost:
		s.handleReflectorConfigUpdate(w, r)
	default:
		WriteMethodNotAllowed(w)
	}
}

// handleReflectorConfigUpdate handles POST requests to update reflector configuration.
func (s *Server) handleReflectorConfigUpdate(w http.ResponseWriter, r *http.Request) {
	var cfg ReflectorConfig
	if !decodeJSONStrict(w, r, &cfg) {
		return
	}

	err := s.validateReflectorProfile(cfg.Profile)
	if err != nil {
		logging.Warn("reflector config update failed: invalid profile", "profile", cfg.Profile)
		WriteInvalidRequest(w, "Invalid reflector profile")
		return
	}

	changes, dpUpdate, err := s.buildReflectorConfigUpdate(&cfg)
	if err != nil {
		logging.Warn("reflector config update failed", "error", err)
		WriteInvalidRequest(w, "Invalid configuration values")
		return
	}

	err = s.applyReflectorDataplaneUpdate(dpUpdate, changes)
	if err != nil {
		logging.Error("failed to update reflector dataplane config", "error", err)
		WriteInternalError(w, err)
		return
	}

	if len(changes) > 0 {
		logging.Info("reflector config updated", "changes", changes)
	}

	writeJSON(w, StatusResponse{Status: "updated"})
}

// validateReflectorProfile validates the reflector profile value.
func (s *Server) validateReflectorProfile(profile string) error {
	if profile == "" {
		return nil
	}
	validProfiles := map[string]bool{"netally": true, "msn": true, "all": true, "custom": true}
	if !validProfiles[profile] {
		return fmt.Errorf("invalid profile: %s", profile)
	}
	return nil
}

// buildReflectorConfigUpdate builds the config update and returns changes list.
func (s *Server) buildReflectorConfigUpdate(cfg *ReflectorConfig) ([]string, *reflectorDP.ConfigUpdate, error) {
	var changes []string

	s.statsMu.Lock()
	defer s.statsMu.Unlock()

	exec := s.reflectorExec

	var dpUpdate *reflectorDP.ConfigUpdate
	if exec != nil {
		dpUpdate = &reflectorDP.ConfigUpdate{
			Port:            nil,
			FilterOUI:       nil,
			OUI:             nil,
			FilterMAC:       nil,
			Mode:            nil,
			SignatureFilter: nil,
		}
	}

	if cfg.Profile != "" {
		s.reflectorConfig.Profile = cfg.Profile
		changes = append(changes, "profile="+cfg.Profile)
		if dpUpdate != nil {
			mode := cfg.Profile
			if mode == "netally" || mode == "msn" {
				mode = "all"
			}
			dpUpdate.Mode = &mode
		}
	}

	if cfg.OUIFilter != "" {
		s.reflectorConfig.OUIFilter = cfg.OUIFilter
		changes = append(changes, "ouiFilter="+cfg.OUIFilter)
		if dpUpdate != nil {
			dpUpdate.OUI = &cfg.OUIFilter
			filterOUI := true
			dpUpdate.FilterOUI = &filterOUI
		}
	}

	if cfg.PortFilter > 0 {
		safePort, ok := safeIntToUint16(cfg.PortFilter)
		if !ok {
			return nil, nil, fmt.Errorf("port %d out of valid range (0-%d)", cfg.PortFilter, math.MaxUint16)
		}
		s.reflectorConfig.PortFilter = cfg.PortFilter
		changes = append(changes, fmt.Sprintf("portFilter=%d", cfg.PortFilter))
		if dpUpdate != nil {
			dpUpdate.Port = &safePort
		}
	}

	if cfg.SignatureFilter != nil {
		s.reflectorConfig.SignatureFilter = cfg.SignatureFilter
		changes = append(changes, "signatureFilter updated")
		if dpUpdate != nil && len(cfg.SignatureFilter) > 0 {
			sigFilter := cfg.SignatureFilter[0]
			dpUpdate.SignatureFilter = &sigFilter
		}
	}

	return changes, dpUpdate, nil
}

// applyReflectorDataplaneUpdate applies the config update to the dataplane.
func (s *Server) applyReflectorDataplaneUpdate(dpUpdate *reflectorDP.ConfigUpdate, changes []string) error {
	s.statsMu.RLock()
	exec := s.reflectorExec
	s.statsMu.RUnlock()

	if exec == nil || dpUpdate == nil {
		return nil
	}

	dp := exec.Dataplane()
	if dp != nil {
		err := dp.UpdateConfig(dpUpdate)
		if err != nil {
			return fmt.Errorf("update dataplane config: %w", err)
		}
		logging.Info("reflector dataplane config updated", "changes", changes)
	}
	return nil
}

// handleReflectorStats returns reflector-specific statistics.
func (s *Server) handleReflectorStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
		return
	}

	s.statsMu.RLock()
	exec := s.reflectorExec
	elapsed := time.Since(s.startTime).Seconds()
	s.statsMu.RUnlock()

	// If we have an active reflector executor, get stats from it.
	if exec != nil {
		writeJSON(w, s.buildActiveReflectorStats(exec, elapsed))
		return
	}

	// Fallback: return stats from internal counters (for non-CGO builds or when reflector not active).
	writeJSON(w, s.buildFallbackReflectorStats(elapsed))
}

// buildActiveReflectorStats builds stats from the active reflector executor.
func (s *Server) buildActiveReflectorStats(exec *reflector.Executor, elapsed float64) ReflectorStats {
	dpStats := exec.GetStats()

	// Calculate rates based on uptime.
	ratePPS := float64(defaultZeroRatio)
	rateMbps := float64(defaultZeroRatio)
	if elapsed > 0 {
		ratePPS = float64(dpStats.PacketsReflected) / elapsed
		rateMbps = float64(dpStats.BytesReflected) * bitsPerByte / (elapsed * bitsPerMegabit)
	}

	return ReflectorStats{
		Running:          exec.IsRunning(),
		PacketsReceived:  dpStats.PacketsReceived,
		PacketsReflected: dpStats.PacketsReflected,
		BytesReceived:    dpStats.BytesReceived,
		BytesReflected:   dpStats.BytesReflected,
		TxErrors:         dpStats.TxErrors,
		RxInvalid:        dpStats.RxInvalid,
		RatePPS:          ratePPS,
		RateMbps:         rateMbps,
		Signatures: struct {
			ProbeOT uint64 `json:"probeot"`
			DataOT  uint64 `json:"dataot"`
			Latency uint64 `json:"latency"`
			RFC2544 uint64 `json:"rfc2544"`
			Y1564   uint64 `json:"y1564"`
			MSN     uint64 `json:"msn"`
		}{
			ProbeOT: dpStats.SigProbeOT,
			DataOT:  dpStats.SigDataOT,
			Latency: dpStats.SigLatency,
			RFC2544: dpStats.SigRFC2544,
			Y1564:   dpStats.SigY1564,
			MSN:     dpStats.SigMSN,
		},
		Latency: struct {
			MinUs float64 `json:"minUs"`
			AvgUs float64 `json:"avgUs"`
			MaxUs float64 `json:"maxUs"`
			Count uint64  `json:"count"`
		}{
			MinUs: dpStats.LatencyMin,
			AvgUs: dpStats.LatencyAvg,
			MaxUs: dpStats.LatencyMax,
			Count: dpStats.LatencyCount,
		},
		Uptime: elapsed,
	}
}

// buildFallbackReflectorStats builds stats from internal counters when executor is unavailable.
func (s *Server) buildFallbackReflectorStats(elapsed float64) ReflectorStats {
	s.statsMu.RLock()
	defer s.statsMu.RUnlock()

	pps := float64(defaultZeroRatio)
	mbps := float64(defaultZeroRatio)
	if elapsed > 0 && s.stats.PacketsSent > 0 {
		pps = float64(s.stats.PacketsSent) / elapsed
		mbps = float64(s.stats.BytesSent) * bitsPerByte / (elapsed * bitsPerMegabit)
	}

	return ReflectorStats{
		Running:          s.mode == moduleReflector && s.testStatus == statusRunning,
		PacketsReceived:  s.stats.PacketsReceived,
		PacketsReflected: s.stats.PacketsSent,
		BytesReceived:    s.stats.BytesReceived,
		BytesReflected:   s.stats.BytesSent,
		TxErrors:         0,
		RxInvalid:        0,
		RatePPS:          pps,
		RateMbps:         mbps,
		Signatures: struct {
			ProbeOT uint64 `json:"probeot"`
			DataOT  uint64 `json:"dataot"`
			Latency uint64 `json:"latency"`
			RFC2544 uint64 `json:"rfc2544"`
			Y1564   uint64 `json:"y1564"`
			MSN     uint64 `json:"msn"`
		}{
			ProbeOT: 0,
			DataOT:  0,
			Latency: 0,
			RFC2544: 0,
			Y1564:   0,
			MSN:     0,
		},
		Latency: struct {
			MinUs float64 `json:"minUs"`
			AvgUs float64 `json:"avgUs"`
			MaxUs float64 `json:"maxUs"`
			Count uint64  `json:"count"`
		}{
			MinUs: 0,
			AvgUs: 0,
			MaxUs: 0,
			Count: 0,
		},
		Uptime: elapsed,
	}
}
