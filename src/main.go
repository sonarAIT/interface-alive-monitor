package main

import (
	"fmt"

	"github.com/interface-alive-monitor/internal"
)

func main() {
	nlmsgCh := make(chan []internal.NetlinkMsg, 64)
	defer close(nlmsgCh)
	go internal.RoutineNetlinkMessageReceive(nlmsgCh)

	var ifaceManager internal.InterfaceManager
	internal.RegistInterfaces(&ifaceManager)
	ifaceManager.Print()

	for {
		select {
		case nlmsgs := <-nlmsgCh:
			fmt.Println("-----")
			for _, nlmsg := range nlmsgs {
				fmt.Println("nlmsg: ", nlmsg)
				switch nlmsg.MsgType {
				case internal.NewIPAddrMsg:
					if nlmsg.Addr.String() == "invalid IP" {
						continue
					}
					ifaceManager.NewIPAddr(nlmsg.InterfaceName, nlmsg.Addr)
				case internal.DelIPAddrMsg:
					if nlmsg.Addr.String() == "invalid IP" {
						continue
					}
					ifaceManager.DelIPAddr(nlmsg.InterfaceName, nlmsg.Addr)
				case internal.UpLinkMsg:
					ifaceManager.UpLink(nlmsg.InterfaceName)
				case internal.DownLinkMsg:
					ifaceManager.DownLink(nlmsg.InterfaceName)
				}
				ifaceManager.Print()
			}
		}
	}
}
