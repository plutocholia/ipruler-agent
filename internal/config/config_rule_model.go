package config

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

type RuleModel struct {
	SourceIP string `yaml:"sourceIP"`
	Table    int    `yaml:"table"`
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
