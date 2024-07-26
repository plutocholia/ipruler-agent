package config

import (
	"fmt"
	"log"
	"net"

	"github.com/plutocholia/ipruler/internal/utils"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

type RouteModel struct {
	To       string `yaml:"to"`
	Via      string `yaml:"via"`
	Table    int    `yaml:"table"`
	Dev      string `yaml:"dev"`
	Protocol string `yaml:"protocol"`
	OnLink   bool   `yaml:"on-link"`
	Scope    string `yaml:"scope"`
}

func getReachableLink(ip net.IP) netlink.Link {
	routes, err := netlink.RouteGet(ip)
	if err != nil {
		log.Fatalf("RouteGet failed: %v", err)
	}
	link, err := netlink.LinkByIndex(routes[0].LinkIndex)
	if err != nil {
		log.Fatalf("LinkByIndex failed: %v", err)
	}
	return link
}

func (r *RouteModel) String() string {
	return fmt.Sprintf("to: %s - via: %s - table: %d", r.To, r.Via, r.Table)
}

func (r *RouteModel) ToNetlink() interface{} {
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
		link := getReachableLink(route.Gw)
		route.LinkIndex = link.Attrs().Index
	} else { // add `LinkIndex` to route based on route.Dev
		link, err := netlink.LinkByName(r.Dev)
		if err != nil {
			log.Fatalf("Failed to get the network interface: %v\n", err)
		}
		route.LinkIndex = link.Attrs().Index
	}

	// handle protocol
	if r.Protocol != "" {
		if value, exists := utils.RouteProtocols[r.Protocol]; exists {
			route.Protocol = value
		} else {
			log.Fatalf("Route Protocol '%s' does not exist.\n", r.Protocol)
		}
	} else {
		// add `Protocol` to route (`ip route add` command sets RTPROT_BOOT when protocol is not defined)
		route.Protocol = unix.RTPROT_BOOT
	}

	// handle flag
	if r.OnLink {
		if value, exists := utils.RouteFlags["onlink"]; exists {
			route.Flags = value
		} else {
			log.Fatalf("Route flags '%s' does not exist.\n", "onlink")
		}
	} else {
		route.Flags = 0
	}

	// handle Scope
	if r.Scope != "" {
		if value, exists := utils.RouteScopes[r.Scope]; exists {
			route.Scope = netlink.Scope(value)
		} else {
			log.Fatalf("Route scope '%s' does not exist.\n", r.Scope)
		}
	} else {
		route.Scope = unix.RT_SCOPE_UNIVERSE
	}

	// add `Type` to route (`ip route add` command sets RTN_UNICAST when type is not defined)
	route.Type = unix.RTN_UNICAST

	return route
}
