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
			ip, ok := netip.AddrFromSlice(addr.(*net.IPNet).IP)
			if !ok {
				fmt.Println("Failed to AddrFromSlice")
				continue
			}

			(*ifaceManager).NewIPAddr(iface.Name, ip)
		}
	}

	return nil
}
