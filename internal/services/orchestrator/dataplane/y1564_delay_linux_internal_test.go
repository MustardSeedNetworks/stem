//go:build cgo && linux

// SPDX-License-Identifier: BUSL-1.1

package dataplane

import "testing"

// FD and FDV need a frame back to be measured (stem#1732). Against a peer that
// answered nothing, a step sent 1170 frames, got none back and still reported
// FDPass and FDVPass with every delay at 0.
func TestY1564DelayNeedsAFrameBack(t *testing.T) {
	requireRoot(t)
	t.Run("reflected", func(t *testing.T) { y1564DelayJudged(t, true) })
	t.Run("unanswered", func(t *testing.T) { y1564DelayJudged(t, false) })
}

// y1564DelayJudged runs a Y.1564 performance test over a veth pair whose far
// end reflects or stays silent, and requires the delay criteria to be measured
// and passed exactly when frames came back.
func y1564DelayJudged(t *testing.T, reflected bool) {
	t.Helper()
	var master *Context
	if reflected {
		master, _ = newReflectedLink(t, 1, false)
	} else {
		addVethPair(t, 1)
		master = newLinkMaster(t)
	}
	cfg := PeerConfig(vethMaster, reflectorIP, reflectorPort)
	if err := master.Configure(&cfg); err != nil {
		t.Fatalf("configure test master: %v", err)
	}

	res, err := master.RunY1564PerfTest(&Y1564Service{
		ServiceID: 1, ServiceName: "delay", FrameSize: 512, Enabled: true,
		SLA: Y1564SLA{CIRMbps: 10, FDThresholdMs: 10, FDVThresholdMs: 5, FLRThresholdPct: 1},
	}, 1)
	if err != nil {
		t.Fatalf("Y.1564: %v", err)
	}
	t.Logf("tx %d, rx %d, delay measured %t, FD pass %t, FDV pass %t",
		res.FramesTx, res.FramesRx, res.DelayMeasured, res.FDPass, res.FDVPass)
	if res.FramesTx == 0 {
		t.Fatal("sent nothing")
	}
	if got := res.FramesRx > 0; got != reflected {
		t.Fatalf("frames back = %t, want %t", got, reflected)
	}
	if res.DelayMeasured != reflected || res.FDPass != reflected ||
		res.FDVPass != reflected || res.ServicePass != reflected {
		t.Errorf("delay measured %t, FD pass %t, FDV pass %t, service pass %t; want all %t",
			res.DelayMeasured, res.FDPass, res.FDVPass, res.ServicePass, reflected)
	}
}
