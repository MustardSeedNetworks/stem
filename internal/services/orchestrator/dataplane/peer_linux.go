//go:build cgo && linux

package dataplane

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

const (
	defaultSourcePort     = 12345
	neighborPollInterval  = 20 * time.Millisecond
	peerResolutionTimeout = time.Second
)

type resolvedPeer struct {
	localIP    [4]byte
	remoteIP   [4]byte
	remoteMAC  [6]byte
	sourcePort uint16
	remotePort uint16
}

func resolvePeer(ifaceName, host string, port uint16) (resolvedPeer, error) {
	if host == "" || port == 0 {
		return resolvedPeer{}, errors.New("peer host and port are required")
	}
	remoteIP, err := resolveIPv4(host)
	if err != nil {
		return resolvedPeer{}, err
	}
	localIP, err := sourceIPv4(ifaceName, remoteIP, port)
	if err != nil {
		return resolvedPeer{}, err
	}
	nextHop, err := routeNextHop(ifaceName, localIP, remoteIP)
	if err != nil {
		return resolvedPeer{}, err
	}
	remoteMAC, err := resolveNeighbor("/proc/net/arp", ifaceName, localIP, remoteIP, nextHop, port)
	if err != nil {
		return resolvedPeer{}, err
	}
	return resolvedPeer{
		localIP:    [4]byte(localIP.To4()),
		remoteIP:   [4]byte(remoteIP.To4()),
		remoteMAC:  [6]byte(remoteMAC),
		sourcePort: defaultSourcePort,
		remotePort: port,
	}, nil
}

func resolveIPv4(host string) (net.IP, error) {
	ctx, cancel := context.WithTimeout(context.Background(), peerResolutionTimeout)
	defer cancel()
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("resolve peer %q: %w", host, err)
	}
	for _, address := range addresses {
		if ipv4 := address.IP.To4(); ipv4 != nil {
			return ipv4, nil
		}
	}
	return nil, fmt.Errorf("peer %q has no IPv4 address", host)
}

func sourceIPv4(ifaceName string, peer net.IP, port uint16) (net.IP, error) {
	dialer := net.Dialer{Control: bindToDevice(ifaceName)}
	conn, err := dialer.Dial("udp4", net.JoinHostPort(peer.String(), strconv.Itoa(int(port))))
	if err != nil {
		return nil, fmt.Errorf("select source address for %s on %q: %w", peer, ifaceName, err)
	}
	defer conn.Close()
	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || localAddr.IP.To4() == nil {
		return nil, fmt.Errorf("interface %q has no routed IPv4 address for peer %s", ifaceName, peer)
	}
	return localAddr.IP.To4(), nil
}

func bindToDevice(ifaceName string) func(string, string, syscall.RawConn) error {
	return func(_, _ string, rawConn syscall.RawConn) error {
		var socketErr error
		controlErr := rawConn.Control(func(fd uintptr) {
			socketErr = unix.SetsockoptString(int(fd), unix.SOL_SOCKET, unix.SO_BINDTODEVICE, ifaceName)
		})
		if controlErr != nil {
			return controlErr
		}
		return socketErr
	}
}

func resolveNeighbor(path, ifaceName string, localIP, peerIP, nextHop net.IP, port uint16) (net.HardwareAddr, error) {
	dialer := net.Dialer{
		LocalAddr: &net.UDPAddr{IP: localIP},
		Control:   bindToDevice(ifaceName),
	}
	conn, err := dialer.Dial("udp4", net.JoinHostPort(peerIP.String(), strconv.Itoa(int(port))))
	if err != nil {
		return nil, fmt.Errorf("prime neighbor resolution for %s: %w", nextHop, err)
	}
	_, writeErr := conn.Write([]byte{0})
	_ = conn.Close()
	if writeErr != nil {
		return nil, fmt.Errorf("prime neighbor resolution for %s: %w", nextHop, writeErr)
	}

	deadline := time.Now().Add(peerResolutionTimeout)
	ticker := time.NewTicker(neighborPollInterval)
	defer ticker.Stop()
	for {
		mac, lookupErr := lookupARP(path, ifaceName, nextHop)
		if lookupErr == nil {
			return mac, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("resolve next-hop MAC for %s on %q: %w", nextHop, ifaceName, lookupErr)
		}
		<-ticker.C
	}
}

func lookupARP(path, ifaceName string, target net.IP) (net.HardwareAddr, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 6 || fields[0] != target.String() || fields[5] != ifaceName {
			continue
		}
		flags, flagsErr := strconv.ParseUint(fields[2], 0, 16)
		if flagsErr != nil || flags&2 == 0 {
			continue
		}
		mac, parseErr := net.ParseMAC(fields[3])
		if parseErr == nil && len(mac) == 6 {
			return mac, nil
		}
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return nil, scanErr
	}
	return nil, errors.New("neighbor is not reachable")
}
