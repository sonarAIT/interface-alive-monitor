package main

import (
	"net/netip"
	"testing"
)

func TestInterfaceManager(t *testing.T) {
	ifacem := &InterfaceManager{}

	// Add a new link
	ifacem.NewLink("eth0", true)
	if len(ifacem.Interfaces) != 1 {
		t.Errorf("expected 1 interface, got %d", len(ifacem.Interfaces))
	}
	if ifacem.Interfaces[0].Name != "eth0" || !ifacem.Interfaces[0].State {
		t.Errorf("unexpected interface state or name")
	}

	// Add a new IPv4 address
	addr := netip.MustParseAddr("192.168.1.1")
	ifacem.NewIPAddr("eth0", addr)
	if len(ifacem.Interfaces[0].IPv4Addr) != 1 || ifacem.Interfaces[0].IPv4Addr[0] != addr {
		t.Errorf("failed to add IPv4 address")
	}

	// Add a new IPv6 address
	ipv6Addr := netip.MustParseAddr("2001:db8::1")
	ifacem.NewIPAddr("eth0", ipv6Addr)
	if len(ifacem.Interfaces[0].IPv6Addr) != 1 || ifacem.Interfaces[0].IPv6Addr[0] != ipv6Addr {
		t.Errorf("failed to add IPv6 address")
	}

	// Remove IPv4 address
	ifacem.DelIPAddr("eth0", addr)
	if len(ifacem.Interfaces[0].IPv4Addr) != 0 {
		t.Errorf("failed to remove IPv4 address")
	}

	// Down the link
	ifacem.DownLink("eth0")
	if ifacem.Interfaces[0].State {
		t.Errorf("expected interface to be down")
	}

	// Up the link
	ifacem.UpLink("eth0")
	if !ifacem.Interfaces[0].State {
		t.Errorf("expected interface to be up")
	}

	// Delete the link
	ifacem.DelLink("eth0")
	if len(ifacem.Interfaces) != 0 {
		t.Errorf("failed to delete interface")
	}

	// Add a new link by NewIPAddr
	addr2 := netip.MustParseAddr("10.0.0.1")
	ifacem.NewIPAddr("eth1", addr2)

	if len(ifacem.Interfaces) != 1 {
		t.Errorf("expected 1 interfaces, got %d", len(ifacem.Interfaces))
	}

	if ifacem.Interfaces[0].Name != "eth1" || len(ifacem.Interfaces[0].IPv4Addr) != 1 || ifacem.Interfaces[0].IPv4Addr[0] != addr2 {
		t.Errorf("New interface was not added correctly")
	}
}
