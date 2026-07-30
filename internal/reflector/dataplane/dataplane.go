//go:build cgo && linux

// Package dataplane provides CGO bindings to the C reflector dataplane.
//
// This package wraps the high-performance C dataplane library, which uses
// AF_PACKET, AF_XDP, or DPDK for line-rate packet reflection.
package dataplane

/*
#cgo CFLAGS: -I${SRCDIR}/../../../include
#cgo LDFLAGS: -L${SRCDIR}/../../../build -lreflector
#cgo linux LDFLAGS: -lxdp -lbpf -lelf -lz

#include "reflector.h"
#include <stdlib.h>
#include <string.h>

// Helper to create config
static reflector_config_t make_config(
    const char *ifname,
    uint16_t ito_port,
    int filter_oui,
    uint8_t oui0, uint8_t oui1, uint8_t oui2,
    int reflect_mode,
    int use_dpdk,
    int use_af_xdp,
    const char *dpdk_args
) {
    reflector_config_t config = {0};
    config.ito_port = ito_port;
    config.filter_oui = filter_oui ? true : false;
    config.oui[0] = oui0;
    config.oui[1] = oui1;
    config.oui[2] = oui2;
    config.reflect_mode = (reflect_mode_t)reflect_mode;
    config.use_dpdk = use_dpdk ? true : false;
    config.use_af_xdp = use_af_xdp ? true : false;
    config.dpdk_args = (char *)dpdk_args;
    return config;
}
*/
import "C"

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/MustardSeedNetworks/stem/internal/reflector/config"
)

// Available reports whether the CGO + Linux reflector dataplane is
// compiled into this binary. The real (CGO + Linux) build always
// returns true; the stub build (non-Linux, or CGO disabled) returns
// false. Callers use this to gate UX rather than waiting for a Start
// failure to surface "CGO dataplane not available on this platform".
func Available() bool {
	return true
}

// UnsupportedReason returns a short, operator-facing reason describing
// why the reflector dataplane is unavailable. The real build returns an
// empty string; the stub build returns "CGO + Linux required".
func UnsupportedReason() string {
	return ""
}

// Stats holds dataplane statistics
type Stats struct {
	PacketsReceived  uint64
	PacketsReflected uint64
	BytesReceived    uint64
	BytesReflected   uint64
	TxErrors         uint64
	RxInvalid        uint64
	SigProbeOT       uint64
	SigDataOT        uint64
	SigLatency       uint64
	SigRFC2544       uint64
	SigY1564         uint64
	SigMSN           uint64
	LatencyMin       float64
	LatencyAvg       float64
	LatencyMax       float64
	LatencyCount     uint64
}

// ConfigUpdate holds optional configuration updates.
// Only non-nil fields are applied when passed to UpdateConfig.
type ConfigUpdate struct {
	Port            *uint16 // UDP port filter
	FilterOUI       *bool   // Enable OUI filtering
	OUI             *string // OUI value (e.g., "00:c0:17")
	FilterMAC       *bool   // Enable MAC filtering
	Mode            *string // Reflection mode: "mac", "mac-ip", "all"
	SignatureFilter *string // Signature filter: "all", "ito", "rfc2544", etc.
}

// Dataplane wraps the C reflector context
type Dataplane struct {
	ctx      C.reflector_ctx_t
	cfg      *config.Config
	running  bool
	closed   bool
	mu       sync.RWMutex
	dpdkArgs *C.char // Store to prevent dangling pointer
}

