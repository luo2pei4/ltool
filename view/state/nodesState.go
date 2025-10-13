package state

import (
	"errors"
	"fmt"
	"image/color"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/luo2pei4/ltool/pkg/dblayer"
	"github.com/luo2pei4/ltool/pkg/dblayer/repo"
	logger "github.com/luo2pei4/ltool/pkg/log"
	"github.com/luo2pei4/ltool/pkg/utils"
	"gorm.io/gorm"
)

type Node struct {
	IP       string
	User     string
	rawUser  string
	Password string
	rawPwd   string
	Status   string
	Hostname string
	OS       string
	Arch     string
	Kernel   string
	Checked  bool
	NewRec   bool
	Changed  bool
}

type NodesState struct {
	sync.RWMutex
	Records []Node
}

type hostnamectlResult struct {
	ipAddress       string
	user            string
	password        string
	status          string
	hostname        string
	architecture    string
	operationSystem string
	kernel          string
}

func (n *NodesState) LoadAllRecords() error {
	repoNodes, err := dblayer.DB.ListNodes("")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	n.Lock()
	defer n.Unlock()
	if len(n.Records) == 0 {
		for _, repoNode := range repoNodes {
			n.Records = append(n.Records, Node{
				IP:       repoNode.IPAddress,
				User:     repoNode.UserName,
				rawUser:  repoNode.UserName,
				Password: repoNode.Password,
				rawPwd:   repoNode.Password,
				Status:   "unknown",
				Hostname: repoNode.Hostname,
				Arch:     repoNode.Architecture,
				OS:       repoNode.OS,
				Kernel:   repoNode.Kernel,
			})
		}
	}

	repoNodesMap := make(map[string]repo.Node, len(repoNodes))
	for _, repoNode := range repoNodes {
		repoNodesMap[repoNode.IPAddress] = repoNode
	}
	for i, nod := range n.Records {
		repoNode, ok := repoNodesMap[nod.IP]
		if ok {
			n.Records[i].NewRec = false
			n.Records[i].Changed = false
			n.Records[i].User = repoNode.UserName
			n.Records[i].rawUser = repoNode.UserName
			n.Records[i].Password = repoNode.Password
			n.Records[i].rawPwd = repoNode.Password
			n.Records[i].Hostname = repoNode.Hostname
			n.Records[i].Arch = repoNode.Architecture
			n.Records[i].OS = repoNode.OS
			n.Records[i].Kernel = repoNode.Kernel
		}
	}
	pageNodesMap := make(map[string]Node, len(n.Records))
	for _, nod := range n.Records {
		pageNodesMap[nod.IP] = nod
	}
	for _, repoNode := range repoNodes {
		nod, ok := pageNodesMap[repoNode.IPAddress]
		if !ok {
			nod.Status = "unknown"
			n.Records = append(n.Records, nod)
		}
	}
	return nil
}

func (n *NodesState) MakeStatsMsg() string {
	n.RLock()
	defer n.RUnlock()
	total := len(n.Records)
	selected := 0
	newRecs := 0
	changed := 0
	for _, rec := range n.Records {
		if rec.Checked {
			selected++
		}
		if rec.NewRec {
			newRecs++
		} else if rec.Changed {
			changed++
		}
	}
	return fmt.Sprintf("Total: %d, New: %d, Changed: %d, Checked: %d", total, newRecs, changed, selected)
}

func (n *NodesState) AddNode(ip, user, password string) {

	n.RLock()
	tmpMap := make(map[string]struct{})
	for _, rec := range n.Records {
		tmpMap[rec.IP] = struct{}{}
	}
	n.RUnlock()

	// sort records by ip address
	defer func() {
		sort.SliceStable(n.Records, func(i, j int) bool {
			return n.Records[i].IP < n.Records[j].IP
		})
	}()

	arr := strings.Split(ip, "-")
	if len(arr) == 1 {
		if _, ok := tmpMap[ip]; ok {
			return
		}
		n.Lock()
		defer n.Unlock()
		n.Records = append(n.Records, Node{
			IP:       ip,
			User:     user,
			Password: password,
			Status:   "unknown",
			NewRec:   true,
		})
		return
	}

	toNodeIP, _ := strconv.Atoi(arr[1])
	tmp := strings.Split(arr[0], ".")
	fromNodeIP, _ := strconv.Atoi(tmp[3])
	if toNodeIP == fromNodeIP {
		if _, ok := tmpMap[arr[0]]; ok {
			return
		}
		n.Lock()
		defer n.Unlock()
		n.Records = append(n.Records, Node{
			IP:       arr[0],
			User:     user,
			Password: password,
			Status:   "unknown",
			NewRec:   true,
		})
		return
	}

	// to is less than from, switch
	if toNodeIP < fromNodeIP {
		fromNodeIP, toNodeIP = toNodeIP, fromNodeIP
	}

	n.Lock()
	defer n.Unlock()
	for ; fromNodeIP <= toNodeIP; fromNodeIP++ {
		tmp[3] = strconv.Itoa(fromNodeIP)
		ip := strings.Join(tmp, ".")
		if _, ok := tmpMap[ip]; ok {
			continue
		}
		n.Records = append(n.Records, Node{
			IP:       ip,
			User:     user,
			Password: password,
			Status:   "unknown",
			NewRec:   true,
		})
	}
}

