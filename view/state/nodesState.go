package state

import (
	"errors"
	"fmt"
	"image/color"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/luo2pei4/ltool/pkg/dblayer"
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
	Checked  bool
	NewRec   bool
	Changed  bool
}

type NodesState struct {
	sync.RWMutex
	Records []Node
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
			})
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
	return fmt.Sprintf("total: %d, new: %d, changed: %d, selected: %d", total, newRecs, changed, selected)
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

func (n *NodesState) GetNodeRecord(id int) Node {
	n.Lock()
	defer n.Unlock()
	return Node{
		IP:       n.Records[id].IP,
		User:     n.Records[id].User,
		Password: n.Records[id].Password,
		Status:   n.Records[id].Status,
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
	var ipList []string
	ipList = make([]string, 0, len(n.Records))
	n.RLock()
	for _, rec := range n.Records {
		ipList = append(ipList, rec.IP)
	}
	n.RUnlock()
	n.detectStatus(ipList)
}

func (n *NodesState) detectStatus(ipList []string) {
	if len(ipList) == 0 {
		return
	}
	resultCh := make(chan string, len(ipList))
	defer close(resultCh)

	var wg sync.WaitGroup
	for _, ip := range ipList {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			if pingable, err := utils.Ping(ip); err == nil {
				if pingable {
					resultCh <- ip + "-online"
				} else {
					resultCh <- ip + "-offline"
				}
			} else {
				fmt.Printf("ping error, %v", err)
				resultCh <- ip + "-unknown"
			}
		}(ip)
	}
	wg.Wait()
	cnt := 0
	statusMap := make(map[string]string, len(ipList))
	for res := range resultCh {
		arr := strings.Split(res, "-")
		statusMap[arr[0]] = arr[1]
		if cnt++; cnt == len(ipList) {
			break
		}
	}

	n.Lock()
	defer n.Unlock()
	for idx, rec := range n.Records {
		if status, ok := statusMap[rec.IP]; ok {
			if n.Records[idx].Status != status {
				fmt.Printf("status changed, ip: %s, new status: %s\n", rec.IP, status)
				n.Records[idx].Status = status
			}
		}
	}
}
