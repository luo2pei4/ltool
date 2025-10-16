package state

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/luo2pei4/ltool/pkg/dblayer"
	logger "github.com/luo2pei4/ltool/pkg/log"
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
	AltName  string
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
var ipOLinkReg = regexp.MustCompile(`^\d+: (\w+): <([^>]+)> mtu (\d+) .* state (\w+) .* link/(\w+) ([^ ]+) (?:altname (\w+))?`)

// exec: lnetctl net show
func (n *NetInfo) LoadLnetCtlInfo() error {
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
func (n *NetInfo) LoadLinkInfo() error {

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
		if len(matches) < 7 {
			logger.Errorf("failed to parse the input string: %s", line)
			continue
		}

		// Parse MTU to int
		mtu, err := strconv.Atoi(matches[3])
		if err != nil {
			logger.Errorf("failed to parse MTU: %v", err)
			continue
		}

		// Parse flags to slice
		flags := strings.Split(matches[2], ",")
		nif := NetInterface{
			Name:     matches[1],
			Flags:    flags,
			MTU:      mtu,
			State:    matches[4],
			LinkType: matches[5],
			MAC:      matches[6],
		}
		if len(matches) == 8 && matches[7] != "" {
			nif.AltName = matches[7]
		}
		interfaces[matches[1]] = nif
	}
	if len(interfaces) != 0 {
		n.NetInterfacesmap = interfaces
	}
	return nil
}
