package state

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/luo2pei4/ltool/pkg/dblayer"
	"github.com/luo2pei4/ltool/pkg/utils"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
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

type NetInfo struct {
	Conn             SSHConnection
	LnetCtl          LnetCtl
	NetInterfacesmap map[string]NetInterface // key: interface name
}

type NetState struct {
	sync.RWMutex
	NodeList []string
	NodeNet  map[string]NetInfo
}

func (n *NetState) LoadNodeList() error {
	repoNodes, err := dblayer.DB.ListNodes("")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	n.Lock()
	defer n.Unlock()
	if len(repoNodes) == 0 {
		return nil
	}
	n.NodeList = make([]string, 0, len(repoNodes))
	n.NodeNet = make(map[string]NetInfo, len(repoNodes))
	for _, repoNode := range repoNodes {
		n.NodeList = append(n.NodeList, repoNode.IPAddress)
		n.NodeNet[repoNode.IPAddress] = NetInfo{
			Conn: SSHConnection{
				IPAddress: repoNode.IPAddress,
				User:      repoNode.UserName,
				Password:  repoNode.Password,
			},
		}
	}
	return nil
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
func (n *NetInfo) GetLnetCtlInfo() error {
	data, err := utils.RemoteCmd(n.Conn.IPAddress, n.Conn.User, n.Conn.Password, "lnetctl net show")
	if err != nil {
		return err
	}
	var lnetctl LnetCtl
	if err := yaml.Unmarshal([]byte(data), &lnetctl); err != nil {
		return err
	}
	n.LnetCtl = lnetctl
	return nil
}

// exec: ip -o link show
func (n *NetInfo) GetLinkInfo() error {

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
