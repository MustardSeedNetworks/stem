//go:build cgo && linux

// SPDX-License-Identifier: BUSL-1.1

package dataplane

import (
	"encoding/json"
	"math/big"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/reflector/config"
	reflectordp "github.com/MustardSeedNetworks/stem/internal/reflector/dataplane"
)

const (
	vethMaster    = "stmxdp0"
	vethReflector = "stmxdp1"
	masterIP      = "10.251.32.1"
	reflectorIP   = "10.251.32.2"
	reflectorPort = 3842
)

func ip(t *testing.T, args ...string) []byte {
	t.Helper()
	out, err := exec.Command("ip", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("ip %v: %v\n%s", args, err, out)
	}
	return out
}

// newReflectedLink links a test master to a running reflector over a veth
// pair with the given number of queues at each end, and returns the test
// master's context. Native XDP on veth refuses a peer with more transmit
// queues than it has receive queues, so both ends match. reflectorAFXDP
// selects the reflector's platform.
func newReflectedLink(t *testing.T, queues int, reflectorAFXDP bool) *Context {
	t.Helper()
	n := strconv.Itoa(queues)
	ip(t, "link", "add", vethMaster, "numrxqueues", n, "numtxqueues", n,
		"type", "veth", "peer", "name", vethReflector, "numrxqueues", n, "numtxqueues", n)
	t.Cleanup(func() { _ = exec.Command("ip", "link", "del", vethMaster).Run() })
	ip(t, "addr", "add", masterIP+"/24", "dev", vethMaster)
	ip(t, "link", "set", vethMaster, "up")
	ip(t, "link", "set", vethReflector, "up")

	reflector, err := reflectordp.New(&config.Config{
		Interface:       vethReflector,
		SignatureFilter: "all",
		Filtering:       config.FilterConfig{Port: reflectorPort},
		Reflection:      config.ReflectConfig{Mode: "all"},
		Platform:        config.PlatformConfig{UseAFXDP: reflectorAFXDP},
	})
	if err != nil {
		t.Fatalf("reflector: %v", err)
	}
	t.Cleanup(reflector.Close)
	if err = reflector.Start(); err != nil {
		t.Fatalf("start reflector: %v", err)
	}
	t.Cleanup(reflector.Stop)

	// The reflector end carries no address because it reflects at layer 2,
	// so pin the neighbour entry ARP would have made. After the reflector
	// starts: its socket setup flushes the entry.
	var links []struct {
		Address string `json:"address"`
	}
	if err = json.Unmarshal(ip(t, "-j", "link", "show", "dev", vethReflector), &links); err != nil || len(links) != 1 {
		t.Fatalf("read %s MAC: %v (%d links)", vethReflector, err, len(links))
	}
	ip(t, "neigh", "replace", reflectorIP, "lladdr", links[0].Address, "dev", vethMaster, "nud", "permanent")

	master, err := NewContext(vethMaster)
	if err != nil {
		t.Fatalf("test master: %v", err)
	}
	t.Cleanup(master.Close)
	return master
}

// trialReturnsWhatItSent runs one RFC 2544 trial at rate and requires the
// reflected frames back within 1 %.
func trialReturnsWhatItSent(t *testing.T, master *Context, rate float64) {
	t.Helper()
	cfg := PeerConfig(vethMaster, reflectorIP, reflectorPort)
	cfg.InitialRatePct = rate
	cfg.MaxIterations = 1
	cfg.TrialDuration = 2 * time.Second
	if err := master.Configure(&cfg); err != nil {
		t.Fatalf("configure test master: %v", err)
	}
	results, err := master.runThroughputTestInternal(512)
	if err != nil || len(results) != 1 {
		t.Fatalf("trial: %v (%d results)", err, len(results))
	}
	sent, recv := results[0].FramesTested, results[0].FramesReceived
	t.Logf("rate=%v%% sent=%d recv=%d", rate, sent, recv)
	if sent == 0 {
		t.Fatal("trial sent nothing")
	}
	if float64(sent)-float64(recv) > float64(sent)/100 {
		t.Errorf("received %d of %d sent, want within 1%%", recv, sent)
	}
}

