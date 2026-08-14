//go:build cgo && linux

package dataplane

const (
	defaultCustomFrameSize = 1518
	defaultCustomRatePct   = 10.0
	defaultCustomDuration  = 10
	defaultCustomWarmup    = 1
)

func customStreamParameters(cfg *TrafficGenConfig) (uint32, float64, uint32, uint32, uint32) {
	frameSize := uint32(defaultCustomFrameSize)
	ratePct := defaultCustomRatePct
	durationSec := uint32(defaultCustomDuration)
	warmupSec := uint32(defaultCustomWarmup)
	streamID := uint32(0)
	if cfg == nil {
		return frameSize, ratePct, durationSec, warmupSec, streamID
	}
	if cfg.FrameSize > 0 {
		frameSize = cfg.FrameSize
	}
	if cfg.RatePct > 0 {
		ratePct = cfg.RatePct
	}
	if cfg.DurationSec > 0 {
		durationSec = cfg.DurationSec
	}
	if cfg.WarmupSec > 0 {
		warmupSec = cfg.WarmupSec
	}
	if cfg.StreamID > 0 {
		streamID = cfg.StreamID
	}
	return frameSize, ratePct, durationSec, warmupSec, streamID
}
