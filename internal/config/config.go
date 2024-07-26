package config

import (
	"github.com/plutocholia/ipruler/internal/utils"
	"github.com/vishvananda/netlink"
)

type Config struct {
	Rules    []*netlink.Rule
	Routes   []*netlink.Route
	Vlans    []*netlink.Vlan
	Settings Settings
}

type Settings struct {
	TableHardSync map[int]bool
}

func (c *Config) String() string {
	var result string = ""

	result += "\nrules:"
	for _, rule := range c.Rules {
		result += "\n\t" + rule.String()
	}
	result += "\nroutes:"
	for _, route := range c.Routes {
		result += "\n\t" + route.String()
	}
	result += "\nvlans:"
	for _, vlan := range c.Vlans {
		result += "\n\t" + utils.VlanToString(vlan)
	}
	return result
}

func CreateConfig(configModel *ConfigModel) *Config {
	config := &Config{}

	config.AddSettings(configModel.Settings)
	config.AddVlans(configModel.Vlans)
	config.AddRoutes(configModel.Routes)
	config.AddRules(configModel.Rules)

	return config
}

func (config *Config) AddVlans(vlans []VlanModel) {
	for _, vlan := range vlans {
		if res, ok := vlan.ToNetlink().(*netlink.Vlan); ok {
			config.Vlans = append(config.Vlans, res)
		}
	}
}

func (config *Config) AddRoutes(routes []RouteModel) {
	for _, route := range routes {
		if res, ok := route.ToNetlink().(*netlink.Route); ok {
			config.Routes = append(config.Routes, res)
		}
	}
}

func (config *Config) AddRules(rules []RuleModel) {
	for _, rule := range rules {
		if res, ok := rule.ToNetlink().(*netlink.Rule); ok {
			config.Rules = append(config.Rules, res)
		}
	}
}

func (config *Config) AddSettings(settings SettingsModel) {
	config.Settings.TableHardSync = make(map[int]bool)
	for _, table := range settings.TableHardSync {
		config.Settings.TableHardSync[table] = true
	}
}
