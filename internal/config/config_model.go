package config

import (
	"fmt"
	"log"

	"gopkg.in/yaml.v2"
)

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
