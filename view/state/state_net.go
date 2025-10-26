package state

import (
	"errors"
	"regexp"
	"sort"
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
	IfAlias  string
	Flags    []string
	State    string
	MAC      string
	MTU      int
	LinkType string
	AltNames []string
	IPv4     string
	IPv6     string
}

type NetInfo struct {
	Conn             SSHConnection
	LnetCtl          LnetCtl
	NetInterfacesMap map[string]NetInterface // key: interface name
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
// var ipOLinkReg = regexp.MustCompile(`^\d+: (\w+): <([^>]+)> mtu (\d+) .* state (\w+) .* link/(\w+) ([^ ]+) (?:altname (\w+))?`)

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

	reLine := regexp.MustCompile(`^(\d+):\s+([^:]+):\s*(.*)$`)
	reFlags := regexp.MustCompile(`<([^>]*)>`)
	reMTU := regexp.MustCompile(`\bmtu\s+(\d+)`)
	reState := regexp.MustCompile(`\bstate\s+([A-Z]+)`)
	reLink := regexp.MustCompile(`\blink/(\S+)(?:\s+([0-9A-Fa-f:]+))?`)
	reAltName := regexp.MustCompile(`\baltname\s+(\S+)`)

	interfaces := make(map[string]NetInterface, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		m := reLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		index, _ := strconv.Atoi(m[1])
		nameWithAlias := strings.TrimSpace(m[2])
		rest := m[3]

		name := nameWithAlias
		ifAlias := ""
		if at := strings.Index(nameWithAlias, "@"); at >= 0 {
			name = nameWithAlias[:at]
			ifAlias = nameWithAlias[at+1:]
		}
		if name == "lo" {
			continue
		}

		info := NetInterface{
			Index:   index,
			Name:    name,
			IfAlias: ifAlias,
		}

		if fm := reFlags.FindStringSubmatch(rest); fm != nil {
			flags := strings.Split(strings.TrimSpace(fm[1]), ",")
			for i := range flags {
				flags[i] = strings.TrimSpace(flags[i])
			}
			info.Flags = flags
		}

		if mm := reMTU.FindStringSubmatch(rest); mm != nil {
			if mtuVal, err := strconv.Atoi(mm[1]); err == nil {
				info.MTU = mtuVal
			}
		}

		if sm := reState.FindStringSubmatch(rest); sm != nil {
			info.State = sm[1]
		}

		if lm := reLink.FindStringSubmatch(rest); lm != nil {
			info.LinkType = lm[1]
			if lm[2] != "" {
				info.MAC = lm[2]
			}
		}

		// multiple altname
		for _, am := range reAltName.FindAllStringSubmatch(rest, -1) {
			if len(am) > 1 {
				info.AltNames = append(info.AltNames, am[1])
			}
		}

		interfaces[info.Name] = info
	}
	// ip addresses
	data, err = utils.RemoteCmd(n.Conn.IPAddress, n.Conn.User, n.Conn.Password, "ip -o address show")
	if err != nil {
		return err
	}
	lines = strings.Split(strings.TrimSpace(string(data)), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		// fields[0]: serial，example "2:"
		// fields[1]: adapter name，example "eth0"
		// fields[2]: protocal，"inet" or "inet6"
		// fields[3]: IP/mask，exampl "192.168.1.100/24" or "fe80::1/64"

		ifName := fields[1]
		family := fields[2]
		ipWithMask := fields[3]

		// split IP and mask
		ip := strings.Split(ipWithMask, "/")[0]

		// filt loopback and invalid address
		if ifName == "lo" {
			continue
		}
		if ip == "" {
			continue
		}

		iinfo, ok := interfaces[ifName]
		if !ok {
			continue
		}

		if family == "inet" {
			iinfo.IPv4 = ip
		} else if family == "inet6" {
			if strings.HasPrefix(ip, "fe80:") {
				continue
			}
			iinfo.IPv6 = ip
		} else {
			continue
		}
		interfaces[ifName] = iinfo
	}

	if len(interfaces) != 0 {
		n.NetInterfacesMap = interfaces
	}
	return nil
}

func (n *NetState) GetNetInterfaceRecord(nodeIP string, id int) *NetInterface {
	n.RLock()
	defer n.RUnlock()
	nodeNet, ok := n.NodeNet[nodeIP]
	if !ok {
		return nil
	}
	interfaceList := make([]*NetInterface, 0, len(nodeNet.NetInterfacesMap))
	for _, info := range nodeNet.NetInterfacesMap {
		interfaceList = append(interfaceList, &info)
	}
	sort.SliceStable(interfaceList, func(i, j int) bool {
		return interfaceList[i].Name < interfaceList[j].Name
	})
	return interfaceList[id]
}
