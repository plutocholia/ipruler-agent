package config

import (
	"fmt"
	"log"
	"net"

	"github.com/vishvananda/netlink"
	"gopkg.in/yaml.v2"
)

type RuleModel struct {
	SourceIP string `yaml:"sourceIP"`
	Table    int    `yaml:"table"`
}

type SettingsModel struct {
	TableHardSync []int `yaml:"table-hard-sync"`
}

type ConfigModel struct {
	Rules    []RuleModel   `yaml:"rules"`
	Settings SettingsModel `yaml:"settings"`
	// Routes []RouteModel `yaml:"routes"`
}

func CreateConfigModel(data []byte) *ConfigModel {
	configModel := ConfigModel{}
	err := yaml.Unmarshal(data, &configModel)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return &configModel
}

// type RouteModel struct {
// 	SourceIP string `yaml:"sourceIP"`
// 	Table    int    `yaml:"table"`
// }

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

func (c *ConfigModel) String() string {
	var res string = ""
	var config_size int = len(c.Rules)
	for i, rule := range c.Rules {
		if i+1 == config_size {
			res += fmt.Sprintf("%d => (%s)", i, rule.String())
		} else {
			res += fmt.Sprintf("%d => (%s)\n", i, rule.String())
		}
	}
	return res
}
