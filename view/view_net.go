package view

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
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
	n.searchBtn = widget.NewButtonWithIcon("", theme.SearchIcon(), func() {})
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
