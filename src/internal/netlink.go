package internal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/netip"
	"os"
	"syscall"
)

type MsgType int

const (
	NewIPAddrMsg MsgType = iota
	DelIPAddrMsg
	UpLinkMsg
	DownLinkMsg
)

type NetlinkMsg struct {
	MsgType       MsgType
	InterfaceName string
	Addr          netip.Addr
}

// CreateNetlinkSocket create netlink socket
func createNetlinkSocket() (int, error) {
	// create socket
	fd, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW, syscall.NETLINK_ROUTE)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create socket: %v\n", err)
		return -1, err
	}

	// set address family and groups
	sa := &syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Groups: (1 << (syscall.RTNLGRP_LINK - 1)) |
			(1 << (syscall.RTNLGRP_IPV4_IFADDR - 1)) |
			(1 << (syscall.RTNLGRP_IPV6_IFADDR - 1)),
	}

	// bind
	if err := syscall.Bind(fd, sa); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bind socket: %v\n", err)
		return -1, err
	}

	return fd, err
}

func handleNetlinkMessage(buf []byte) []NetlinkMsg {
	var netlinkMsgs []NetlinkMsg

	for len(buf) > 0 {
		// parse netlink message header
		nlHdr := syscall.NlMsghdr{}
		binary.Read(bytes.NewReader(buf), binary.LittleEndian, &nlHdr)

		if nlHdr.Len <= 0 {
			break
		}

		// switch netlink message type
		if nlHdr.Type == syscall.RTM_NEWADDR {
			ifaceName, ipAddr := parseAddrMessage(buf[syscall.NLMSG_HDRLEN:])
			netlinkMsg := NetlinkMsg{MsgType: NewIPAddrMsg, InterfaceName: ifaceName, Addr: ipAddr}
			netlinkMsgs = append(netlinkMsgs, netlinkMsg)
		} else if nlHdr.Type == syscall.RTM_DELADDR {
			ifaceName, ipAddr := parseAddrMessage(buf[syscall.NLMSG_HDRLEN:])
			netlinkMsg := NetlinkMsg{MsgType: DelIPAddrMsg, InterfaceName: ifaceName, Addr: ipAddr}
			netlinkMsgs = append(netlinkMsgs, netlinkMsg)
		} else if nlHdr.Type == syscall.RTM_NEWLINK {
			ifaceName, isUp := parseLinkMessage(buf[syscall.NLMSG_HDRLEN:])
			if isUp {
				netlinkMsg := NetlinkMsg{MsgType: UpLinkMsg, InterfaceName: ifaceName}
				netlinkMsgs = append(netlinkMsgs, netlinkMsg)
			} else {
				netlinkMsg := NetlinkMsg{MsgType: DownLinkMsg, InterfaceName: ifaceName}
				netlinkMsgs = append(netlinkMsgs, netlinkMsg)
			}
		}

		// next message
		buf = buf[nlHdr.Len:]
	}

	return netlinkMsgs
}

func parseAddrMessage(buf []byte) (string, netip.Addr) {
	var ifaceName string
	var ipAddr netip.Addr

	// read ifaddrmsg
	ifaMsg := syscall.IfAddrmsg{}
	binary.Read(bytes.NewReader(buf), binary.LittleEndian, &ifaMsg)

	// read rtaddrs
	attrBuf := buf[syscall.SizeofIfAddrmsg:]
	for len(attrBuf) > syscall.SizeofRtAttr {
		rta := syscall.RtAttr{}
		binary.Read(bytes.NewReader(attrBuf), binary.LittleEndian, &rta)

		if rta.Len <= 0 {
			break
		}

		// if exist IFLA_IFNAME attr. get IP address.
		if rta.Type == syscall.IFLA_IFNAME {
			ifaceNameBinary := attrBuf[syscall.SizeofRtAttr:rta.Len]
			ifaceName = string(ifaceNameBinary)
		}

		// if exist IFA_LOCAL attr. get IP address.
		if rta.Type == syscall.IFA_LOCAL {
			ipBinary := attrBuf[syscall.SizeofRtAttr:rta.Len]
			var ok bool

			if ifaMsg.Family == syscall.AF_INET {
				ipAddr, ok = netip.AddrFromSlice(ipBinary[:4]) // IPv4)
				if !ok {
					fmt.Println("Invalid IP byte slice")
				}
			} else if ifaMsg.Family == syscall.AF_INET6 {
				ipAddr, ok = netip.AddrFromSlice(ipBinary[:16]) // IPv6
				if !ok {
					fmt.Println("Invalid IP byte slice")
				}
			}
		}

		// next attr
		attrBuf = attrBuf[rta.Len:]
	}

	return ifaceName, ipAddr
}

func parseLinkMessage(buf []byte) (string, bool) {
	var ifaceName string
	var isUp bool

	// read ifaddrmsg
	ifinfoMsg := syscall.IfInfomsg{}
	binary.Read(bytes.NewReader(buf), binary.LittleEndian, &ifinfoMsg)

	// read IFF_UP
	isUp = ifinfoMsg.Flags&syscall.IFF_UP != 0

	// read rtaddrs
	attrBuf := buf[syscall.SizeofIfInfomsg:]
	for len(attrBuf) > syscall.SizeofRtAttr {
		rta := syscall.RtAttr{}
		binary.Read(bytes.NewReader(attrBuf), binary.LittleEndian, &rta)

		if rta.Len <= 0 {
			break
		}

		// if exist IFLA_IFNAME attr. get IP address.
		if rta.Type == syscall.IFLA_IFNAME {
			ifaceNameBinary := attrBuf[syscall.SizeofRtAttr:rta.Len]
			ifaceName = string(ifaceNameBinary)
		}

		// next attr
		attrBuf = attrBuf[rta.Len:]
	}

	return ifaceName, isUp
}

// RoutineNetlinkMessageReceive receive netlink message
func RoutineNetlinkMessageReceive(nlmsgCh chan []NetlinkMsg) {
	fd, err := createNetlinkSocket()
	if err != nil {
		return
	}
	defer syscall.Close(fd)

	fmt.Println("Listening for Netlink messages...")

	// loop of receive netlink message
	buf := make([]byte, 4096)
	for {
		// receive message
		n, _, err := syscall.Recvfrom(fd, buf, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error receiving message: %v\n", err)
			continue
		}

		// handle netlink message
		nlmsg := handleNetlinkMessage(buf[:n])
		nlmsgCh <- nlmsg
	}
}
