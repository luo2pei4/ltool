package dblayer

import (
	"fmt"

	"github.com/luo2pei4/ltool/pkg/dblayer/repo"
)

var supportedDB = map[string]struct{}{
	"sqlite": {},
}

var (
	dbs = map[string]func(string) (dblayer, error){}
	DB  dblayer
)

func regist(name string, createInstance func(string) (dblayer, error)) {
	dbs[name] = createInstance
}

func Init(name, dsn string) error {
	if _, ok := supportedDB[name]; !ok {
		return fmt.Errorf("%s is unsupported database", name)
	}
	f, ok := dbs[name]
	if !ok {
		return fmt.Errorf("%s is not registed", name)
	}
	instance, err := f(dsn)
	if err != nil {
		return err
	}
	DB = instance
	return nil
}

type dblayer interface {
	// table nodes operations
	// FindNode
	FindNode(ip string) (*repo.Node, error)
	// ListNodes
	ListNodes(ip string) ([]repo.Node, error)
	// AddNodes
	AddNodes(nodes []repo.Node) error
	// UpdateNode
	UpdateNode(n *repo.Node) error
	// DeleteNode
	DeleteNode(ip string) error
}
