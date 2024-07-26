package ipruler

import (
	"log"
	"net"
	"syscall"
	"testing"

	"github.com/vishvananda/netlink"
)

func TestAddingRules(t *testing.T) {
	var c1 string = `
rules:
- from: 172.31.201.11/32
  table: 101
- from: 172.31.201.12/32
  table: 102
`

	var c2 string = `
rules:
- from: 172.31.201.11/32
  table: 101
- from: 172.31.201.12/32
  table: 102
- from: 172.31.201.13/32
  table: 103
`

	configLifeCycle := CreateConfigLifeCycle()

	log.Println("adding c1")
	configLifeCycle.Update([]byte(c1))
	configLifeCycle.SyncState()

	log.Println("adding c2")
	configLifeCycle.Update([]byte(c2))
	log.Println("Current Configuration: ", configLifeCycle.CurrentConfig.String())
	log.Println("Old Configuration: ", configLifeCycle.OldConfig.String())
	configLifeCycle.SyncState()
}

func TestRemovingRules(t *testing.T) {
	var c1 string = `
rules:
- from: 172.31.201.11/32
  table: 101
- from: 172.31.201.12/32
  table: 102
`

	var c2 string = `
rules:
- from: 172.31.201.11/32
  table: 101
`

	configLifeCycle := CreateConfigLifeCycle()

	log.Println("adding c1")
	configLifeCycle.Update([]byte(c1))
	configLifeCycle.SyncState()

	// log.Println("Sleeping for 10s ...")
	// time.Sleep(10 * time.Second)

	log.Println("adding c2")
	configLifeCycle.Update([]byte(c2))
	log.Println("Current Configuration: ", configLifeCycle.CurrentConfig.String())
	log.Println("Old Configuration: ", configLifeCycle.OldConfig.String())
	configLifeCycle.SyncState()
}

func TestRule_HardConfiguration(t *testing.T) {
	var c1 string = `
settings:
 table-hard-sync:
 - 102
 - 101
rules:
- from: 172.31.201.11/32
  table: 101
- from: 172.31.201.12/32
  table: 102
`

	configLifeCycle := CreateConfigLifeCycle()
	configLifeCycle.Update([]byte(c1))
	log.Println(configLifeCycle.CurrentConfig.Settings)
	configLifeCycle.SyncRulesState()
	//	log.Println("adding c1")
	// log.Println(configLifeCycle.CurrentConfig)
	// SyncRulesState(configLifeCycle)
}

func TestRoute_AddAndHardSync(t *testing.T) {
	testRoute := &netlink.Route{LinkIndex: 4, Gw: net.ParseIP("172.31.201.1"), Dst: &net.IPNet{IP: net.ParseIP("172.31.201.3"), Mask: net.CIDRMask(32, 32)}, Table: 102}
	err := netlink.RouteAdd(testRoute)
	if err == syscall.EEXIST {
		log.Printf("Test Route (%s) exists.", testRoute)
	} else if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Test Route (%s) is added", testRoute)
	}

	var c1 string = `
settings:
  table-hard-sync:
  - 102
  - 101
rules:
- from: 172.31.201.11/32
  table: 101
- from: 172.31.201.12/32
  table: 102
routes:
- to: default
  via: 172.31.201.1
  table: 102
- to: 172.31.201.4/32
  via: 172.31.201.1
  table: 102
`
	configLifeCycle := CreateConfigLifeCycle()
	configLifeCycle.Update([]byte(c1))
	configLifeCycle.SyncRoutesState()
	configLifeCycle.PersistState()
}

func TestRoute_ConfigChange(t *testing.T) {
	var c1 string = `
settings:
  table-hard-sync:
  - 101
rules:
- from: 172.31.201.11/32
  table: 101
- from: 172.31.201.12/32
  table: 102
routes:
- to: default
  via: 172.31.201.1
  table: 102
- to: 172.31.201.4/32
  via: 172.31.201.1
  table: 102
`

	var c2 string = `
settings:
  table-hard-sync:
  - 101
rules:
- from: 172.31.201.11/32
  table: 101
- from: 172.31.201.12/32
  table: 102
routes:
#- to: default
#  via: 172.31.201.1
#  table: 102
- to: 172.31.201.4/32
  via: 172.31.201.1
  table: 102
`
	configLifeCycle := CreateConfigLifeCycle()
	configLifeCycle.Update([]byte(c1))
	configLifeCycle.SyncRoutesState()
	// PersistState(configLifeCycle)

	configLifeCycle.Update([]byte(c2))
	configLifeCycle.SyncRoutesState()
	// PersistState(configLifeCycle)
}
