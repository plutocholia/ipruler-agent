package config

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

type RuleModel struct {
	From  string `yaml:"from"`
	Table int    `yaml:"table"`
}

// RuleModel Methods
func (r *RuleModel) String() string {
	return fmt.Sprintf("Src: %s - Table: %d", r.From, r.Table)
}

func (r *RuleModel) ToNetlink() interface{} {
	rule := netlink.NewRule()
	rule.Table = r.Table

	if _, ipnet, err := net.ParseCIDR(r.From); err != nil {
		// Handle the Error!
	} else {
		rule.Src = ipnet
	}

	return rule
}