// New creates a new dataplane instance
func New(cfg *config.Config) (*Dataplane, error) {
	dp := &Dataplane{
		cfg: cfg,
	}

	// Parse OUI
	oui, err := cfg.ParseOUI()
	if err != nil {
		return nil, fmt.Errorf("failed to parse OUI: %w", err)
	}
	sigFilter, err := signatureFilterValue(cfg.SignatureFilter)
	if err != nil {
		return nil, err
	}

	// Create C config
	ifname := C.CString(cfg.Interface)
	defer C.free(unsafe.Pointer(ifname))

	var dpdkArgs *C.char
	if cfg.Platform.DPDKArgs != "" {
		dpdkArgs = C.CString(cfg.Platform.DPDKArgs)
		dp.dpdkArgs = dpdkArgs // Store for cleanup in Close()
	}

	filterOUI := 0
	if cfg.Filtering.FilterOUI {
		filterOUI = 1
	}

	useDPDK := 0
	if cfg.Platform.UseDPDK {
		useDPDK = 1
	}
	useAFXDP := 0
	if cfg.Platform.UseAFXDP {
		useAFXDP = 1
	}

	cConfig := C.make_config(
		ifname,
		C.uint16_t(cfg.Filtering.Port),
		C.int(filterOUI),
		C.uint8_t(oui[0]), C.uint8_t(oui[1]), C.uint8_t(oui[2]),
		C.int(cfg.ReflectModeInt()),
		C.int(useDPDK),
		C.int(useAFXDP),
		dpdkArgs,
	)

	// Initialize reflector
	if C.reflector_init(&dp.ctx, ifname) < 0 {
		return nil, fmt.Errorf("failed to initialize reflector on %s", cfg.Interface)
	}

	// Preserve interface and worker settings detected by reflector_init.
	dp.ctx.config.ito_port = cConfig.ito_port
	dp.ctx.config.filter_oui = cConfig.filter_oui
	dp.ctx.config.oui[0] = cConfig.oui[0]
	dp.ctx.config.oui[1] = cConfig.oui[1]
	dp.ctx.config.oui[2] = cConfig.oui[2]
	dp.ctx.config.filter_dst_mac = C.bool(cfg.Filtering.FilterMAC)
	dp.ctx.config.reflect_mode = cConfig.reflect_mode
	dp.ctx.config.sig_filter = C.sig_filter_t(sigFilter)
	dp.ctx.config.use_dpdk = cConfig.use_dpdk
	dp.ctx.config.use_af_xdp = cConfig.use_af_xdp
	dp.ctx.config.dpdk_args = cConfig.dpdk_args

	return dp, nil
}

func signatureFilterValue(filter string) (int, error) {
	if err := config.ValidateSignatureFilter(filter); err != nil {
		return 0, err
	}
	values := map[string]int{
		"all": 0, "ito": 1, "rfc2544": 2,
		"probeot": 6, "dataot": 7, "latency": 8,
		"y1564": 3, "custom": 4, "msn": 5,
	}
	return values[filter], nil
}

func reflectionModeValue(mode string) (int, error) {
	values := map[string]int{"mac": 0, "mac-ip": 1, "all": 2}
	value, ok := values[mode]
	if !ok {
		return 0, fmt.Errorf("invalid reflection mode %q", mode)
	}
	return value, nil
}

// Start begins packet processing
func (dp *Dataplane) Start() error {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	if dp.running {
		return fmt.Errorf("dataplane already running")
	}
	if dp.closed {
		return fmt.Errorf("dataplane is closed")
	}

	if C.reflector_start(&dp.ctx) < 0 {
		return fmt.Errorf("failed to start reflector")
	}

	dp.running = true
	return nil
}

// Stop halts packet processing
func (dp *Dataplane) Stop() {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	if !dp.running {
		return
	}

	C.reflector_stop(&dp.ctx)
	dp.running = false
}

// Close cleans up dataplane resources
func (dp *Dataplane) Close() {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	if dp.closed {
		return
	}
	if dp.running {
		C.reflector_stop(&dp.ctx)
		dp.running = false
	}
	C.reflector_cleanup(&dp.ctx)
	// Free stored C strings
	if dp.dpdkArgs != nil {
		C.free(unsafe.Pointer(dp.dpdkArgs))
		dp.dpdkArgs = nil
	}
	dp.closed = true
}

// GetStats returns current statistics
func (dp *Dataplane) GetStats() Stats {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	if dp.closed {
		return Stats{}
	}

	var cStats C.reflector_stats_t
	C.reflector_get_stats(&dp.ctx, &cStats)

	return Stats{
		PacketsReceived:  uint64(cStats.packets_received),
		PacketsReflected: uint64(cStats.packets_reflected),
		BytesReceived:    uint64(cStats.bytes_received),
		BytesReflected:   uint64(cStats.bytes_reflected),
		TxErrors:         uint64(cStats.tx_errors),
		RxInvalid:        uint64(cStats.rx_invalid),
		SigProbeOT:       uint64(cStats.sig_probeot_count),
		SigDataOT:        uint64(cStats.sig_dataot_count),
		SigLatency:       uint64(cStats.sig_latency_count),
		SigRFC2544:       uint64(cStats.sig_rfc2544_count),
		SigY1564:         uint64(cStats.sig_y1564_count),
		SigMSN:           uint64(cStats.sig_msn_count),
		LatencyMin:       float64(cStats.latency.min_ns) / 1000.0,
		LatencyAvg:       float64(cStats.latency.avg_ns) / 1000.0,
		LatencyMax:       float64(cStats.latency.max_ns) / 1000.0,
		LatencyCount:     uint64(cStats.latency.count),
	}
}

