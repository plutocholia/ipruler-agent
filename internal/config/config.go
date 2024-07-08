package config

import (
	"github.com/vishvananda/netlink"
)

type Config struct {
	Rules    []*netlink.Rule
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
	return result
}

func CreateConfig(configModel *ConfigModel) *Config {
	config := &Config{}
	// add rules from configModel to config
	for _, rule := range configModel.Rules {
		config.Rules = append(config.Rules, rule.ToNetlinkRule())
	}
	// add TableHardSync from configModel to config
	config.Settings.TableHardSync = make(map[int]bool)
	for _, table := range configModel.Settings.TableHardSync {
		config.Settings.TableHardSync[table] = true
	}
	return config
}
