package config

import (
	"fmt"
	"log"

	"github.com/vishvananda/netlink"
)

type VlanModel struct {
	Name     string `yaml:"name"`
	Link     string `yaml:"link"`
	ID       int    `yaml:"id"`
	Protocol string `yaml:"protocol"`
}

func (v *VlanModel) IsEmpty() bool {
	if v.Name == "" &&
		v.Link == "" &&
		v.ID == 0 &&
		v.Protocol == "" {
		return true
	}
	return false

}

func (v *VlanModel) String() string {
	return fmt.Sprintf("name: %s - link: %s - id: %d - protocol: %s", v.Name, v.Link, v.ID, v.Protocol)
}

func (v *VlanModel) ToNetlink() interface{} {
	parentLink, err := netlink.LinkByName(v.Link)
	if err != nil {
		log.Fatalf("Failed to find parent link: %v", err)
	}

	vlanAttrs := netlink.NewLinkAttrs()
	vlanAttrs.ParentIndex = parentLink.Attrs().Index
	vlanAttrs.Name = v.Name

	vlan := &netlink.Vlan{
		LinkAttrs: vlanAttrs,
		VlanId:    v.ID,
	}

	if v.Protocol != "" {
		if value, exists := netlink.StringToVlanProtocolMap[v.Protocol]; exists {
			vlan.VlanProtocol = value
		} else {
			log.Fatalf("Vlan protocol %s is not valid", v.Protocol)
		}
	} else {
		vlan.VlanProtocol = netlink.VLAN_PROTOCOL_8021Q
	}

	return vlan
}
