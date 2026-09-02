//go:build cgo && linux

// Package dataplane provides CGO bindings to the C reflector dataplane.
//
// This package wraps the high-performance C dataplane library, which uses
// AF_PACKET or AF_XDP for line-rate packet reflection.
package dataplane

/*
#cgo CFLAGS: -I${SRCDIR}/../../../include
#cgo LDFLAGS: -L${SRCDIR}/../../../build -lreflector
#cgo linux LDFLAGS: -lxdp -lbpf -lelf -lz

#include "reflector.h"
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

// Helper to create config
static reflector_config_t make_config(
    const char *ifname,
    uint16_t ito_port,
    int filter_oui,
    uint8_t oui0, uint8_t oui1, uint8_t oui2,
    int reflect_mode,
    int use_af_xdp
) {
    reflector_config_t config = {0};
    config.ito_port = ito_port;
    config.filter_oui = filter_oui ? true : false;
    config.oui[0] = oui0;
    config.oui[1] = oui1;
    config.oui[2] = oui2;
    config.reflect_mode = (reflect_mode_t)reflect_mode;
    config.use_af_xdp = use_af_xdp ? true : false;
    return config;
}

static void stop_reflector(uintptr_t ctx) {
    reflector_stop((reflector_ctx_t *)ctx);
}

static int init_reflector(uintptr_t ctx, const char *ifname) {
    return reflector_init((reflector_ctx_t *)ctx, ifname);
}

static int start_reflector(uintptr_t ctx) {
    return reflector_start((reflector_ctx_t *)ctx);
}

static void cleanup_reflector(uintptr_t ctx) {
    reflector_cleanup((reflector_ctx_t *)ctx);
}

static void get_reflector_stats(uintptr_t ctx, uintptr_t stats) {
    reflector_get_stats((reflector_ctx_t *)ctx, (reflector_stats_t *)stats);
}

static int update_reflector_filter(uintptr_t ctx, uint16_t port, sig_filter_t filter) {
    return reflector_update_filter((reflector_ctx_t *)ctx, port, filter);
}

static void reset_reflector_stats(uintptr_t ctx) {
    reflector_reset_stats((reflector_ctx_t *)ctx);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"

	"github.com/MustardSeedNetworks/stem/internal/reflector/config"
)

const nanosecondsPerMicrosecond = 1000

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

// Stats holds dataplane statistics.
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

// Dataplane wraps the C reflector context.
type Dataplane struct {
	ctx     *C.reflector_ctx_t
	cfg     *config.Config
	running bool
	closed  bool
	mu      sync.RWMutex
}

// New creates a new dataplane instance.
func New(cfg *config.Config) (*Dataplane, error) {
	dp := &Dataplane{
		ctx: (*C.reflector_ctx_t)(C.calloc(1, C.size_t(C.sizeof_reflector_ctx_t))),
		cfg: cfg,
	}
	if dp.ctx == nil {
		return nil, errors.New("failed to allocate reflector context")
	}

	// Parse OUI
	oui, err := cfg.ParseOUI()
	if err != nil {
		C.free(unsafe.Pointer(dp.ctx))
		return nil, fmt.Errorf("failed to parse OUI: %w", err)
	}
	sigFilter, err := signatureFilterValue(cfg.SignatureFilter)
	if err != nil {
		C.free(unsafe.Pointer(dp.ctx))
		return nil, err
	}

	// Create C config
	ifname := C.CString(cfg.Interface)
	defer C.free(unsafe.Pointer(ifname))

	filterOUI := 0
	if cfg.Filtering.FilterOUI {
		filterOUI = 1
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
		C.int(useAFXDP),
	)

	// Initialize reflector
	ctx := C.uintptr_t(uintptr(unsafe.Pointer(dp.ctx)))
	if C.init_reflector(ctx, ifname) < 0 {
		C.free(unsafe.Pointer(dp.ctx))
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
	dp.ctx.config.use_af_xdp = cConfig.use_af_xdp

	return dp, nil
}

func signatureFilterValue(filter string) (int, error) {
	if err := config.ValidateSignatureFilter(filter); err != nil {
		return 0, err
	}
	values := map[string]int{
		"all": int(C.SIG_FILTER_ALL), "ito": int(C.SIG_FILTER_ITO),
		"rfc2544": int(C.SIG_FILTER_RFC2544), "y1564": int(C.SIG_FILTER_Y1564),
		"custom": int(C.SIG_FILTER_CUSTOM), "msn": int(C.SIG_FILTER_MSN),
		"probeot": int(C.SIG_FILTER_PROBEOT), "dataot": int(C.SIG_FILTER_DATAOT),
		"latency": int(C.SIG_FILTER_LATENCY),
	}
	return values[filter], nil
}

func reflectionModeValue(mode string) (int, error) {
	values := map[string]int{
		"mac": int(C.REFLECT_MODE_MAC), "mac-ip": int(C.REFLECT_MODE_MAC_IP),
		"all": int(C.REFLECT_MODE_ALL),
	}
	value, ok := values[mode]
	if !ok {
		return 0, fmt.Errorf("invalid reflection mode %q", mode)
	}
	return value, nil
}

// Start begins packet processing.
func (dp *Dataplane) Start() error {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	if dp.running {
		return errors.New("dataplane already running")
	}
	if dp.closed {
		return errors.New("dataplane is closed")
	}
	if dp.ctx == nil {
		return errors.New("dataplane is not initialized")
	}

	ctx := C.uintptr_t(uintptr(unsafe.Pointer(dp.ctx)))
	if C.start_reflector(ctx) < 0 {
		return errors.New("failed to start reflector")
	}

	dp.running = true
	return nil
}

// Stop halts packet processing.
func (dp *Dataplane) Stop() {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	if !dp.running {
		return
	}

	C.stop_reflector(C.uintptr_t(uintptr(unsafe.Pointer(dp.ctx))))
	dp.running = false
}

// Close cleans up dataplane resources.
func (dp *Dataplane) Close() {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	if dp.closed {
		return
	}
	if dp.running {
		C.stop_reflector(C.uintptr_t(uintptr(unsafe.Pointer(dp.ctx))))
		dp.running = false
	}
	if dp.ctx != nil {
		C.cleanup_reflector(C.uintptr_t(uintptr(unsafe.Pointer(dp.ctx))))
		C.free(unsafe.Pointer(dp.ctx))
		dp.ctx = nil
	}
	// Free stored C strings
	dp.closed = true
}

// GetStats returns current statistics.
func (dp *Dataplane) GetStats() Stats {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	if dp.closed || dp.ctx == nil {
		return Stats{}
	}

	var cStats C.reflector_stats_t
	ctx := C.uintptr_t(uintptr(unsafe.Pointer(dp.ctx)))
	stats := C.uintptr_t(uintptr(unsafe.Pointer(&cStats)))
	C.get_reflector_stats(ctx, stats)

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
		LatencyMin:       float64(cStats.latency.min_ns) / nanosecondsPerMicrosecond,
		LatencyAvg:       float64(cStats.latency.avg_ns) / nanosecondsPerMicrosecond,
		LatencyMax:       float64(cStats.latency.max_ns) / nanosecondsPerMicrosecond,
		LatencyCount:     uint64(cStats.latency.count),
	}
}

// IsRunning returns whether the dataplane is active.
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

// Config returns the configuration.
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
	sigFilter, mode, parsedOUI, err := validateConfigUpdate(update)
	if err != nil {
		return err
	}

	dp.mu.Lock()
	defer dp.mu.Unlock()
	if dp.closed {
		return errors.New("dataplane is closed")
	}
	if dp.running {
		return errors.New("reflector configuration cannot change while running")
	}
	if dp.ctx == nil || dp.cfg == nil {
		return errors.New("dataplane is not initialized")
	}

	if filterErr := dp.updatePacketFilter(update, sigFilter); filterErr != nil {
		return filterErr
	}

	dp.applyConfigUpdate(update, mode, parsedOUI)
	return nil
}

func validateConfigUpdate(update *ConfigUpdate) (int, int, [3]byte, error) {
	sigFilter, mode := 0, 0
	if update.SignatureFilter != nil {
		value, err := signatureFilterValue(*update.SignatureFilter)
		if err != nil {
			return 0, 0, [3]byte{}, err
		}
		sigFilter = value
	}
	if update.Mode != nil {
		value, err := reflectionModeValue(*update.Mode)
		if err != nil {
			return 0, 0, [3]byte{}, err
		}
		mode = value
	}
	var parsedOUI [3]byte
	if update.OUI != nil {
		candidate := config.Config{Filtering: config.FilterConfig{OUI: *update.OUI}}
		value, err := candidate.ParseOUI()
		if err != nil {
			return 0, 0, [3]byte{}, fmt.Errorf("invalid OUI format %q: %w", *update.OUI, err)
		}
		parsedOUI = value
	}
	return sigFilter, mode, parsedOUI, nil
}

func (dp *Dataplane) applyConfigUpdate(update *ConfigUpdate, mode int, parsedOUI [3]byte) {
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
}

func (dp *Dataplane) updatePacketFilter(update *ConfigUpdate, sigFilter int) error {
	if update.Port == nil && update.SignatureFilter == nil {
		return nil
	}
	port := dp.cfg.Filtering.Port
	if update.Port != nil {
		port = *update.Port
	}
	if update.SignatureFilter == nil {
		var err error
		sigFilter, err = signatureFilterValue(dp.cfg.SignatureFilter)
		if err != nil {
			return err
		}
	}
	ctx := C.uintptr_t(uintptr(unsafe.Pointer(dp.ctx)))
	if C.update_reflector_filter(ctx, C.uint16_t(port), C.sig_filter_t(sigFilter)) != 0 {
		return errors.New("failed to update reflector packet filter")
	}
	return nil
}

// ResetStats resets the statistics counters.
func (dp *Dataplane) ResetStats() {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	if dp.closed || dp.ctx == nil {
		return
	}
	C.reset_reflector_stats(C.uintptr_t(uintptr(unsafe.Pointer(dp.ctx))))
}
