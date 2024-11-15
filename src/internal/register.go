package internal

import (
	"fmt"
	"net"
	"net/netip"
)

func RegistInterfaces(ifaceManager *InterfaceManager) error {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Can't Retrieve interfaces")
		return err
	}

	for _, iface := range interfaces {
		isUp := iface.Flags&net.FlagUp != 0
		(*ifaceManager).NewLink(iface.Name, isUp)

		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			prefix, err := netip.ParsePrefix(addr.String())
			if err != nil {
				fmt.Println("Failed to ParsePrefix")
				continue
			}

			(*ifaceManager).NewIPAddr(iface.Name, prefix.Addr())
		}
	}

	return nil
}