// xdpAttached reports whether an XDP program is attached to the interface.
func xdpAttached(t *testing.T, iface string) bool {
	t.Helper()
	var links []struct {
		XDP json.RawMessage `json:"xdp"`
	}
	if err := json.Unmarshal(ip(t, "-j", "-d", "link", "show", "dev", iface), &links); err != nil || len(links) != 1 {
		t.Fatalf("read %s XDP state: %v (%d links)", iface, err, len(links))
	}
	return links[0].XDP != nil
}

func requireRoot(t *testing.T) {
	t.Helper()
	if os.Geteuid() != 0 {
		t.Skip("needs root to create a veth pair and an AF_XDP socket")
	}
}

// The AF_XDP test master must receive what a reflecting peer sends back
// (stem#1328). It created its socket without loading an XDP program, so
// nothing redirected received frames into it: every trial sent and got 0
// back. Both rates run on one socket: reconfiguring for the second must not
// rebind the queue, which the kernel frees asynchronously.
func TestAFXDPTrialReceivesFromReflectingPeer(t *testing.T) {
	requireRoot(t)
	master := newReflectedLink(t, 1, false)

	for _, rate := range []float64{1, 50} {
		t.Run(strconv.FormatFloat(rate, 'f', -1, 64)+"pct", func(t *testing.T) {
			trialReturnsWhatItSent(t, master, rate)
		})
	}

	// The trials prove nothing about AF_XDP if the test master fell back to
	// AF_PACKET; an attached program is what the AF_XDP path adds.
	if !xdpAttached(t, vethMaster) {
		t.Error("no XDP program on the test master's interface: the trials did not run over AF_XDP")
	}
	master.Close()
	if xdpAttached(t, vethMaster) {
		t.Error("XDP program still attached after the test master closed")
	}
}

// reflectInto makes an AF_PACKET reflector transmit every frame on one queue,
// which a veth delivers to the same receive queue at the test master. A single
// flow's RSS hash picks a queue that changes with the kernel's boot-time seed,
// so without this a trial could land on queue 0 and prove nothing.
func reflectInto(t *testing.T, queue, queues int) {
	t.Helper()
	all := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(runtime.NumCPU())), big.NewInt(1))
	for q := range queues {
		mask := "0"
		if q == queue {
			mask = all.Text(16)
		}
		path := "/sys/class/net/" + vethReflector + "/queues/tx-" + strconv.Itoa(q) + "/xps_cpus"
		if err := os.WriteFile(path, []byte(mask), 0); err != nil {
			t.Fatalf("steer reflector transmit: %v", err)
		}
	}
}

// One AF_XDP socket reads one receive queue, so on a multi-queue interface the
// test master needs one per queue, or it counts frames on the others as loss
// (stem#1533). Before, it refused AF_XDP there and measured over AF_PACKET.
func TestAFXDPReceivesOnEveryQueue(t *testing.T) {
	requireRoot(t)
	for _, queues := range []int{2, 4} {
		t.Run(strconv.Itoa(queues)+"queues", func(t *testing.T) {
			// An AF_PACKET reflector, so XPS picks its transmit queue.
			master := newReflectedLink(t, queues, false)
			reflectInto(t, queues-1, queues)
			for _, rate := range []float64{1, 50} {
				trialReturnsWhatItSent(t, master, rate)
			}
			if !xdpAttached(t, vethMaster) {
				t.Error("no XDP program on the test master's interface: the trials did not run over AF_XDP")
			}
			master.Close()
			if xdpAttached(t, vethMaster) {
				t.Error("XDP program still attached after the test master closed")
			}
		})
	}
}

// veth has native XDP but no zero-copy, like every copy-mode driver. The
// reflector demanded XDP_ZEROCOPY, so its bind failed EOPNOTSUPP and it
// always fell back to AF_PACKET there (stem#1532).
func TestReflectorAFXDPRunsInCopyMode(t *testing.T) {
	requireRoot(t)
	master := newReflectedLink(t, 1, true)

	if !xdpAttached(t, vethReflector) {
		t.Error("no XDP program on the reflector's interface: it fell back to AF_PACKET")
	}
	trialReturnsWhatItSent(t, master, 1)
}
