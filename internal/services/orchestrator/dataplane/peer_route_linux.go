//go:build cgo && linux

package dataplane

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

const routeMessageSize = unix.SizeofNlMsghdr + unix.SizeofRtMsg

func routeNextHop(ifaceName string, localIP, peer net.IP) (net.IP, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil, fmt.Errorf("find interface %q: %w", ifaceName, err)
	}
	request := routeRequest(iface.Index, localIP.To4(), peer.To4())
	reply, err := queryRoute(request)
	if err != nil {
		return nil, fmt.Errorf("resolve route to %s on %q: %w", peer, ifaceName, err)
	}
	return routeGateway(reply, iface.Index, peer.To4())
}

func routeRequest(ifaceIndex int, localIP, peer net.IP) []byte {
	request := make([]byte, routeMessageSize)
	binary.NativeEndian.PutUint32(request[0:4], uint32(routeMessageSize))
	binary.NativeEndian.PutUint16(request[4:6], unix.RTM_GETROUTE)
	binary.NativeEndian.PutUint16(request[6:8], unix.NLM_F_REQUEST)
	request[16] = unix.AF_INET
	request[17] = 32
	request[18] = 32
	request = appendRouteAttribute(request, unix.RTA_DST, peer)
	request = appendRouteAttribute(request, unix.RTA_SRC, localIP)
	index := make([]byte, 4)
	binary.NativeEndian.PutUint32(index, uint32(ifaceIndex))
	request = appendRouteAttribute(request, unix.RTA_OIF, index)
	binary.NativeEndian.PutUint32(request[0:4], uint32(len(request)))
	return request
}

func appendRouteAttribute(message []byte, attrType uint16, value []byte) []byte {
	length := unix.SizeofRtAttr + len(value)
	aligned := (length + 3) &^ 3
	attr := make([]byte, aligned)
	binary.NativeEndian.PutUint16(attr[0:2], uint16(length))
	binary.NativeEndian.PutUint16(attr[2:4], attrType)
	copy(attr[unix.SizeofRtAttr:], value)
	return append(message, attr...)
}

func queryRoute(request []byte) (syscall.NetlinkMessage, error) {
	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW|unix.SOCK_CLOEXEC, unix.NETLINK_ROUTE)
	if err != nil {
		return syscall.NetlinkMessage{}, err
	}
	defer unix.Close(fd)
	if err = unix.Sendto(fd, request, 0, &unix.SockaddrNetlink{Family: unix.AF_NETLINK}); err != nil {
		return syscall.NetlinkMessage{}, err
	}
	buffer := make([]byte, 8192)
	count, _, err := unix.Recvfrom(fd, buffer, 0)
	if err != nil {
		return syscall.NetlinkMessage{}, err
	}
	messages, err := syscall.ParseNetlinkMessage(buffer[:count])
	if err != nil {
		return syscall.NetlinkMessage{}, err
	}
	return selectRouteMessage(messages)
}

func selectRouteMessage(messages []syscall.NetlinkMessage) (syscall.NetlinkMessage, error) {
	for _, message := range messages {
		if message.Header.Type == unix.RTM_NEWROUTE {
			return message, nil
		}
		if message.Header.Type == unix.NLMSG_ERROR && len(message.Data) >= 4 {
			errno := -int32(binary.NativeEndian.Uint32(message.Data[:4]))
			return syscall.NetlinkMessage{}, syscall.Errno(errno)
		}
	}
	return syscall.NetlinkMessage{}, errors.New("kernel returned no route")
}

func routeGateway(message syscall.NetlinkMessage, ifaceIndex int, peer net.IP) (net.IP, error) {
	attrs, err := syscall.ParseNetlinkRouteAttr(&message)
	if err != nil {
		return nil, err
	}
	selectedInterface := 0
	var gateway net.IP
	for _, attr := range attrs {
		switch attr.Attr.Type {
		case unix.RTA_OIF:
			if len(attr.Value) >= 4 {
				selectedInterface = int(binary.NativeEndian.Uint32(attr.Value))
			}
		case unix.RTA_GATEWAY:
			if len(attr.Value) >= net.IPv4len {
				gateway = net.IP(attr.Value[:net.IPv4len]).To4()
			}
		}
	}
	if selectedInterface != ifaceIndex {
		return nil, fmt.Errorf("kernel selected interface index %d, want %d", selectedInterface, ifaceIndex)
	}
	if gateway != nil {
		return gateway, nil
	}
	return peer.To4(), nil
}
