package main

import (
	"net/netip"
	"sync"
)

// Interface is represent machine network interface
type Interface struct {
	Name     string
	IPv4Addr []netip.Addr
	IPv6Addr []netip.Addr
	State    bool
}

// InterfaceManager is interface list holder of machine
type InterfaceManager struct {
	Interfaces []Interface
	sync.RWMutex
}

// NewIPAddr add IP Address to IP Address list in Interface
func (ifacem *InterfaceManager) NewIPAddr(ifaceName string, addr netip.Addr) {
	ifacem.Lock()
	defer ifacem.Unlock()

	for i, iface := range ifacem.Interfaces {
		if iface.Name != ifaceName {
			continue
		}

		if addr.Is4() {
			ifacem.Interfaces[i].IPv4Addr = append(iface.IPv4Addr, addr)
		} else if addr.Is6() {
			ifacem.Interfaces[i].IPv6Addr = append(iface.IPv6Addr, addr)
		}

		return
	}

	newIface := Interface{Name: ifaceName}
	if addr.Is4() {
		newIface.IPv4Addr = []netip.Addr{addr}
	} else {
		newIface.IPv6Addr = []netip.Addr{addr}
	}
	ifacem.Interfaces = append(ifacem.Interfaces, newIface)
}

// NewIPAddr delete IP Address from IP Address list in Interface
func (ifacem *InterfaceManager) DelIPAddr(ifaceName string, addr netip.Addr) {
	ifacem.Lock()
	defer ifacem.Unlock()

	for i, iface := range ifacem.Interfaces {
		if iface.Name != ifaceName {
			continue
		}

		if addr.Is4() {
			removeAddr(&ifacem.Interfaces[i].IPv4Addr, addr)
		} else if addr.Is6() {
			removeAddr(&ifacem.Interfaces[i].IPv6Addr, addr)
		}
		return
	}
}

// NewLink add network interface to interface manager
func (ifacem *InterfaceManager) NewLink(ifaceName string, state bool) {
	ifacem.Lock()
	defer ifacem.Unlock()

	newIface := Interface{Name: ifaceName, State: state}
	ifacem.Interfaces = append(ifacem.Interfaces, newIface)
}

// NewLink delete network interface from interface manager
func (ifacem *InterfaceManager) DelLink(ifaceName string) {
	ifacem.Lock()
	defer ifacem.Unlock()

	removeIface(&ifacem.Interfaces, ifaceName)
}

// UpLink sets the state of the interface in the interface manager to up
func (ifacem *InterfaceManager) UpLink(ifaceName string) {
	ifacem.Lock()
	defer ifacem.Unlock()

	for i, iface := range ifacem.Interfaces {
		if iface.Name != ifaceName {
			continue
		}

		ifacem.Interfaces[i].State = true
		return
	}
}

// DownLink sets the state of the interface in the interface manager to down
func (ifacem *InterfaceManager) DownLink(ifaceName string) {
	ifacem.Lock()
	defer ifacem.Unlock()

	for i, iface := range ifacem.Interfaces {
		if iface.Name != ifaceName {
			continue
		}

		ifacem.Interfaces[i].State = false
		return
	}
}

// removeAddr remove IP Address from netip.Addr slice
func removeAddr(slice *[]netip.Addr, addr netip.Addr) {
	for i, v := range *slice {
		if v == addr {
			*slice = (*slice)[:i+copy((*slice)[i:], (*slice)[i+1:])]
			return
		}
	}
}

// removeIface remove Interface from Interface slice
func removeIface(slice *[]Interface, ifaceName string) {
	for i, v := range *slice {
		if v.Name == ifaceName {
			*slice = (*slice)[:i+copy((*slice)[i:], (*slice)[i+1:])]
		}
	}
}
