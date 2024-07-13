package config

import (
	"fmt"
	"log"
	"net"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

type RouteModel struct {
	To       string `yaml:"to"`
	Via      string `yaml:"via"`
	Table    int    `yaml:"table"`
	Dev      string `yaml:"dev"`
	Protocol string `yaml:"protocol"`
}

var Protocols map[string]int = map[string]int{
	"kernel": 2,
	"boot":   3,
	"static": 4,
}

// RouteModel Methods
func (r *RouteModel) String() string {
	return fmt.Sprintf("to: %s - via: %s - table: %d", r.To, r.Via, r.Table)
}

func (r *RouteModel) ToNetlinkRoute() *netlink.Route {
	route := &netlink.Route{}

	// add `Dst` to route
	if r.To == "default" {
		route.Dst = nil
	} else {
		if _, ipnet, err := net.ParseCIDR(r.To); err != nil {
			log.Fatalf("Could not Parse CIDR of (%s)", r.To)
		} else {
			route.Dst = ipnet
		}
	}

	// add `Table` to route
	route.Table = r.Table

	// add `Gw` to route
	if r.Via != "" {
		gw := net.ParseIP(r.Via)
		if gw == nil {
			log.Fatalf("Invalid gateway IP address: %s", gw)
		}
		route.Gw = gw
	}

	// add `LinkIndex` to route based on route.Gw if Dev is not defined.
	if r.Dev == "" {
		links, err := netlink.LinkList()
		if err != nil {
			log.Fatalf("Failed to list links: %v", err)
		}
		var linkIndex int
		found := false
		for _, link := range links {
			addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
			if err != nil {
				log.Fatalf("Failed to list addresses: %v", err)
			}
			for _, addr := range addrs {
				if addr.IPNet != nil && addr.IPNet.Contains(route.Gw) {
					linkIndex = link.Attrs().Index
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			log.Fatalf("No interface found supporting the IP address: %s", route.Gw)
		}
		route.LinkIndex = linkIndex
	} else { // add `LinkIndex` to route based on route.Dev
		link, err := netlink.LinkByName(r.Dev)
		if err != nil {
			log.Fatalf("Failed to get the network interface: %v\n", err)
		}
		route.LinkIndex = link.Attrs().Index
	}

	// add `Type` to route (`ip route add` command sets RTN_UNICAST when type is not defined)
	if r.Protocol != "" {
		if value, exists := Protocols[r.Protocol]; exists {
			route.Protocol = value
		} else {
			log.Fatalf("Route Protocol '%s' does not exist.\n", r.Protocol)
		}
	} else {
		// add `Protocol` to route (`ip route add` command sets RTPROT_BOOT when protocol is not defined)
		route.Protocol = unix.RTPROT_BOOT
	}

	route.Type = unix.RTN_UNICAST

	return route
}
