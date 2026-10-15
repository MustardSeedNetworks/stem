//go:build cgo && linux

// SPDX-License-Identifier: BUSL-1.1

package dataplane

import (
	"encoding/json"
	"math/big"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"

	"github.com/MustardSeedNetworks/stem/internal/reflector/config"
	reflectordp "github.com/MustardSeedNetworks/stem/internal/reflector/dataplane"
)

const (
	vethMaster    = "stmxdp0"
	vethReflector = "stmxdp1"
	// Fixed, locally administered MACs. A kernel-random MAC is one udev's
	// MACAddressPolicy=persistent rewrites some time after the link appears,
	// and a MAC change on the test master flushes its neighbour table,
	// permanent entries included (stem#1692).
	masterMAC     = "02:57:45:4d:00:01"
	reflectorMAC  = "02:57:45:4d:00:02"
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
// master's context and the reflector. Native XDP on veth refuses a peer with
// more transmit queues than it has receive queues, so both ends match.
// reflectorAFXDP selects the reflector's platform.
func newReflectedLink(t *testing.T, queues int, reflectorAFXDP bool) (*Context, *reflectordp.Dataplane) {
	t.Helper()
	addVethPair(t, queues)
	reflector := startReflector(t, reflectorAFXDP)
	return newLinkMaster(t), reflector
}

// startReflector runs a reflector on the veth pair's reflector end.
func startReflector(t *testing.T, afxdp bool) *reflectordp.Dataplane {
	t.Helper()
	reflector, err := reflectordp.New(&config.Config{
		Interface:       vethReflector,
		SignatureFilter: "all",
		Filtering:       config.FilterConfig{Port: reflectorPort},
		Reflection:      config.ReflectConfig{Mode: "all"},
		Platform:        config.PlatformConfig{UseAFXDP: afxdp},
	})
	if err != nil {
		t.Fatalf("reflector: %v", err)
	}
	t.Cleanup(reflector.Close)
	if err = reflector.Start(); err != nil {
		t.Fatalf("start reflector: %v", err)
	}
	t.Cleanup(reflector.Stop)
	return reflector
}

// addVethPair creates the test master's veth pair with the given number of
// queues at each end and brings both ends up.
func addVethPair(t *testing.T, queues int) {
	t.Helper()
	n := strconv.Itoa(queues)
	ip(t, "link", "add", vethMaster, "address", masterMAC, "numrxqueues", n, "numtxqueues", n,
		"type", "veth", "peer", "name", vethReflector, "address", reflectorMAC,
		"numrxqueues", n, "numtxqueues", n)
	t.Cleanup(func() { _ = exec.Command("ip", "link", "del", vethMaster).Run() })
	// Without IPv6 the link carries only the test's own frames, so a reflector
	// count can be compared with the test master's exactly.
	for _, iface := range []string{vethMaster, vethReflector} {
		if err := os.WriteFile("/proc/sys/net/ipv6/conf/"+iface+"/disable_ipv6", []byte("1"), 0); err != nil {
			t.Fatalf("disable IPv6 on %s: %v", iface, err)
		}
	}
	ip(t, "addr", "add", masterIP+"/24", "dev", vethMaster)
	ip(t, "link", "set", vethMaster, "up")
	ip(t, "link", "set", vethReflector, "up")
}

// newLinkMaster returns a test master context on the veth pair. The reflector
// end carries no address because a reflector answers at layer 2, so this pins
// the neighbour entry ARP would have made.
func newLinkMaster(t *testing.T) *Context {
	t.Helper()
	ip(t, "neigh", "replace", reflectorIP, "lladdr", reflectorMAC, "dev", vethMaster, "nud", "permanent")

	master, err := NewContext(vethMaster)
	if err != nil {
		t.Fatalf("test master: %v", err)
	}
	t.Cleanup(master.Close)
	return master
}

// runTrial runs one RFC 2544 trial at rate and returns the frames it sent and
// received back.
func runTrial(t *testing.T, master *Context, rate float64) (uint64, uint64) {
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
	return sent, recv
}

// trialsAccountForEveryFrame runs one RFC 2544 trial at each rate, stops the
// reflector and returns its statistics. Every frame the test master sent must
// have been recognised by the reflector or dropped by its kernel socket, and
// every frame the reflector sent back must have arrived. A loss tolerance
// fails on a busy host, where the reflector's receive ring overflows; the
// reflector's drop count explains that loss exactly (stem#1725).
func trialsAccountForEveryFrame(
	t *testing.T, master *Context, reflector *reflectordp.Dataplane, rates ...float64,
) reflectordp.Stats {
	t.Helper()
	var sent, recv uint64
	for _, rate := range rates {
		trialSent, trialRecv := runTrial(t, master, rate)
		sent += trialSent
		recv += trialRecv
	}
	reflector.Stop()
	st := reflector.GetStats()
	t.Logf("reflector recognised=%d dropped=%d reflected=%d", st.SigRFC2544, st.PacketsDropped, st.PacketsReflected)
	if recv == 0 {
		t.Error("the test master received nothing back")
	}
	if st.SigRFC2544+st.PacketsDropped != sent {
		t.Errorf("test master sent %d frames, reflector recognised %d and dropped %d",
			sent, st.SigRFC2544, st.PacketsDropped)
	}
	if st.PacketsReflected != recv {
		t.Errorf("reflector reflected %d frames, the test master received %d", st.PacketsReflected, recv)
	}
	return st
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
	master, reflector := newReflectedLink(t, 1, false)
	trialsAccountForEveryFrame(t, master, reflector, 1, 50)

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
	steerTransmit(t, vethReflector, queue, queues)
}

// steerTransmit makes every CPU transmit on one of iface's queues.
func steerTransmit(t *testing.T, iface string, queue, queues int) {
	t.Helper()
	all := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(runtime.NumCPU())), big.NewInt(1))
	for q := range queues {
		mask := "0"
		if q == queue {
			mask = all.Text(16)
		}
		path := "/sys/class/net/" + iface + "/queues/tx-" + strconv.Itoa(q) + "/xps_cpus"
		if err := os.WriteFile(path, []byte(mask), 0); err != nil {
			t.Fatalf("steer %s transmit: %v", iface, err)
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
			master, reflector := newReflectedLink(t, queues, false)
			reflectInto(t, queues-1, queues)
			trialsAccountForEveryFrame(t, master, reflector, 1, 50)
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
	master, reflector := newReflectedLink(t, 1, true)

	if !xdpAttached(t, vethReflector) {
		t.Error("no XDP program on the reflector's interface: it fell back to AF_PACKET")
	}
	// A received frame the reflector does not recognise was overwritten in a
	// UMEM frame handed to the kernel twice, and a reflected frame that never
	// arrives was left in its TX ring (stem#1711).
	st := trialsAccountForEveryFrame(t, master, reflector, 1, 50)
	// Each trial primes the test master's neighbour entry with one UDP
	// datagram, the only frame on the link that is not a test frame.
	const trials = 2
	if unrecognised := st.PacketsReceived - st.SigRFC2544; unrecognised != trials {
		t.Errorf("reflector received %d frames that are not test frames, want %d", unrecognised, trials)
	}
}

// xdpRedirected returns how many frames iface's XDP program has redirected
// into an AF_XDP socket on one receive queue.
func xdpRedirected(t *testing.T, iface string, queue int) uint64 {
	t.Helper()
	out, err := exec.Command("ethtool", "-S", iface).CombinedOutput()
	if err != nil {
		t.Fatalf("ethtool -S %s: %v\n%s", iface, err, out)
	}
	name := "rx_queue_" + strconv.Itoa(queue) + "_xdp_redirect:"
	for line := range strings.Lines(string(out)) {
		if value, ok := strings.CutPrefix(strings.TrimSpace(line), name); ok {
			n, parseErr := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
			if parseErr != nil {
				t.Fatalf("parse %s %q: %v", name, value, parseErr)
			}
			return n
		}
	}
	t.Fatalf("ethtool -S %s has no %s", iface, name)
	return 0
}

// The reflector's first worker attaches the XDP program and every other
// worker's socket joins its map under its own queue. The program passes a
// queue's frames to the stack when its socket is missing from the map, so
// every queue must redirect what arrives on it (stem#1747). Counted per queue
// at the veth, which the reflector's own counters are not.
func TestReflectorAFXDPReceivesOnEveryQueue(t *testing.T) {
	requireRoot(t)
	if _, err := exec.LookPath("ethtool"); err != nil {
		t.Skip("needs ethtool to read the veth's per-queue XDP counters")
	}
	const queues = 4
	addVethPair(t, queues)
	startReflector(t, true)
	ip(t, "neigh", "replace", reflectorIP, "lladdr", reflectorMAC, "dev", vethMaster, "nud", "permanent")
	if !xdpAttached(t, vethReflector) {
		t.Fatal("no XDP program on the reflector's interface: it fell back to AF_PACKET")
	}

	payload := make([]byte, 64)
	copy(payload, "RFC254")
	const frames = 100
	for q := range queues {
		// A new socket per queue: a connected socket keeps the queue its
		// first send picked.
		steerTransmit(t, vethMaster, q, queues)
		conn, err := net.Dial("udp4", net.JoinHostPort(reflectorIP, strconv.Itoa(reflectorPort)))
		if err != nil {
			t.Fatalf("dial reflector: %v", err)
		}
		before := xdpRedirected(t, vethReflector, q)
		for range frames {
			if _, err = conn.Write(payload); err != nil {
				t.Fatalf("send on queue %d: %v", q, err)
			}
		}
		_ = conn.Close()
		if got := xdpRedirected(t, vethReflector, q) - before; got < frames {
			t.Errorf("queue %d redirected %d of %d frames to the reflector", q, got, frames)
		}
	}
}

// The reflector's XDP program redirected every frame on its interface to its
// socket, and the reflector drops what is not a test frame, so the host's own
// ARP, ICMP, SSH and DHCP never reached the stack. A reflector on a host's only
// NIC cut the host off the network (stem#1767). The test master's end sits in
// its own network namespace so its frames cross the veth instead of being
// delivered locally. Named outside the packaged-capabilities pattern because
// creating a namespace needs CAP_SYS_ADMIN.
func TestHostTrafficPassesAFXDPReflector(t *testing.T) {
	requireRoot(t)
	const netns = "stmxdpns"
	ip(t, "netns", "add", netns)
	t.Cleanup(func() { _ = exec.Command("ip", "netns", "del", netns).Run() })
	addVethPair(t, 1)
	ip(t, "link", "set", vethMaster, "netns", netns)
	inNetns := func(args ...string) *exec.Cmd {
		return exec.Command("ip", append([]string{"netns", "exec", netns}, args...)...)
	}
	for _, args := range [][]string{
		{"ip", "addr", "add", masterIP + "/24", "dev", vethMaster},
		{"ip", "link", "set", vethMaster, "up"},
	} {
		if out, err := inNetns(args...).CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
	ip(t, "addr", "add", reflectorIP+"/24", "dev", vethReflector)

	other, err := net.ListenPacket("udp4", net.JoinHostPort(reflectorIP, "0"))
	if err != nil {
		t.Fatalf("listen on the reflector host: %v", err)
	}
	defer other.Close()
	otherPort := other.LocalAddr().(*net.UDPAddr).Port

	reflector := startReflector(t, true)
	if !xdpAttached(t, vethReflector) {
		t.Fatal("no XDP program on the reflector's interface: it fell back to AF_PACKET")
	}

	// ARP and ICMP: the host must stay reachable.
	if out, pingErr := inNetns("ping", "-c", "3", "-W", "1", reflectorIP).CombinedOutput(); pingErr != nil {
		t.Errorf("the reflector host stopped answering ping: %v\n%s", pingErr, out)
	}

	// UDP to another port, as DNS and DHCP are, reaches its socket; a test
	// frame to the reflector's port reaches the reflector.
	send := func(port int) {
		t.Helper()
		script := `printf 'RFC2544%057d' 0 > /dev/udp/` + reflectorIP + "/" + strconv.Itoa(port)
		if out, sendErr := inNetns("bash", "-c", script).CombinedOutput(); sendErr != nil {
			t.Fatalf("send to port %d: %v\n%s", port, sendErr, out)
		}
	}
	send(otherPort)
	if err = other.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	if _, _, err = other.ReadFrom(make([]byte, 128)); err != nil {
		t.Errorf("UDP to port %d never reached the host: %v", otherPort, err)
	}
	send(reflectorPort)
	reflector.Stop()
	if st := reflector.GetStats(); st.SigRFC2544 != 1 {
		t.Errorf("reflector recognised %d test frames, want 1", st.SigRFC2544)
	}
}

// The packaged service runs with CAP_BPF, CAP_NET_ADMIN and CAP_NET_RAW and
// without CAP_SYS_ADMIN, which libxdp's own loader needed, so every packaged
// run fell back to AF_PACKET (stem#1747). Run the AF_XDP tests again with
// exactly that set. systemd raises the memlock limit for the service
// (LimitMEMLOCK), which the reduced set cannot do itself.
func TestPackagedCapabilitiesRunAFXDP(t *testing.T) {
	requireRoot(t)
	setpriv, err := exec.LookPath("setpriv")
	if err != nil {
		t.Skip("needs setpriv to run with the packaged capabilities")
	}
	unlimited := unix.Rlimit{Cur: unix.RLIM_INFINITY, Max: unix.RLIM_INFINITY}
	if err = unix.Setrlimit(unix.RLIMIT_MEMLOCK, &unlimited); err != nil {
		t.Fatalf("raise memlock limit: %v", err)
	}
	out, err := exec.Command(setpriv, "--bounding-set", "-all,+bpf,+net_admin,+net_raw",
		"--inh-caps", "-all", "--", os.Args[0],
		"-test.run", "^Test(AFXDP|ReflectorAFXDP)", "-test.count=1", "-test.v").CombinedOutput()
	if err != nil {
		t.Fatalf("AF_XDP tests with the packaged capabilities: %v\n%s", err, out)
	}
}
