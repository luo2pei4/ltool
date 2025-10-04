package dblayer

import (
	"github.com/luo2pei4/ltool/pkg/consts"
	"github.com/luo2pei4/ltool/pkg/dblayer/repo"
	"gorm.io/gorm"
)

func (s *sqliteLayer) FindNode(ip string) (*repo.Node, error) {
	var node repo.Node
	result := s.Table(consts.TableNodes).Where("ip_address = ?", ip).First(&node)
	if result.Error != nil {
		return nil, result.Error
	}
	return &node, nil
}

func (s *sqliteLayer) ListNodes(ip string) ([]repo.Node, error) {
	var (
		nodes []repo.Node
		err   error
	)
	if ip == "" {
		err = s.Table(consts.TableNodes).Find(&nodes).Error
	} else {
		condition := "ip_address like " + "%" + ip + "%"
		err = s.Table(consts.TableNodes).Find(&nodes, condition).Error
	}
	return nodes, err
}

func (s *sqliteLayer) AddNodes(nodes []repo.Node) error {
	return s.Table(consts.TableNodes).Transaction(func(tx *gorm.DB) error {
		for _, node := range nodes {
			if err := tx.Create(&node).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *sqliteLayer) UpdateNode(n *repo.Node) error {
	return s.Table(consts.TableNodes).Where("ip_address = ?", n.IPAddress).Updates(n).Error
}

func (s *sqliteLayer) DeleteNode(ip string) error {
	return s.Table(consts.TableNodes).Delete(&repo.Node{}, "ip_address = ?", ip).Error
}
