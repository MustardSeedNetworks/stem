//go:build cgo && linux

// SPDX-License-Identifier: BUSL-1.1

package dataplane

import (
	"errors"
	"testing"
	"time"
)

// y1731Run runs one Y.1731 test type and reports its sent and received
// frame counts.
type y1731Run func(c *Context, cfg *Y1731Config) (sent, received uint64, err error)

// y1731Runs returns every Y.1731 test type by name.
func y1731Runs() map[string]y1731Run {
	return map[string]y1731Run{
		"delay": func(c *Context, cfg *Y1731Config) (uint64, uint64, error) {
			r, err := c.RunY1731DelayTest(cfg)
			if err != nil {
				return 0, 0, err
			}
			if lost := r.FramesSent - r.FramesReceived; r.FramesLost != lost {
				return 0, 0, errors.New("delay test FramesLost is not sent - received")
			}
			return uint64(r.FramesSent), uint64(r.FramesReceived), nil
		},
		"loss": func(c *Context, cfg *Y1731Config) (uint64, uint64, error) {
			r, err := c.RunY1731LossTest(cfg)
			if err != nil {
				return 0, 0, err
			}
			return r.FramesTx, r.FramesRx, availabilityMatchesLoss(r)
		},
		"slm": func(c *Context, cfg *Y1731Config) (uint64, uint64, error) {
			r, err := c.RunY1731SyntheticLossTest(cfg)
			if err != nil {
				return 0, 0, err
			}
			return r.FramesTx, r.FramesRx, availabilityMatchesLoss(r)
		},
		"loopback": func(c *Context, cfg *Y1731Config) (uint64, uint64, error) {
			r, err := c.RunY1731LoopbackTest(cfg)
			if err != nil {
				return 0, 0, err
			}
			return r.LBMSent, r.LBRReceived, nil
		},
	}
}

func availabilityMatchesLoss(r *Y1731LossResult) error {
	if r.AvailabilityPct != 100*(1-r.NearEndLossRatio) {
		return errors.New("availability is not 100 % less the loss ratio")
	}
	return nil
}

func y1731TestConfig() *Y1731Config {
	return &Y1731Config{
		MEPID: 1, MEGLevel: 4, MEGID: "MSN-MEG-01", Priority: 6,
		DurationSec: 1, IntervalMs: 100, Count: 10, FrameSize: 128,
	}
}

// Every Y.1731 test must measure a reflecting peer and fail against a silent
// one (stem#1731): with no frame back each one passed, reporting 100 % loss,
// 0 % availability or a zero delay as a measurement.
func TestY1731ReflectedByPeer(t *testing.T) {
	requireRoot(t)
	master := newReflectedLink(t, 1, false)
	cfg := PeerConfig(vethMaster, reflectorIP, reflectorPort)
	if err := master.Configure(&cfg); err != nil {
		t.Fatalf("configure test master: %v", err)
	}

	for name, run := range y1731Runs() {
		sent, received, err := run(master, y1731TestConfig())
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		t.Logf("%s: sent=%d received=%d", name, sent, received)
		if received < sent*99/100 || received > sent {
			t.Errorf("%s: received %d of %d sent, want within 1%%", name, received, sent)
		}
	}
}

// The count-based tests send exactly Count frames, one every IntervalMs
// (stem#1736): they ran a padded timed trial at a floored line-rate share,
// so 10 frames at 100 ms sent 504 delay frames and 1,261 LBMs.
func TestY1731SendsConfiguredCountAtInterval(t *testing.T) {
	requireRoot(t)
	master := newReflectedLink(t, 1, false)
	cfg := PeerConfig(vethMaster, reflectorIP, reflectorPort)
	if err := master.Configure(&cfg); err != nil {
		t.Fatalf("configure test master: %v", err)
	}

	runs := y1731Runs()
	for _, name := range []string{"delay", "slm", "loopback"} {
		tc := y1731TestConfig()
		start := time.Now()
		sent, _, err := runs[name](master, tc)
		elapsed := time.Since(start)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if sent != uint64(tc.Count) {
			t.Errorf("%s: sent %d frames, want the configured %d", name, sent, tc.Count)
		}
		// Count frames, each followed by one interval for its reply.
		want := time.Duration(tc.Count) * time.Duration(tc.IntervalMs) * time.Millisecond
		if elapsed < want || elapsed > want+time.Second {
			t.Errorf("%s: took %v, want %v at %d ms per frame", name, elapsed, want, tc.IntervalMs)
		}
	}
}

func TestY1731FailsWhenNothingReturns(t *testing.T) {
	requireRoot(t)
	addVethPair(t, 1)
	master := newLinkMaster(t)
	cfg := PeerConfig(vethMaster, reflectorIP, reflectorPort)
	if err := master.Configure(&cfg); err != nil {
		t.Fatalf("configure test master: %v", err)
	}

	for name, run := range y1731Runs() {
		if _, _, err := run(master, y1731TestConfig()); !errors.Is(err, ErrNoFramesReturned) {
			t.Errorf("%s to a silent peer = %v, want %v", name, err, ErrNoFramesReturned)
		}
	}
}
