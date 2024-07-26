package utils

import (
	"fmt"
	"log"
	"strings"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

var RouteProtocols map[string]int = map[string]int{
	"kernel": unix.RTPROT_KERNEL,
	"boot":   unix.RTPROT_BOOT,
	"static": unix.RTPROT_STATIC,
}

var RouteFlags map[string]int = map[string]int{
	"onlink":    unix.RTNH_F_ONLINK,
	"pervasive": unix.RTNH_F_PERVASIVE,
}

var RouteScopes map[string]int = map[string]int{
	"global": unix.RT_SCOPE_UNIVERSE,
	"link":   unix.RT_SCOPE_LINK,
	"host":   unix.RT_SCOPE_HOST,
}

func reverseMap(m map[string]int) map[int]string {
	n := make(map[int]string, len(m))
	for k, v := range m {
		n[v] = k
	}
	return n
}

func RuleToIPCommand(r *netlink.Rule) string {
	return fmt.Sprintf("ip rule add from %s table %d", r.Src.IP.String(), r.Table)
}

func RouteToIPCommand(r *netlink.Route) string {
	content := "ip route add"

	scope := reverseMap(RouteScopes)[int(r.Scope)]
	protocol := reverseMap(RouteProtocols)[r.Protocol]
	flag := reverseMap(RouteFlags)[r.Flags]

	dev := ""
	links, err := netlink.LinkList()
	if err != nil {
		log.Fatalf("Failed to list links: %v", err)
	}
	for _, link := range links {
		if link.Attrs().Index == r.LinkIndex {
			dev = link.Attrs().Name
			break
		}
	}

	to := "default"
	if r.Dst != nil {
		to = r.Dst.String()
	}

	content += fmt.Sprintf(" to %s", to)
	if r.Gw != nil {
		content += fmt.Sprintf(" via %s", r.Gw)
	}
	if r.Table != 0 {
		content += fmt.Sprintf(" table %d", r.Table)
	}
	content += fmt.Sprintf(" dev %s", dev)
	content += fmt.Sprintf(" proto %s", protocol)
	content += fmt.Sprintf(" scope %s", scope)
	if flag != "" {
		content += fmt.Sprintf(" %s", flag)
	}

	return content
}

func VlanToIPCommand(v *netlink.Vlan) string {
	// Example: ip link add link eth2 name eth2.104 type vlan id 104; ip link set eth2.104 up;
	content := "ip link add"

	parentLink, _ := netlink.LinkByIndex(v.ParentIndex)
	content += fmt.Sprintf(" link %s", parentLink.Attrs().Name)

	content += fmt.Sprintf(" name %s", v.Name)

	content += "type vlan"

	content += fmt.Sprintf(" id %d", v.VlanId)

	content += fmt.Sprintf("; ip link set %s up", v.Name)

	return content
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

// custom netlink.route equality check (not used)
func RouteEquality(r *netlink.Route, x *netlink.Route) bool {
	if r.Gw.Equal(x.Gw) && r.Table == x.Table {
		if (r.Dst == nil && x.Dst == nil) ||
			(r.Dst != nil && x.Dst != nil && r.Dst.IP.Equal(x.Dst.IP)) {
			return true
		}
	}
	return false
}

func VlanEquality(v1 *netlink.Vlan, v2 *netlink.Vlan) bool {
	// Note: Equality on LinkAttrs.Index makes logical fault due to increamental behavior of this param
	return v1.VlanId == v2.VlanId &&
		v1.VlanProtocol == v2.VlanProtocol &&
		v1.LinkAttrs.ParentIndex == v2.LinkAttrs.ParentIndex &&
		v1.LinkAttrs.Name == v2.LinkAttrs.Name &&
		v1.LinkAttrs.MTU == v2.LinkAttrs.MTU &&
		v1.LinkAttrs.TxQLen == v2.LinkAttrs.TxQLen
}

func VlanToString(v *netlink.Vlan) string {
	return fmt.Sprintf("link: (%s), id: %d, proto: %s", LinkAttrsToString(&v.LinkAttrs), v.VlanId, v.VlanProtocol)
}

func LinkAttrsToString(l *netlink.LinkAttrs) string {
	return fmt.Sprintf("Index: %d, ParentIndex: %d, name: %s, mtu: %d, txqlen: %d", l.Index, l.ParentIndex, l.Name, l.MTU, l.TxQLen)
}
