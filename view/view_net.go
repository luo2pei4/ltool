package view

import (
	"encoding/json"
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

func (v *NetMainUI) CreateView(w fyne.Window) fyne.CanvasObject {
	v.nodeList = widget.NewSelectEntry([]string{})
	if err := v.state.LoadNodeList(); err == nil {
		v.nodeList.SetOptions(v.state.NodeList)
	} else {
		logger.Errorf("load node list failed, %v\n", err)
	}
	v.searchBtn = widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		netInfo, ok := v.state.NodeNet[v.nodeList.Text]
		if !ok {
			netInfo = state.NetInfo{}
		}
		if err := netInfo.LoadLnetCtlInfo(); err != nil {
			logger.Errorf("load lnetctl info failed, %v\n", err)
			return
		}
		if err := netInfo.LoadLinkInfo(); err != nil {
			logger.Errorf("load link info failed, %v\n", err)
			return
		}
		v.state.NodeNet[v.nodeList.Text] = netInfo
		data, err := json.MarshalIndent(&netInfo, "", "  ")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(data))
	})
	inputArea := container.NewGridWithColumns(2, v.nodeList, v.searchBtn)
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
