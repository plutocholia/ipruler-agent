package ipruler

import (
	"log"
	"os"
	"syscall"

	"github.com/plutocholia/ipruler/internal/config"
	"github.com/plutocholia/ipruler/internal/utils"
	"github.com/vishvananda/netlink"
)

const (
	PERSIST_PATH = "/etc/networkd-dispatcher/routable.d/00-ipruler"
)

type ConfigLifeCycle struct {
	CurrentConfig *config.Config
	OldConfig     *config.Config
}

func CreateConfigLifeCycle() *ConfigLifeCycle {
	return &ConfigLifeCycle{}
}

func (c *ConfigLifeCycle) Update(data []byte) {
	configModel := config.CreateConfigModel(data)
	if c.CurrentConfig == nil {
		c.CurrentConfig = config.CreateConfig(configModel)
	} else {
		c.OldConfig = c.CurrentConfig
		c.CurrentConfig = config.CreateConfig(configModel)
	}
}

// It's equivalent to Update method which does syncs in proper order
func (c *ConfigLifeCycle) WeaveSync(data []byte) error {
	configModel := config.CreateConfigModel(data)

	if configModel.IsEmpty() {
		return CreateEmptyConfigError()
	}

	newConfig := c.CreateNewConfig()

	newConfig.AddSettings(configModel.Settings)
	newConfig.AddVlans(configModel.Vlans)
	c.SyncVlansState()

	newConfig.AddRoutes(configModel.Routes)
	c.SyncRoutesState()

	newConfig.AddRules(configModel.Rules)
	c.SyncRulesState()

	return nil
}

func (c *ConfigLifeCycle) CreateNewConfig() *config.Config {
	newConfig := &config.Config{}
	if c.CurrentConfig == nil {
		c.CurrentConfig = newConfig
	} else {
		c.OldConfig = c.CurrentConfig
		c.CurrentConfig = newConfig
	}
	return newConfig
}

