package view

import (
	"encoding/json"
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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
		popup := showProgressing(w, "Searching, please wait...", 400)
		go func() {
			netInfo, ok := v.state.NodeNet[v.nodeList.Text]
			if !ok {
				netInfo = state.NetInfo{}
			}
			err := netInfo.LoadLnetCtlInfo()
			if err != nil {
				logger.Errorf("load lnetctl info failed, %v\n", err)
			} else {
				err = netInfo.LoadLinkInfo()
				if err != nil {
					logger.Errorf("load link info failed, %v\n", err)
				}
			}
			v.state.NodeNet[v.nodeList.Text] = netInfo
			if data, err := json.MarshalIndent(&netInfo, "", "  "); err == nil {
				fmt.Println(string(data))
			}
			fyne.Do(func() {
				if popup != nil {
					popup.Hide()
				}
				if ok && err != nil {
					// draw error dialog
					errLabel := widget.NewLabel(err.Error())
					errLabel.Wrapping = fyne.TextWrapWord
					bg := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
					bg.SetMinSize(fyne.NewSize(400, 160))
					content := container.NewStack(bg, container.NewVBox(errLabel))
					dialog.ShowCustom("Error", "Close", content, w)
				}
			})
		}()
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
