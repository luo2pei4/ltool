package state

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/luo2pei4/ltool/pkg/dblayer"
	logger "github.com/luo2pei4/ltool/pkg/log"
	"github.com/luo2pei4/ltool/pkg/utils"
	"golang.org/x/crypto/ssh"
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
	Mask     int
	Gateway  string
}

type NetDetail struct {
	Name      string
	IfAlias   string
	Flags     string
	State     string
	MAC       string
	MTU       int
	LinkType  string
	AltNames  string
	IPv4      string
	IPv6      string
	NID       string
	NIDIP     string
	NetType   string
	SuffixIdx string
	Mask      int
	Gateway   string
}

type NetInfo struct {
	Conn             SSHConnection
	LnetCtl          LnetCtl
	NetInterfacesMap map[string]NetInterface // key: interface name
}

type NetState struct {
	sync.RWMutex
	NodeList []string
	SSHCon   map[string]SSHConnection
	Details  []NetDetail
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
	n.SSHCon = make(map[string]SSHConnection, len(repoNodes))
	for _, repoNode := range repoNodes {
		n.NodeList = append(n.NodeList, repoNode.IPAddress)
		n.SSHCon[repoNode.IPAddress] = SSHConnection{
			IPAddress: repoNode.IPAddress,
			User:      repoNode.UserName,
			Password:  repoNode.Password,
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
func loadLnetCtlInfo(ip, user, pwd string) (*LnetCtl, error) {
	data, err := utils.RemoteCmd(ip, user, pwd, "lnetctl net show")
	if err != nil {
		return nil, err
	}
	var lnetctl LnetCtl
	if err := yaml.Unmarshal([]byte(data), &lnetctl); err != nil {
		return nil, err
	}
	return &lnetctl, nil
}

// exec: ip -o link show / ip -o address show
func loadLinkInfo(ip, user, pwd string) (map[string]NetInterface, error) {

	data, err := utils.RemoteCmd(ip, user, pwd, "ip -o link show")
	if err != nil {
		return nil, err
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
	data, err = utils.RemoteCmd(ip, user, pwd, "ip -o address show")
	if err != nil {
		return nil, err
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
		// fields[4]: brd
		// fields[5]: 192.168.50.255

		ifName := fields[1]
		family := fields[2]
		ipWithMask := fields[3]

		// split IP and mask
		arr := strings.Split(ipWithMask, "/")
		ip := arr[0]
		mask := 24
		if len(arr) == 2 {
			mask, _ = strconv.Atoi(arr[1])
		}

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
		iinfo.Mask = mask
		if fields[4] == "brd" {
			iinfo.Gateway = fields[5]
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
	return interfaces, nil
}

func (n *NetState) LoadInterfaceDetail(ip, user, pwd string) error {

	n.RLock()
	defer n.RUnlock()

	ifMap, err := loadLinkInfo(ip, user, pwd)
	if err != nil {
		return err
	}
	details := make([]NetDetail, 0, len(ifMap))
	for _, info := range ifMap {
		details = append(details, NetDetail{
			Name:     info.Name,
			IfAlias:  info.IfAlias,
			Flags:    strings.Join(info.Flags, ","),
			State:    info.State,
			MAC:      info.MAC,
			MTU:      info.MTU,
			LinkType: info.LinkType,
			AltNames: strings.Join(info.AltNames, ","),
			IPv4:     info.IPv4,
			IPv6:     info.IPv6,
			Mask:     info.Mask,
			Gateway:  info.Gateway,
		})
	}
	sort.SliceStable(details, func(i, j int) bool {
		return details[i].Name < details[j].Name
	})
	n.Details = details

	lnetInfo, err := loadLnetCtlInfo(ip, user, pwd)
	if err != nil {
		return err
	}
	if len(lnetInfo.Net) == 0 {
		return nil
	}

	nidMap := make(map[string]string)
	for _, net := range lnetInfo.Net {
		for _, ni := range net.LocalNIs {
			for _, iname := range ni.Interfaces {
				nidMap[iname] = ni.NID
			}
		}
	}
	for i, detail := range details {
		nid, ok := nidMap[detail.Name]
		if !ok {
			continue
		}
		details[i].NID = nid
		arr := strings.Split(nid, "@")
		netType := ""
		idx := ""
		if strings.HasPrefix(arr[1], "tcp") {
			netType = "tcp"
			idx = strings.TrimPrefix(arr[1], "tcp")
		} else if strings.HasPrefix(arr[1], "o2ib") {
			netType = "o2ib"
			idx = strings.TrimPrefix(arr[1], "o2ib")
		}
		details[i].NIDIP = arr[0]
		details[i].NetType = netType
		details[i].SuffixIdx = idx
	}
	n.Details = details
	return nil
}

func (n *NetDetail) SetIPv4(ip, user, pwd string) error {
	// check command exist
	if _, err := utils.RemoteCmd(ip, user, pwd, "nmcli"); err != nil {
		logger.Errorf("check cmd 'nmcli' error, %v", err)
		return errors.New("unable to complete the operation, check whether the 'nmcli' command is installed")
	}
	// check interface exist
	if _, err := utils.RemoteCmd(ip, user, pwd, "nmcli device show "+n.Name); err != nil {
		logger.Errorf("find iface '%s' error, %v", n.Name, err)
		return fmt.Errorf("find iface '%s' error: %v", n.Name, err)
	}
	cmd := utils.AssembleCmd("nmcli", "con", "show", n.Name)
	if _, err := utils.RemoteCmd(ip, user, pwd, cmd); err != nil {
		logger.Errorf("show iface error, %v", err)
		if exitErr, ok := err.(*ssh.ExitError); ok {
			if exitErr.ExitStatus() != 10 {
				return err
			}
		}
		cmd = utils.AssembleCmd("nmcli", "con", "add", "type", "ethernet", "ifname", n.Name, "con-name", n.Name)
		if _, err := utils.RemoteCmd(ip, user, pwd, cmd); err != nil {
			logger.Errorf("set ipv4 address error, %v", err)
			return err
		}
	}
	if len(n.Gateway) != 0 {
		cmd = utils.AssembleCmd(
			"nmcli", "con", "mod", n.Name, "ipv4.method", "manual",
			"ipv4.addr", fmt.Sprintf("%s/%d", n.IPv4, n.Mask), "ipv4.gateway", n.Gateway,
		)
	} else {
		cmd = utils.AssembleCmd(
			"nmcli", "con", "mod", n.Name, "ipv4.method", "manual",
			"ipv4.addr", fmt.Sprintf("%s/%d", n.IPv4, n.Mask))
	}
	if _, err := utils.RemoteCmd(ip, user, pwd, cmd); err != nil {
		logger.Errorf("modify gateway error, %v", err)
		return err
	}
	if _, err := utils.RemoteCmd(ip, user, pwd, "nmcli connection up "+n.Name); err != nil {
		logger.Errorf("up connection error, %v", err)
		return err
	}
	return nil
}
