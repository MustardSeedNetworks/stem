//go:build cgo && linux

package dataplane

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

const (
	routeMessageSize       = unix.SizeofNlMsghdr + unix.SizeofRtMsg
	ipv4PrefixLength       = 32
	netlinkAlignment       = 4
	routeReceiveBufferSize = 8 * 1024
)

func routeNextHop(ifaceName string, localIP, peer net.IP) (net.IP, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil, fmt.Errorf("find interface %q: %w", ifaceName, err)
	}
	ifaceIndex, err := checkedUint32(iface.Index)
	if err != nil {
		return nil, fmt.Errorf("invalid interface index for %q: %w", ifaceName, err)
	}
	request, err := routeRequest(ifaceIndex, localIP.To4(), peer.To4())
	if err != nil {
		return nil, err
	}
	reply, err := queryRoute(request)
	if err != nil {
		return nil, fmt.Errorf("resolve route to %s on %q: %w", peer, ifaceName, err)
	}
	return routeGateway(reply, ifaceIndex, peer.To4())
}

func routeRequest(ifaceIndex uint32, localIP, peer net.IP) ([]byte, error) {
	request := make([]byte, routeMessageSize)
	binary.NativeEndian.PutUint32(request[0:4], uint32(routeMessageSize))
	binary.NativeEndian.PutUint16(request[4:6], unix.RTM_GETROUTE)
	binary.NativeEndian.PutUint16(request[6:8], unix.NLM_F_REQUEST)
	request[16] = unix.AF_INET
	request[17] = ipv4PrefixLength
	request[18] = ipv4PrefixLength
	var err error
	if request, err = appendRouteAttribute(request, unix.RTA_DST, peer); err != nil {
		return nil, err
	}
	if request, err = appendRouteAttribute(request, unix.RTA_SRC, localIP); err != nil {
		return nil, err
	}
	index := make([]byte, netlinkAlignment)
	binary.NativeEndian.PutUint32(index, ifaceIndex)
	if request, err = appendRouteAttribute(request, unix.RTA_OIF, index); err != nil {
		return nil, err
	}
	requestLength, err := checkedUint32(len(request))
	if err != nil {
		return nil, fmt.Errorf("encode route request length: %w", err)
	}
	binary.NativeEndian.PutUint32(request[0:4], requestLength)
	return request, nil
}

func appendRouteAttribute(message []byte, attrType uint16, value []byte) ([]byte, error) {
	length := unix.SizeofRtAttr + len(value)
	aligned := (length + netlinkAlignment - 1) &^ (netlinkAlignment - 1)
	attr := make([]byte, aligned)
	attrLength, err := checkedUint16(length)
	if err != nil {
		return nil, fmt.Errorf("encode route attribute length: %w", err)
	}
	binary.NativeEndian.PutUint16(attr[0:2], attrLength)
	binary.NativeEndian.PutUint16(attr[2:4], attrType)
	copy(attr[unix.SizeofRtAttr:], value)
	return append(message, attr...), nil
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
	buffer := make([]byte, routeReceiveBufferSize)
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
		if message.Header.Type == unix.NLMSG_ERROR && len(message.Data) >= netlinkAlignment {
			encodedErrno := binary.NativeEndian.Uint32(message.Data[:netlinkAlignment])
			if encodedErrno == 0 {
				continue
			}
			return syscall.NetlinkMessage{}, syscall.Errno(uintptr(^encodedErrno) + 1)
		}
	}
	return syscall.NetlinkMessage{}, errors.New("kernel returned no route")
}

func routeGateway(message syscall.NetlinkMessage, ifaceIndex uint32, peer net.IP) (net.IP, error) {
	attrs, err := syscall.ParseNetlinkRouteAttr(&message)
	if err != nil {
		return nil, err
	}
	var selectedInterface uint32
	var gateway net.IP
	for _, attr := range attrs {
		switch attr.Attr.Type {
		case unix.RTA_OIF:
			if len(attr.Value) >= netlinkAlignment {
				selectedInterface = binary.NativeEndian.Uint32(attr.Value)
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

func checkedUint32(value int) (uint32, error) {
	if value < 0 || uint64(value) > math.MaxUint32 {
		return 0, fmt.Errorf("value %d exceeds uint32", value)
	}
	return uint32(value), nil
}

func checkedUint16(value int) (uint16, error) {
	if value < 0 || value > math.MaxUint16 {
		return 0, fmt.Errorf("value %d exceeds uint16", value)
	}
	return uint16(value), nil
}