func (c *ConfigLifeCycle) PersistState() {
	headContent := `#!/bin/bash
LOCK_FILE="/var/run/networkd-dispatcher-routable.lock" 

if [ -f "$LOCK_FILE" ]; then
	echo "Script already executed once. Exiting."
	exit 0
fi

`

	footerContent := `
touch $LOCK_FILE
echo "Script executed and lock file created."
`

	mainContent := ``

	// Contert Vlan list to it's corresponding `ip link add` linux command
	for _, vlan := range c.CurrentConfig.Vlans {
		mainContent += utils.VlanToIPCommand(vlan) + ";\n"
	}

	// Contert route list to it's corresponding `ip route add` linux command
	for _, route := range c.CurrentConfig.Routes {
		mainContent += utils.RouteToIPCommand(route) + ";\n"
	}

	// Convert rule list to it's corresponding `ip rule add` linux command.
	for _, rule := range c.CurrentConfig.Rules {
		mainContent += utils.RuleToIPCommand(rule) + ";\n"
	}

	content := headContent + mainContent + footerContent

	file, err := os.Create(PERSIST_PATH)
	if err != nil {
		log.Fatalln("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		log.Fatalln("Error writing to file:", err)
		return
	}

	err = os.Chmod(PERSIST_PATH, 0755)
	if err != nil {
		log.Fatalln("Error making file executable:", err)
		return
	}
	// log.Println("persisting configurations at ", PERSIST_PATH)
}

func (c *ConfigLifeCycle) SyncRulesState() {
	machineRules, _ := netlink.RuleList(netlink.FAMILY_V4)
	curSettings := c.CurrentConfig.Settings
	curRules := c.CurrentConfig.Rules
	// remove rules base on table-hard-sync
	if len(curSettings.TableHardSync) != 0 {
		for _, machineRule := range machineRules {
			if machineRule.Src != nil && curSettings.TableHardSync[machineRule.Table] {
				machineRuleExists := false
				for _, curRule := range curRules {
					if machineRule.Src.IP.Equal(curRule.Src.IP) && machineRule.Table == curRule.Table {
						machineRuleExists = true
						break
					}
				}
				if !machineRuleExists {
					log.Printf("[table-hard-sync] Rule (%s) does not exist in current config.", machineRule)
					err := netlink.RuleDel(&machineRule)
					if err != nil && err != syscall.ENOENT {
						log.Fatalf("[table-Hard-sync] Error in deleting (%s) : %s", machineRule, err)
					} else if err == syscall.ENOENT {
						log.Printf("[table-hard-sync] Rule (%s) has already been deleted.", machineRule)
					} else {
						log.Printf("[table-hard-sync] Rule (%s) is deleted.", machineRule)
					}
				}
			}
		}
	}
	// delete removed rules based on old config
	if c.OldConfig != nil {
		oldRules := c.OldConfig.Rules
		for _, oldRule := range oldRules {
			ruleExists := false
			for _, curRule := range curRules {
				if oldRule.Src.IP.Equal(curRule.Src.IP) && oldRule.Table == curRule.Table {
					ruleExists = true
					break
				}
			}
			if !ruleExists {
				log.Printf("[sync-removed-config] Rule (%s) is no more in current config", oldRule)
				err := netlink.RuleDel(oldRule)
				if err != nil && err != syscall.ENOENT {
					log.Fatalf("[sync-removed-config] Error in deleting rule (%s) : %s", oldRule, err)
				} else if err == syscall.ENOENT {
					log.Printf("[sync-removed-config] Rule (%s) has already been deleted.", oldRule)
				} else {
					log.Printf("[sync-removed-config] Rule (%s) is deleted.", oldRule)
				}
			}
		}
	}
	// add rules
	for _, rule := range curRules {
		ruleExists := false
		for _, machineRule := range machineRules {
			if machineRule.Src != nil && rule.Src.IP.Equal(machineRule.Src.IP) && (machineRule.Table == rule.Table) {
				ruleExists = true
				break
			}
		}
		if ruleExists {
			// log.Printf("rule (%s) exists", rule)
		} else {
			log.Printf("Rule (%s) is added", rule)
			netlink.RuleAdd(rule)
		}
	}
}

func (c *ConfigLifeCycle) SyncRoutesState() {
	curRoutes := c.CurrentConfig.Routes
	curSettings := c.CurrentConfig.Settings

	// remove routes base on table-hard-sync
	for table := range curSettings.TableHardSync {
		machineRoutes, _ := netlink.RouteListFiltered(netlink.FAMILY_V4, &netlink.Route{Table: table}, netlink.RT_FILTER_TABLE)
		for _, machineRoute := range machineRoutes {
			routeExists := false
			for _, route := range curRoutes {
				if machineRoute.Equal(*route) {
					routeExists = true
					break
				}
			}
			if !routeExists {
				log.Printf("[table-hard-sync] Route (%s) does not exist in current config.", machineRoute)
				err := netlink.RouteDel(&machineRoute)
				if err != nil && err != syscall.ESRCH {
					log.Fatalf("[table-hard-sync] Error in deleting route (%s) : %s", machineRoute, err)
				} else if err == syscall.ESRCH {
					log.Printf("[table-hard-sync] Route (%s) has already been deleted.", machineRoute)
				} else {
					log.Printf("[table-hard-sync] Route (%s) is deleted.", machineRoute)
				}
			}
		}
	}
	// delete removed routes based on old config
	if c.OldConfig != nil {
		oldRoutes := c.OldConfig.Routes
		for _, oldRoute := range oldRoutes {
			routeExists := false
			for _, curRoute := range curRoutes {
				if oldRoute.Equal(*curRoute) {
					routeExists = true
					break
				}
			}
			if !routeExists {
				log.Printf("[sync-removed-config] Route (%s) is no more in current config", oldRoute)
				err := netlink.RouteDel(oldRoute)
				if err != nil && err != syscall.ESRCH {
					log.Fatalf("[sync-removed-config] Error in deleting Route (%s) : %s", oldRoute, err)
				} else if err == syscall.ESRCH {
					log.Printf("[sync-removed-config] Route (%s) has already been deleted.", oldRoute)
				} else {
					log.Printf("[sync-removed-config] Route (%s) is deleted.", oldRoute)
				}
			}
		}
	}
	// add routes
	for _, route := range curRoutes {
		err := netlink.RouteAdd(route)
		if err == syscall.EEXIST {
			// log.Printf("Route (%s) exists.", route)
		} else if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Route (%s) is added", route)
		}
	}
}

func (c *ConfigLifeCycle) SyncVlansState() {
	curVlans := c.CurrentConfig.Vlans

	// delete removed vlans based on old config
	if c.OldConfig != nil {
		oldVlans := c.OldConfig.Vlans
		for _, oldVlan := range oldVlans {
			vlanExists := false
			for _, curVlan := range curVlans {
				if utils.VlanEquality(oldVlan, curVlan) {
					vlanExists = true
					break
				}
			}
			if !vlanExists {
				log.Printf("[sync-removed-config] vlan (%s) is no more in current config", utils.VlanToString(oldVlan))
				err := netlink.LinkDel(oldVlan)
				if err != nil && err != syscall.ESRCH {
					log.Fatalf("[sync-removed-config] Error in deleting Vlan (%s) : %s", utils.VlanToString(oldVlan), err)
				} else if err == syscall.ESRCH {
					log.Printf("[sync-removed-config] Vlan (%s) has already been deleted.", utils.VlanToString(oldVlan))
				} else {
					log.Printf("[sync-removed-config] Vlan (%s) is deleted.", utils.VlanToString(oldVlan))
				}
			}
		}
	}
	// add vlans
	for _, vlan := range curVlans {
		err := netlink.LinkAdd(vlan)
		if err == syscall.EEXIST {
			// log.Printf("Vlan (%s) exists.", utils.VlanToString(vlan))
		} else if err != nil {
			log.Fatal(err)
		} else {
			if err := netlink.LinkSetUp(vlan); err != nil {
				log.Fatalf("Unable to up the %s link", utils.VlanToString(vlan))
			}
			log.Printf("Vlan (%s) is added", utils.VlanToString(vlan))
		}
	}
}

func (c *ConfigLifeCycle) SyncState() {
	c.SyncVlansState()
	c.SyncRoutesState()
	c.SyncRulesState()
}