// IsRunning returns whether the dataplane is active
func (dp *Dataplane) IsRunning() bool {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	return dp.running
}

// Interface returns the network interface name. Safe on a zero-value
// receiver — matches the stub build's contract so tests that construct
// `&Dataplane{}` directly work under both CGO and non-CGO builds.
func (dp *Dataplane) Interface() string {
	if dp == nil || dp.cfg == nil {
		return ""
	}
	return dp.cfg.Interface
}

// Config returns the configuration
func (dp *Dataplane) Config() *config.Config {
	return dp.cfg
}

// UpdateConfig updates runtime configuration using typed ConfigUpdate struct.
// Only non-nil fields in the update are applied.
// Some settings take effect immediately, others require restart.
func (dp *Dataplane) UpdateConfig(update *ConfigUpdate) error {
	if update == nil {
		return nil
	}
	sigFilter, mode := 0, 0
	if update.SignatureFilter != nil {
		value, err := signatureFilterValue(*update.SignatureFilter)
		if err != nil {
			return err
		}
		sigFilter = value
	}
	if update.Mode != nil {
		value, err := reflectionModeValue(*update.Mode)
		if err != nil {
			return err
		}
		mode = value
	}
	var parsedOUI [3]byte
	if update.OUI != nil {
		candidate := config.Config{Filtering: config.FilterConfig{OUI: *update.OUI}}
		value, err := candidate.ParseOUI()
		if err != nil {
			return fmt.Errorf("invalid OUI format %q: %w", *update.OUI, err)
		}
		parsedOUI = value
	}

	dp.mu.Lock()
	defer dp.mu.Unlock()
	if dp.closed {
		return fmt.Errorf("dataplane is closed")
	}
	if dp.running {
		return fmt.Errorf("reflector configuration cannot change while running")
	}

	if update.Port != nil || update.SignatureFilter != nil {
		currentPort := dp.cfg.Filtering.Port
		if update.Port != nil {
			currentPort = *update.Port
		}
		currentSignatureFilter := sigFilter
		if update.SignatureFilter != nil {
			currentSignatureFilter = sigFilter
		} else {
			var currentErr error
			currentSignatureFilter, currentErr = signatureFilterValue(dp.cfg.SignatureFilter)
			if currentErr != nil {
				return currentErr
			}
		}
		if C.reflector_update_filter(
			&dp.ctx,
			C.uint16_t(currentPort),
			C.sig_filter_t(currentSignatureFilter),
		) != 0 {
			return fmt.Errorf("failed to update reflector packet filter")
		}
	}

	// Apply each non-nil field
	if update.Port != nil {
		dp.cfg.Filtering.Port = *update.Port
	}

	if update.FilterOUI != nil {
		dp.cfg.Filtering.FilterOUI = *update.FilterOUI
		dp.ctx.config.filter_oui = C.bool(*update.FilterOUI)
	}

	if update.OUI != nil {
		dp.cfg.Filtering.OUI = *update.OUI
		dp.ctx.config.oui[0] = C.uint8_t(parsedOUI[0])
		dp.ctx.config.oui[1] = C.uint8_t(parsedOUI[1])
		dp.ctx.config.oui[2] = C.uint8_t(parsedOUI[2])
	}

	if update.FilterMAC != nil {
		dp.cfg.Filtering.FilterMAC = *update.FilterMAC
		dp.ctx.config.filter_dst_mac = C.bool(*update.FilterMAC)
	}

	if update.Mode != nil {
		dp.cfg.Reflection.Mode = *update.Mode
		dp.ctx.config.reflect_mode = C.reflect_mode_t(mode)
	}

	if update.SignatureFilter != nil {
		dp.cfg.SignatureFilter = *update.SignatureFilter
	}

	return nil
}

// ResetStats resets the statistics counters
func (dp *Dataplane) ResetStats() {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	if dp.closed {
		return
	}
	C.reflector_reset_stats(&dp.ctx)
}
