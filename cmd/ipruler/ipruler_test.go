package ipruler

import (
	"log"
	"testing"
)

func TestAddingRules(t *testing.T) {
	var c1 string = `
rules:
- sourceIP: 172.31.201.11/32
  table: 101
- sourceIP: 172.31.201.12/32
  table: 102
`

	var c2 string = `
rules:
- sourceIP: 172.31.201.11/32
  table: 101
- sourceIP: 172.31.201.12/32
  table: 102
- sourceIP: 172.31.201.13/32
  table: 103
`

	configLifeCycle := CreateConfigLifeCycle()

	log.Println("adding c1")
	configLifeCycle.Update([]byte(c1))
	SyncState(configLifeCycle)

	log.Println("adding c2")
	configLifeCycle.Update([]byte(c2))
	log.Println("Current Configuration: ", configLifeCycle.CurrentConfig.String())
	log.Println("Old Configuration: ", configLifeCycle.OldConfig.String())
	SyncState(configLifeCycle)
}

func TestRemovingRules(t *testing.T) {
	var c1 string = `
rules:
- sourceIP: 172.31.201.11/32
  table: 101
- sourceIP: 172.31.201.12/32
  table: 102
`

	var c2 string = `
rules:
- sourceIP: 172.31.201.11/32
  table: 101
`

	configLifeCycle := CreateConfigLifeCycle()

	log.Println("adding c1")
	configLifeCycle.Update([]byte(c1))
	SyncState(configLifeCycle)

	// log.Println("Sleeping for 10s ...")
	// time.Sleep(10 * time.Second)

	log.Println("adding c2")
	configLifeCycle.Update([]byte(c2))
	log.Println("Current Configuration: ", configLifeCycle.CurrentConfig.String())
	log.Println("Old Configuration: ", configLifeCycle.OldConfig.String())
	SyncState(configLifeCycle)
}

func TestIPRuleConfiguration(t *testing.T) {
	var c1 string = `
settings:
 table-hard-sync:
 - 102
 - 101
rules:
- sourceIP: 172.31.201.11/32
  table: 101
- sourceIP: 172.31.201.12/32
  table: 102
`

	configLifeCycle := CreateConfigLifeCycle()
	configLifeCycle.Update([]byte(c1))
	log.Println(configLifeCycle.CurrentConfig.Settings)
	SyncState(configLifeCycle)
	//	log.Println("adding c1")
	// log.Println(configLifeCycle.CurrentConfig)
	// SyncState(configLifeCycle)
}