func (n *NodesState) SelectAllRecords() {
	n.Lock()
	defer n.Unlock()
	for i := range n.Records {
		n.Records[i].Checked = true
	}
}

func (n *NodesState) UnselectAllRecords() {
	n.Lock()
	defer n.Unlock()
	for i := range n.Records {
		n.Records[i].Checked = false
	}
}

func (n *NodesState) GetCheckedRecordsCount() int {
	checkedRec := 0
	n.RLock()
	defer n.RUnlock()
	for _, rec := range n.Records {
		if rec.Checked {
			checkedRec++
		}
	}
	return checkedRec
}

func (n *NodesState) DeleteRecords() error {
	if len(n.Records) == 0 {
		return nil
	}
	newRecs := []Node{}
	for _, rec := range n.Records {
		if rec.Checked {
			if !rec.NewRec {
				if err := dblayer.DB.DeleteNode(rec.IP); err != nil {
					return err
				}
			}
			continue
		}
		newRecs = append(newRecs, rec)
	}
	n.Lock()
	defer n.Unlock()
	n.Records = newRecs
	return nil
}

func (n *NodesState) SaveRecords() error {
	newRepos := make([]repo.Node, 0, len(n.Records))
	updRepos := make([]repo.Node, 0, len(n.Records))
	n.Lock()
	defer n.Unlock()
	for _, rec := range n.Records {
		nowaTime := time.Now().Local()
		if rec.NewRec {
			newRepos = append(newRepos, repo.Node{
				IPAddress:    rec.IP,
				UserName:     rec.User,
				Password:     rec.Password,
				Hostname:     rec.Hostname,
				Architecture: rec.Arch,
				OS:           rec.OS,
				Kernel:       rec.Kernel,
				CreateTime:   nowaTime,
				UpdateTime:   nowaTime,
			})
			continue
		}
		if rec.Changed {
			updRepos = append(updRepos, repo.Node{
				IPAddress:    rec.IP,
				UserName:     rec.User,
				Password:     rec.Password,
				Hostname:     rec.Hostname,
				Architecture: rec.Arch,
				OS:           rec.OS,
				Kernel:       rec.Kernel,
				UpdateTime:   nowaTime,
			})
		}
	}
	if len(newRepos) > 0 {
		if err := dblayer.DB.AddNodes(newRepos); err != nil {
			return err
		}
	}
	if len(updRepos) > 0 {
		for _, r := range updRepos {
			if err := dblayer.DB.UpdateNode(&r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *NodesState) GetNodeRecord(id int) Node {
	n.Lock()
	defer n.Unlock()
	return Node{
		IP:       n.Records[id].IP,
		User:     n.Records[id].User,
		Password: n.Records[id].Password,
		Status:   n.Records[id].Status,
		Hostname: n.Records[id].Hostname,
		Arch:     n.Records[id].Arch,
		OS:       n.Records[id].OS,
		Kernel:   n.Records[id].Kernel,
		Checked:  n.Records[id].Checked,
		NewRec:   n.Records[id].NewRec,
		Changed:  n.Records[id].Changed,
	}
}

func (n *NodesState) CheckedRecord(id int, checked bool) {
	n.Lock()
	defer n.Unlock()
	n.Records[id].Checked = checked
}

func (n *NodesState) ChangeUser(id int, user string) {
	n.Lock()
	defer n.Unlock()
	if n.Records[id].User == user {
		n.Records[id].Changed = false
		return
	}
	if n.Records[id].rawUser == user {
		n.Records[id].User = user
		n.Records[id].Changed = false
		return
	}
	n.Records[id].User = user
	n.Records[id].Changed = true
}

func (n *NodesState) ChangePassword(id int, password string) {
	n.Lock()
	defer n.Unlock()
	if n.Records[id].Password == password {
		n.Records[id].Changed = false
		return
	}
	if n.Records[id].rawPwd == password {
		n.Records[id].Password = password
		n.Records[id].Changed = false
		return
	}
	n.Records[id].Password = password
	n.Records[id].Changed = true
}

func (n *NodesState) GetFillColor(id int) color.Color {
	n.RLock()
	defer n.RUnlock()
	if n.Records[id].NewRec {
		return color.RGBA{R: 34, G: 177, B: 76, A: 255} // light green
	} else if n.Records[id].Changed {
		return color.RGBA{R: 50, G: 130, B: 246, A: 255} // light blue
	} else {
		return color.Transparent
	}
}

func (n *NodesState) GetStatusColor(status string) color.Color {
	switch status {
	case "online":
		return color.RGBA{R: 34, G: 177, B: 76, A: 255}
	default:
		return color.RGBA{R: 235, G: 51, B: 36, A: 255}
	}
}

func (n *NodesState) CheckNodesStatus() {
	ipList := make([]hostnamectlResult, 0, len(n.Records))
	n.RLock()
	for _, rec := range n.Records {
		ipList = append(ipList,
			hostnamectlResult{
				ipAddress: rec.IP,
				user:      rec.User,
				password:  rec.Password,
			},
		)
	}
	n.RUnlock()
	n.detectStatus(ipList)
}

func (n *NodesState) detectStatus(ipList []hostnamectlResult) {
	if len(ipList) == 0 {
		return
	}
	resultCh := make(chan *hostnamectlResult, len(ipList))
	defer close(resultCh)

	var wg sync.WaitGroup
	for _, hnc := range ipList {
		wg.Add(1)
		go func(hnc *hostnamectlResult) {
			defer wg.Done()
			if pingable, err := utils.Ping(hnc.ipAddress); err == nil {
				hnc.status = "online"
				if pingable {
					if err := hnc.getHostnamectl(); err != nil {
						logger.Errorf("get hostnamectl command result failed, %v\n", err)
					}
					resultCh <- hnc
				} else {
					hnc.status = "offline"
					resultCh <- hnc
				}
			} else {
				logger.Errorf("ping '%s' error, %v\n", hnc.ipAddress, err)
				hnc.status = "unknown"
				resultCh <- hnc
			}
		}(&hnc)
	}
	wg.Wait()
	cnt := 0
	statusMap := make(map[string]*hostnamectlResult, len(ipList))
	for hnc := range resultCh {
		statusMap[hnc.ipAddress] = hnc
		if cnt++; cnt == len(ipList) {
			break
		}
	}

	n.Lock()
	defer n.Unlock()
	for idx, rec := range n.Records {
		if hnc, ok := statusMap[rec.IP]; ok {
			if n.Records[idx].Status != hnc.status {
				n.Records[idx].Status = hnc.status
			}
			if hnc.hostname != "" && n.Records[idx].Hostname != hnc.hostname {
				n.Records[idx].Hostname = hnc.hostname
				n.Records[idx].Changed = true
			}
			if hnc.architecture != "" && n.Records[idx].Arch != hnc.architecture {
				n.Records[idx].Arch = hnc.architecture
				n.Records[idx].Changed = true
			}
			if hnc.operationSystem != "" && n.Records[idx].OS != hnc.operationSystem {
				n.Records[idx].OS = hnc.operationSystem
				n.Records[idx].Changed = true
			}
			kernel := strings.TrimPrefix(hnc.kernel, "Linux ")
			if hnc.kernel != "" && n.Records[idx].Kernel != kernel {
				n.Records[idx].Kernel = kernel
				n.Records[idx].Changed = true
			}
		}
	}
}

func (hnc *hostnamectlResult) getHostnamectl() error {
	data, err := utils.RemoteCmd(hnc.ipAddress, hnc.user, hnc.password, "hostnamectl")
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	items := strings.Split(string(data), "\n")
	for _, item := range items {
		tmp := strings.TrimSpace(item)
		switch {
		case strings.HasPrefix(tmp, "Static hostname"):
			hnc.hostname = strings.TrimPrefix(tmp, "Static hostname: ")
		case strings.HasPrefix(tmp, "Operating System"):
			hnc.operationSystem = strings.TrimPrefix(tmp, "Operating System: ")
		case strings.HasPrefix(tmp, "Architecture"):
			hnc.architecture = strings.TrimPrefix(tmp, "Architecture: ")
		case strings.HasPrefix(tmp, "Kernel"):
			hnc.kernel = strings.TrimPrefix(tmp, "Kernel: ")
		default:
		}
	}
	return nil
}
