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
	fmt.Println(ifaceManager)

	for {
		select {
		case nlmsg := <-nlmsgCh:
			fmt.Println(nlmsg)
		}
	}
}
