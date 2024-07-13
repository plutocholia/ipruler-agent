package ipruler

import (
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

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

	// Convert rule list to it's corresponding `ip rule add` linux command.
	for _, rule := range configLifeCycle.CurrentConfig.Rules {
		mainContent += fmt.Sprintf("ip rule add from %s table %d;\n", rule.Src.IP.String(), rule.Table)
	}

	// Contert route list to it's corresponding `ip route add` linux command
	for _, route := range configLifeCycle.CurrentConfig.Routes {
		dst := "default"
		if route.Dst != nil {
			dst = route.Dst.IP.String()
		}
		mainContent += fmt.Sprintf("ip route add to %s via %s table %d;\n", dst, route.Gw.String(), route.Table)
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

func SyncRulesState(configLifeCycle *ConfigLifeCycle) {
	machineRules, _ := netlink.RuleList(netlink.FAMILY_V4)
	curSettings := configLifeCycle.CurrentConfig.Settings
	curRules := configLifeCycle.CurrentConfig.Rules
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

// custom netlink.route equality check (not used)
func CheckRouteEquality(r *netlink.Route, x *netlink.Route) bool {
	if r.Gw.Equal(x.Gw) && r.Table == x.Table {
		if (r.Dst == nil && x.Dst == nil) ||
			(r.Dst != nil && x.Dst != nil && r.Dst.IP.Equal(x.Dst.IP)) {
			return true
		}
	}
	return false
}

func PrintFullRoute(r *netlink.Route) string {
	elems := []string{}
	if len(r.MultiPath) == 0 {
		elems = append(elems, fmt.Sprintf("Ifindex: %d", r.LinkIndex))
	}
	if r.MPLSDst != nil {
		elems = append(elems, fmt.Sprintf("Dst: %d", r.MPLSDst))
	} else {
		elems = append(elems, fmt.Sprintf("Dst: %s", r.Dst))
	}
	if r.NewDst != nil {
		elems = append(elems, fmt.Sprintf("NewDst: %s", r.NewDst))
	}
	if r.Encap != nil {
		elems = append(elems, fmt.Sprintf("Encap: %s", r.Encap))
	}
	elems = append(elems, fmt.Sprintf("Src: %s", r.Src))
	if len(r.MultiPath) > 0 {
		elems = append(elems, fmt.Sprintf("Gw: %s", r.MultiPath))
	} else {
		elems = append(elems, fmt.Sprintf("Gw: %s", r.Gw))
	}
	elems = append(elems, fmt.Sprintf("Flags: %s", r.ListFlags()))
	elems = append(elems, fmt.Sprintf("Table: %d", r.Table))

	// Added
	elems = append(elems, fmt.Sprintf("ILinkIndex: %d", r.ILinkIndex))
	elems = append(elems, fmt.Sprintf("Scope: %d", r.Scope))
	elems = append(elems, fmt.Sprintf("Protocol: %d", r.Protocol))
	elems = append(elems, fmt.Sprintf("Priority: %d", r.Priority))
	elems = append(elems, fmt.Sprintf("Type: %d", r.Type))
	elems = append(elems, fmt.Sprintf("Tos: %d", r.Tos))
	elems = append(elems, fmt.Sprintf("MTU: %d", r.MTU))
	elems = append(elems, fmt.Sprintf("AdvMSS: %d", r.AdvMSS))
	elems = append(elems, fmt.Sprintf("Hoplimit: %d", r.Hoplimit))

	return fmt.Sprintf("{%s}", strings.Join(elems, " "))
}

func SyncRoutesState(configLifeCycle *ConfigLifeCycle) {
	curRoutes := configLifeCycle.CurrentConfig.Routes
	curSettings := configLifeCycle.CurrentConfig.Settings

	// remove routes base on table-hard-sync
	for table := range curSettings.TableHardSync {
		machineRoutes, _ := netlink.RouteListFiltered(netlink.FAMILY_V4, &netlink.Route{Table: table}, netlink.RT_FILTER_TABLE)
		for _, machineRoute := range machineRoutes {
			routeExists := false
			for _, route := range curRoutes {
				// fmt.Printf("route: %s\n", PrintFullRoute(route))
				// fmt.Printf("machi: %s\n", PrintFullRoute(&machineRoute))
				// if CheckRouteEquality(&machineRoute, route)
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
	if configLifeCycle.OldConfig != nil {
		oldRoutes := configLifeCycle.OldConfig.Routes
		for _, oldRoute := range oldRoutes {
			routeExists := false
			for _, curRoute := range curRoutes {
				// if CheckRouteEquality(oldRoute, curRoute)
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

func SyncState(configLifeCycle *ConfigLifeCycle) {
	SyncRulesState(configLifeCycle)
	SyncRoutesState(configLifeCycle)
}
