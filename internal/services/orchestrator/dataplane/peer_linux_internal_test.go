//go:build cgo && linux

package dataplane

import (
	"encoding/binary"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
)

func TestRouteGatewayUsesKernelGateway(t *testing.T) {
	data := make([]byte, unix.SizeofRtMsg)
	index := make([]byte, netlinkAlignment)
	binary.NativeEndian.PutUint32(index, 7)
	data, err := appendRouteAttribute(data, unix.RTA_OIF, index)
	if err != nil {
		t.Fatalf("append output-interface attribute: %v", err)
	}
	data, err = appendRouteAttribute(data, unix.RTA_GATEWAY, net.ParseIP("10.44.30.1").To4())
	if err != nil {
		t.Fatalf("append gateway attribute: %v", err)
	}
	message := syscall.NetlinkMessage{Header: syscall.NlMsghdr{Type: unix.RTM_NEWROUTE}, Data: data}

	nextHop, err := routeGateway(message, 7, net.ParseIP("10.44.40.23"))
	if err != nil {
		t.Fatalf("routeGateway: %v", err)
	}
	if got, want := nextHop.String(), "10.44.30.1"; got != want {
		t.Fatalf("next hop = %s, want %s", got, want)
	}
}

func TestRouteNextHopQueriesKernelRoute(t *testing.T) {
	loopback := net.ParseIP("127.0.0.1")
	nextHop, err := routeNextHop("lo", loopback, loopback)
	if err != nil {
		t.Fatalf("routeNextHop: %v", err)
	}
	if got, want := nextHop.String(), "127.0.0.1"; got != want {
		t.Fatalf("next hop = %s, want %s", got, want)
	}
}

func writePeerFixture(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fixture")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

func TestLookupARPRequiresReachableNeighborOnSelectedInterface(t *testing.T) {
	path := writePeerFixture(t, "IP address HW type Flags HW address Mask Device\n"+
		"10.44.0.1 0x1 0x6 aa:bb:cc:dd:ee:ff * ens18\n"+
		"10.44.0.1 0x1 0x2 11:22:33:44:55:66 * eth9\n")

	mac, err := lookupARP(path, "ens18", net.ParseIP("10.44.0.1"))
	if err != nil {
		t.Fatalf("lookupARP: %v", err)
	}
	if got, want := mac.String(), "aa:bb:cc:dd:ee:ff"; got != want {
		t.Fatalf("MAC = %s, want %s", got, want)
	}
}
