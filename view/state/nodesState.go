package state

import (
	"errors"
	"fmt"
	"image/color"
	"sync"

	"github.com/luo2pei4/ltool/pkg/dblayer"
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
	Records        []Node
	IPsCh          chan []string
	StatusChangeCh chan struct{}
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
