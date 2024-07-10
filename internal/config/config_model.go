package config

import (
	"fmt"
	"log"
	"net"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v2"
)

type RuleModel struct {
	SourceIP string `yaml:"sourceIP"`
	Table    int    `yaml:"table"`
}

type RouteModel struct {
	To    string `yaml:"to"`
	Via   string `yaml:"via"`
	Table int    `yaml:"table"`
	Dev   string `yaml:"dev"`
}

type SettingsModel struct {
	TableHardSync []int `yaml:"table-hard-sync"`
}

type ConfigModel struct {
	Rules    []RuleModel   `yaml:"rules"`
	Settings SettingsModel `yaml:"settings"`
	Routes   []RouteModel  `yaml:"routes"`
}

// General Functions
func CreateConfigModel(data []byte) *ConfigModel {
	configModel := ConfigModel{}
	err := yaml.Unmarshal(data, &configModel)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return &configModel
}

// RuleModel Methods
func (r *RuleModel) String() string {
	return fmt.Sprintf("Src: %s - Table: %d", r.SourceIP, r.Table)
}

func (r *RuleModel) ToNetlinkRule() *netlink.Rule {
	rule := netlink.NewRule()
	rule.Table = r.Table

	if _, ipnet, err := net.ParseCIDR(r.SourceIP); err != nil {
		// Handle the Error!
	} else {
		rule.Src = ipnet
	}

	return rule
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

	// add `LinkIndex` to route based on route.Gw
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

	// add `Type` to route
	route.Type = unix.RTN_UNICAST

	// add `Protocol` to route (ip route add command adds RTPROT_BOOT when protocol is empty)
	route.Protocol = unix.RTPROT_BOOT

	return route
}

// ConfigModel Methods
func (c *ConfigModel) String() string {
	var res string = ""
	var rules_len int = len(c.Rules)
	var routes_len int = len(c.Routes)
	if rules_len != 0 {
		res += fmt.Sprintln("Rules:")
		for i, rule := range c.Rules {
			if i+1 == rules_len {
				res += fmt.Sprintf("\t%d => (%s)", i, rule.String())
			} else {
				res += fmt.Sprintf("\t%d => (%s)\n", i, rule.String())
			}
		}
	}
	if routes_len != 0 {
		res += fmt.Sprintln("Routes:")
		for i, route := range c.Routes {
			if i+1 == routes_len {
				res += fmt.Sprintf("\t%d => (%s)", i, route.String())
			} else {
				res += fmt.Sprintf("\t%d => (%s)\n", i, route.String())
			}
		}
	}
	return res
}
