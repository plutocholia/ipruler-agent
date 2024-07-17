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
	Flag     string `yaml:"flag"`
	Scope    string `yaml:"scope"`
}

var RouteProtocols map[string]int = map[string]int{
	"kernel": unix.RTPROT_KERNEL,
	"boot":   unix.RTPROT_BOOT,
	"static": unix.RTPROT_STATIC,
}

var RouteFlags map[string]int = map[string]int{
	"onlink":    unix.RTNH_F_ONLINK,
	"pervasive": unix.RTNH_F_PERVASIVE,
}

var RouteScopes map[string]int = map[string]int{
	// "global": netlink.SCOPE_UNIVERSE,
	// "link":   netlink.SCOPE_LINK,
	// "host":   netlink.SCOPE_HOST,
	"global": unix.RT_SCOPE_UNIVERSE,
	"link":   unix.RT_SCOPE_LINK,
	"host":   unix.RT_SCOPE_HOST,
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
		if value, exists := RouteProtocols[r.Protocol]; exists {
			route.Protocol = value
		} else {
			log.Fatalf("Route Protocol '%s' does not exist.\n", r.Protocol)
		}
	} else {
		// add `Protocol` to route (`ip route add` command sets RTPROT_BOOT when protocol is not defined)
		route.Protocol = unix.RTPROT_BOOT
	}

	// handle flag
	if r.Flag != "" {
		if value, exists := RouteFlags[r.Flag]; exists {
			route.Flags = value
		} else {
			log.Fatalf("Route Flags '%s' does not exist.\n", r.Flag)
		}
	} else {
		route.Flags = 0
	}

	// handle Scope
	if r.Scope != "" {
		if value, exists := RouteScopes[r.Scope]; exists {
			route.Scope = netlink.Scope(value)
		} else {
			log.Fatalf("Route Scope '%s' does not exist.\n", r.Scope)
		}
	} else {
		route.Scope = unix.RT_SCOPE_UNIVERSE
	}

	// add `Type` to route (`ip route add` command sets RTN_UNICAST when type is not defined)
	route.Type = unix.RTN_UNICAST

	return route
}

func reverseMap(m map[string]int) map[int]string {
	n := make(map[int]string, len(m))
	for k, v := range m {
		n[v] = k
	}
	return n
}

func RouteToIPRouteCommand(r *netlink.Route) string {
	content := "ip route add"

	scope := reverseMap(RouteScopes)[int(r.Scope)]
	protocol := reverseMap(RouteProtocols)[r.Protocol]
	flag := reverseMap(RouteFlags)[r.Flags]

	dev := ""
	links, err := netlink.LinkList()
	if err != nil {
		log.Fatalf("Failed to list links: %v", err)
	}
	for _, link := range links {
		if link.Attrs().Index == r.LinkIndex {
			dev = link.Attrs().Name
			break
		}
	}

	to := "default"
	if r.Dst != nil {
		to = r.Dst.String()
	}

	content += fmt.Sprintf(" to %s", to)
	if r.Gw != nil {
		content += fmt.Sprintf(" via %s", r.Gw)
	}
	if r.Table != 0 {
		content += fmt.Sprintf(" table %d", r.Table)
	}
	content += fmt.Sprintf(" dev %s", dev)
	content += fmt.Sprintf(" proto %s", protocol)
	content += fmt.Sprintf(" scope %s", scope)
	if flag != "" {
		content += fmt.Sprintf(" %s", flag)
	}

	return content
}
