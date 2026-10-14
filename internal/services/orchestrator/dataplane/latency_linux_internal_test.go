//go:build cgo && linux

// SPDX-License-Identifier: BUSL-1.1

package dataplane

import (
	"os/exec"
	"testing"
	"time"
)

// Latency must be timed at the path's throughput, not at a percentage of the
// interface's line rate (stem#1680). STM-11 ran it over a 74 Mbps path behind
// a NIC taken to be 10 Gbps: every level overran the path and timed ~236 ms
// of queueing. Here the reflector's egress is shaped to 50 Mbit with a 100 ms
// queue, so a line-rate load reads as that queue.
func TestLatencyRunsAtMeasuredThroughput(t *testing.T) {
	requireRoot(t)
	master := newReflectedLink(t, 1, false)
	shapeIngress(t, vethReflector, "50mbit", "100ms")

	cfg := PeerConfig(vethMaster, reflectorIP, reflectorPort)
	cfg.InitialRatePct = 100
	cfg.ResolutionPct = 0.1
	cfg.MaxIterations = 20
	cfg.TrialDuration = time.Second
	if err := master.Configure(&cfg); err != nil {
		t.Fatalf("configure test master: %v", err)
	}
	master.SetFrameSize(512)

	results, err := master.RunLatencyTest([]float64{100})
	if err != nil {
		t.Fatalf("latency: %v", err)
	}
	got := results[0].Latency
	t.Logf("avg=%.1f us max=%.1f us over %d frames", got.AvgNs/1000, got.MaxNs/1000, got.Count)
	if limit := 20 * time.Millisecond; time.Duration(got.AvgNs) > limit {
		t.Errorf("average latency %v at the throughput rate, want under %v: the load overran the path",
			time.Duration(got.AvgNs), limit)
	}
}

// Frame delay must be the path's delay (stem#1681). Over this veth the path
// adds tens of microseconds, yet Y.1564 reported an average of 0.7-0.8 ms and
// a 10 ms maximum: the reflector's TPACKET_V3 ring held each frame until its
// block's 1 ms retire timer fired, the test master read replies only after
// sleeping to its next send slot, and replies to the last frames were read
// after a 10 ms sleep.
func TestY1564FrameDelayIsThePathDelay(t *testing.T) {
	requireRoot(t)
	master := newReflectedLink(t, 1, false)
	cfg := PeerConfig(vethMaster, reflectorIP, reflectorPort)
	if err := master.Configure(&cfg); err != nil {
		t.Fatalf("configure test master: %v", err)
	}

	res, err := master.RunY1564PerfTest(&Y1564Service{
		ServiceID: 1, ServiceName: "delay", FrameSize: 512, Enabled: true,
		SLA: Y1564SLA{CIRMbps: 25, FDThresholdMs: 10, FDVThresholdMs: 5, FLRThresholdPct: 1},
	}, 5)
	if err != nil {
		t.Fatalf("Y.1564: %v", err)
	}
	t.Logf("FD avg/min/max = %.3f/%.3f/%.3f ms, FDV %.3f ms over %d frames",
		res.FDAvgMs, res.FDMinMs, res.FDMaxMs, res.FDVMs, res.FramesRx)
	if res.FDAvgMs > 0.2 {
		t.Errorf("average frame delay %.3f ms over a veth, want under 0.2 ms", res.FDAvgMs)
	}
	if res.FDMaxMs > 5 {
		t.Errorf("maximum frame delay %.3f ms over a veth, want under 5 ms", res.FDMaxMs)
	}
}

// shapeIngress rate-limits what dev receives through a token bucket whose
// queue holds latency worth of frames. Both ends send with
// PACKET_QDISC_BYPASS, so an egress qdisc never sees the test traffic; the
// shaper sits on an ifb device that dev's ingress is redirected through.
func shapeIngress(t *testing.T, dev, rate, latency string) {
	t.Helper()
	const ifb = "stmifb0"
	run(t, "modprobe", "ifb", "numifbs=0")
	ip(t, "link", "add", ifb, "type", "ifb")
	t.Cleanup(func() { _ = exec.Command("ip", "link", "del", ifb).Run() })
	ip(t, "link", "set", ifb, "up")
	run(t, "tc", "qdisc", "add", "dev", ifb, "root", "tbf", "rate", rate, "burst", "32kbit", "latency", latency)
	run(t, "tc", "qdisc", "add", "dev", dev, "handle", "ffff:", "ingress")
	run(t, "tc", "filter", "add", "dev", dev, "parent", "ffff:", "matchall",
		"action", "mirred", "egress", "redirect", "dev", ifb)
}

func run(t *testing.T, name string, args ...string) {
	t.Helper()
	if out, err := exec.Command(name, args...).CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}
