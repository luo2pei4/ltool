package view

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	logger "github.com/luo2pei4/ltool/pkg/log"
	"github.com/luo2pei4/ltool/view/state"
)

type NetMainUI struct {
	state     *state.NetState
	nodeList  *widget.SelectEntry
	searchBtn *widget.Button
	// records  *widget.List
}

func NewNetMainUI() View {
	return &NetMainUI{
		state: &state.NetState{},
	}
}

func (n *NetMainUI) CreateView(w fyne.Window) fyne.CanvasObject {
	n.nodeList = widget.NewSelectEntry([]string{})
	if err := n.state.LoadNodeList(); err == nil {
		n.nodeList.SetOptions(n.state.NodeList)
	} else {
		logger.Errorf("load node list failed, %v\n", err)
	}
	n.searchBtn = widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		fmt.Printf("select: %s\n", n.nodeList.Text)
	})
	inputArea := container.NewGridWithColumns(2, n.nodeList, n.searchBtn)
	content := container.NewBorder(
		container.NewVBox(
			inputArea,
			widget.NewSeparator(),
		),
		nil, // bottom
		nil, // left
		nil, // right
		nil, // fill content space
	)
	return content
}
