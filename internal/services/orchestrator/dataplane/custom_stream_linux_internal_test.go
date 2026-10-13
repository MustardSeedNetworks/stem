//go:build cgo && linux

// SPDX-License-Identifier: BUSL-1.1

package dataplane

import (
	"errors"
	"testing"
)

// A traffic-generator stream must come back from a Stem reflector (stem#1531).
// Its frames carried a signature the reflector did not know, so it reflected
// none of them and every run reported 100 % loss.
func TestCustomStreamReflectedByPeer(t *testing.T) {
	requireRoot(t)
	master := newReflectedLink(t, 1, false)

	cfg := PeerConfig(vethMaster, reflectorIP, reflectorPort)
	if err := master.Configure(&cfg); err != nil {
		t.Fatalf("configure test master: %v", err)
	}
	res, err := master.RunCustomStreamTest(&TrafficGenConfig{FrameSize: 512, RatePct: 1, DurationSec: 2})
	if err != nil {
		t.Fatalf("custom stream: %v", err)
	}
	t.Logf("sent=%d recv=%d loss=%v%%", res.PacketsSent, res.PacketsRecv, res.LossPct)
	if res.PacketsSent == 0 {
		t.Fatal("custom stream sent nothing")
	}
	if float64(res.PacketsSent)-float64(res.PacketsRecv) > float64(res.PacketsSent)/100 {
		t.Errorf("received %d of %d sent, want within 1%%", res.PacketsRecv, res.PacketsSent)
	}
}

// A stream nothing reflects must fail, not pass with 100 % loss reported as a
// measurement (stem#1563).
func TestCustomStreamFailsWhenNothingReturns(t *testing.T) {
	requireRoot(t)
	addVethPair(t, 1)
	master := newLinkMaster(t)

	cfg := PeerConfig(vethMaster, reflectorIP, reflectorPort)
	if err := master.Configure(&cfg); err != nil {
		t.Fatalf("configure test master: %v", err)
	}
	res, err := master.RunCustomStreamTest(&TrafficGenConfig{FrameSize: 512, RatePct: 1, DurationSec: 1})
	if !errors.Is(err, ErrNoFramesReturned) {
		t.Fatalf("custom stream to a silent peer = %+v, %v; want %v", res, err, ErrNoFramesReturned)
	}
}
