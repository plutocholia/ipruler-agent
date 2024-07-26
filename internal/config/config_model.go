package config

import (
	"fmt"
	"log"

	"gopkg.in/yaml.v2"
)

type Model interface {
	String() string
	ToNetlink() interface{}
}

type ConfigModel struct {
	Rules    []RuleModel   `yaml:"rules"`
	Settings SettingsModel `yaml:"settings"`
	Routes   []RouteModel  `yaml:"routes"`
	Vlans    []VlanModel   `yaml:"vlans"`
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

func getStringFromModel(v []Model, identifier string) string {
	res := ""
	the_len := len(v)
	if the_len != 0 {
		res += fmt.Sprintf("%s:\n", identifier)
		for i, value := range v {
			res += fmt.Sprintf("\t%d => (%s)\n", i, value.String())
		}
	}
	return res
}

// func castRuleModelToModel(rules []*RuleModel) []Model {
// 	models := make([]Model, len(rules))
// 	for i, rule := range rules {
// 		models[i] = rule
// 	}
// 	return models
// }

// ConfigModel Methods
func (c *ConfigModel) String() string {
	res := ""
	rulesModelInt := make([]Model, len(c.Rules))
	for i, rule := range c.Rules {
		rulesModelInt[i] = &rule
	}
	routesModelInt := make([]Model, len(c.Routes))
	for i, route := range c.Routes {
		routesModelInt[i] = &route
	}
	vlansModelInt := make([]Model, len(c.Vlans))
	for i, vlan := range c.Vlans {
		vlansModelInt[i] = &vlan
	}
	res += getStringFromModel(rulesModelInt, "Rules")
	res += getStringFromModel(routesModelInt, "Routes")
	res += getStringFromModel(vlansModelInt, "Vlans")

	return res
}
