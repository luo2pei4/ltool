package state

import (
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/luo2pei4/ltool/pkg/utils"
	"gopkg.in/yaml.v3"
)

type LocalNIS struct {
	NID        string         `yaml:"nid"`
	Status     string         `yaml:"status"`
	Interfaces map[int]string `yaml:"interfaces"`
}

type Lnet struct {
	NetType  string     `yaml:"net type"`
	LocalNIs []LocalNIS `yaml:"local NI(s)"`
}

type LnetCtl struct {
	Net []Lnet `yaml:"net"`
}

type NetInterface struct {
	Index    int
	Name     string
	Flags    []string
	State    string
	MAC      string
	MTU      int
	LinkType string
	IPv4     string
	IPv6     string
}

type NetState struct {
	sync.RWMutex
	Conn             *SSHConnection
	LnetCtl          *LnetCtl
	NetInterfacesmap map[string]map[string]NetInterface // key1: ip address from node info, key2: interface name
}

// ipOLinkReg
//
//	$1: interface index (e.g., 1)
//	$2: interface name (e.g., lo)
//	$3: state flags (e.g., BROADCAST,MULTICAST,UP,LOWER_UP)
//	$4: MTU (e.g., 65536)
//	$5: interface state (e.g., UNKNOWN or UP)
//	$6: MAC address (e.g., 00:1a:2b:3c:4d:5e) - optional
//	^(\d+): -> interface index
//	\s+([^:]+): -> interface name
//	\s+<([^>]+)> -> state flags
//	.*?mtu\s+(\d+) -> MTU
//	.*?state\s+([A-Z]+) -> interface State
//	.*?(?:link/ether\s+([0-9a-f:]+)\s+)? -> MAC address
var ipOLinkReg = regexp.MustCompile(`^(\d+):\s+([^:]+):\s+<([^>]+)>.*?mtu\s+(\d+)\s+.*?state\s+([A-Z]+).*?\s+(\S+)(?:\s+([0-9a-f:]+))?\s+`)

// exec: lnetctl net show
func (n *NetState) GetLnetCtlInfo() error {
	data, err := utils.RemoteCmd(n.Conn.IPAddress, n.Conn.User, n.Conn.Password, "lnetctl net show")
	if err != nil {
		return err
	}
	var lnetctl LnetCtl
	if err := yaml.Unmarshal([]byte(data), &lnetctl); err != nil {
		return err
	}
	n.LnetCtl = &lnetctl
	return nil
}

// exec: ip -o link show
func (n *NetState) GetLinkInfo() error {

	data, err := utils.RemoteCmd(n.Conn.IPAddress, n.Conn.User, n.Conn.Password, "ip -o link show")
	if err != nil {
		return err
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	interfaces := make(map[string]NetInterface, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		matches := ipOLinkReg.FindStringSubmatch(line)
		if len(matches) < 6 {
			continue
		}
		// index
		index, _ := strconv.Atoi(matches[1])
		// name
		name := matches[2]
		// flags
		flagsStr := matches[3]
		flags := strings.Split(flagsStr, ",")
		// mtu
		mtu, _ := strconv.Atoi(matches[4])
		// state
		state := matches[5]
		// link type
		linkType := matches[6]
		// mac address
		macAddr := ""
		if len(matches) > 7 && matches[7] != "" {
			macAddr = matches[7]
		} else if strings.HasPrefix(linkType, "loopback") {
			macAddr = "N/A"
		} else if strings.HasSuffix(linkType, "none") {
			macAddr = "N/A"
		}
		interfaces[name] = NetInterface{
			Index:    index,
			Name:     name,
			Flags:    flags,
			State:    state,
			MAC:      macAddr,
			LinkType: linkType,
			MTU:      mtu,
		}
	}
	return nil
}
