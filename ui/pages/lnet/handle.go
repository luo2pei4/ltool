package lnet

import (
	"errors"

	"github.com/luo2pei4/ltool/pkg/dblayer"
	"gorm.io/gorm"
)

func getNodesList() ([]string, error) {
	nodes, err := dblayer.DB.ListNodes("")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	ipList := make([]string, 0, len(nodes))
	for _, node := range nodes {
		ipList = append(ipList, node.IPAddress)
	}
	return ipList, nil
}
