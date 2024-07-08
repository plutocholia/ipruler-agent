package ipruler

import (
	"fmt"
	"log"
	"os"

	"github.com/plutocholia/ipruler/internal/config"
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

func PersistState(configLifeCycle *ConfigLifeCycle) {
	headContent := `#!/bin/bash
LOCK_FILE="/var/run/networkd-dispatcher-routable.lock" 

#if [ -f "$LOCK_FILE" ]; then
#	echo "Script already executed once. Exiting."
#	exit 0
#fi

`

	footerContent := `
touch $LOCK_FILE
echo "Script executed and lock file created."
`

	mainContent := ``

	// Convert rule list to it's corresponding `ip rule add` linux command.
	for _, rule := range configLifeCycle.CurrentConfig.Rules {
		mainContent += fmt.Sprintf("ip rule add from %s table %d;\n", rule.Src.IP.String(), rule.Table)
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

	log.Println("persisting configurations at ", PERSIST_PATH)
}

func SyncState(configLifeCycle *ConfigLifeCycle) {
	machineRules, _ := netlink.RuleList(netlink.FAMILY_V4)
	curSettings := configLifeCycle.CurrentConfig.Settings
	curRules := configLifeCycle.CurrentConfig.Rules
	// delete rules that are not present in config but it's present in the machine
	// if configLifeCycle.CurrentConfig.Settings.TableHardSync
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
					log.Printf("rule (%s) is deleted (not present in the configuration)", machineRule)
					err := netlink.RuleDel(&machineRule)
					if err != nil {
						log.Fatalf("Error in deleting (%s) : %s", machineRule, err)
					}
				}
			}
		}
	}
	// delete removed rules based on old config
	if configLifeCycle.OldConfig != nil {
		oldRules := configLifeCycle.OldConfig.Rules
		for _, oldRule := range oldRules {
			ruleExists := false
			for _, curRule := range curRules {
				if oldRule.Src.IP.Equal(curRule.Src.IP) && oldRule.Table == curRule.Table {
					ruleExists = true
					break
				}
			}
			if !ruleExists {
				log.Printf("rule (%s) is deleted", oldRule)
				netlink.RuleDel(oldRule)
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
			// Adding the rule if it's not exists
			log.Printf("rule (%s) is added", rule)
			netlink.RuleAdd(rule)
		}
	}
}
