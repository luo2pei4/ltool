package state

import (
	"errors"
	"sync"

	"github.com/luo2pei4/ltool/pkg/dblayer"
	"gorm.io/gorm"
)

type Node struct {
	IP       string
	User     string
	Password string
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
				Password: repoNode.Password,
				Status:   "unknown",
			})
		}
	}
	return nil
}
