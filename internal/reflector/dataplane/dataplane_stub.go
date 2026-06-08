//go:build !cgo || !linux

// Package dataplane provides CGO bindings to the C reflector dataplane.
//
// This file contains stub implementations for non-CGO or non-Linux builds.
// The actual dataplane requires CGO and Linux for AF_PACKET/AF_XDP support.
package dataplane

import (
	"errors"

	"github.com/MustardSeedNetworks/stem/internal/reflector/config"
)

// Stats holds reflector packet statistics.
type Stats struct {
	PacketsReceived  uint64  // Number of packets received.
	PacketsReflected uint64  // Number of packets reflected.
	BytesReceived    uint64  // Number of bytes received.
	BytesReflected   uint64  // Number of bytes reflected.
	TxErrors         uint64  // Number of transmit errors.
	RxInvalid        uint64  // Number of invalid packets received.
	SigProbeOT       uint64  // Count of PROBEOT signatures.
	SigDataOT        uint64  // Count of DATA:OT signatures.
	SigLatency       uint64  // Count of LATENCY signatures.
	SigRFC2544       uint64  // Count of RFC2544 signatures.
	SigY1564         uint64  // Count of Y.1564 signatures.
	SigMSN           uint64  // Count of MSN signatures.
	LatencyMin       float64 // Minimum latency in microseconds.
	LatencyAvg       float64 // Average latency in microseconds.
	LatencyMax       float64 // Maximum latency in microseconds.
	LatencyCount     uint64  // Number of latency samples.
}

// ConfigUpdate holds optional configuration updates.
// Only non-nil fields are applied when passed to UpdateConfig.
type ConfigUpdate struct {
	Port            *uint16 // UDP port filter.
	FilterOUI       *bool   // Enable OUI filtering.
	OUI             *string // OUI value (e.g., "00:c0:17").
	FilterMAC       *bool   // Enable MAC filtering.
	Mode            *string // Reflection mode: "mac", "mac-ip", "all".
	SignatureFilter *string // Signature filter: "all", "ito", "rfc2544", etc.
}

// Dataplane is the reflector packet processing engine.
type Dataplane struct {
	cfg *config.Config // Placeholder for CGO implementation.
}

// ErrNotSupported is returned when CGO dataplane is not available.
var ErrNotSupported = errors.New("CGO dataplane not available on this platform")

// unsupportedReasonText is the operator-facing reason surfaced via the
// /api/v1/capabilities endpoint and any platform-guard UX. Industry
// terms ("CGO", "Linux") are kept verbatim so they read identically
// across locales.
const unsupportedReasonText = "CGO + Linux required"

// Available reports whether the CGO + Linux reflector dataplane is
// compiled into this binary. The stub build always returns false; the
// real build returns true. Callers use this to gate UX rather than
// waiting for a Start failure to surface
// "CGO dataplane not available on this platform".
func Available() bool {
	return false
}

// UnsupportedReason returns a short, operator-facing reason describing
// why the reflector dataplane is unavailable on this binary. Empty
// string when [Available] returns true.
func UnsupportedReason() string {
	return unsupportedReasonText
}

// New creates a new dataplane instance (stub).
func New(_ *config.Config) (*Dataplane, error) {
	return nil, ErrNotSupported
}

// Start begins packet reflection (stub).
func (dp *Dataplane) Start() error {
	return ErrNotSupported
}

// Stop halts packet reflection (stub).
func (dp *Dataplane) Stop() {}

// Close releases dataplane resources (stub).
func (dp *Dataplane) Close() {}

// GetStats returns current statistics (stub).
func (dp *Dataplane) GetStats() Stats {
	return Stats{
		PacketsReceived:  0,
		PacketsReflected: 0,
		BytesReceived:    0,
		BytesReflected:   0,
		TxErrors:         0,
		RxInvalid:        0,
		SigProbeOT:       0,
		SigDataOT:        0,
		SigLatency:       0,
		SigRFC2544:       0,
		SigY1564:         0,
		SigMSN:           0,
		LatencyMin:       0,
		LatencyAvg:       0,
		LatencyMax:       0,
		LatencyCount:     0,
	}
}

// IsRunning returns whether the dataplane is active (stub).
func (dp *Dataplane) IsRunning() bool {
	return false
}

// Interface returns the configured network interface.
func (dp *Dataplane) Interface() string {
	if dp != nil && dp.cfg != nil {
		return dp.cfg.Interface
	}
	return ""
}

// Config returns the current configuration.
func (dp *Dataplane) Config() *config.Config {
	if dp != nil {
		return dp.cfg
	}
	return nil
}

// UpdateConfig applies configuration changes to the dataplane (stub).
func (dp *Dataplane) UpdateConfig(_ *ConfigUpdate) error {
	return ErrNotSupported
}

// ResetStats clears all statistics counters (stub).
func (dp *Dataplane) ResetStats() {}
