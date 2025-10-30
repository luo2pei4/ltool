package view

import (
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	logger "github.com/luo2pei4/ltool/pkg/log"
	"github.com/luo2pei4/ltool/view/layout"
	"github.com/luo2pei4/ltool/view/state"
)

type NetMainUI struct {
	state     *state.NetState
	nodeList  *widget.SelectEntry
	searchBtn *widget.Button
	header    *fyne.Container
	records   *widget.List
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
	v.header = container.New(
		&layout.NetRecordsGrid{},
		widget.NewLabel("Interface"),
		widget.NewLabel("IP Address"),
		widget.NewLabel("Mac"),
		widget.NewLabel("Link Type"),
		widget.NewLabel("State"),
		widget.NewLabel("NID"),
	)
	v.header.Hide()

	v.records = widget.NewList(
		func() int {
			if netInfo, ok := v.state.NodeNet[v.nodeList.Text]; ok {
				return len(netInfo.NetInterfacesMap)
			}
			return 0
		},
		func() fyne.CanvasObject {

			adapterLabel := widget.NewLabel("")
			adapterLabel.Selectable = true

			ipLabel := widget.NewLabel("")
			ipLabel.Selectable = true

			macLabel := widget.NewLabel("")
			macLabel.Selectable = true

			linkTypeLabel := widget.NewLabel("")
			linkTypeLabel.Selectable = true

			stateLabel := widget.NewLabel("")
			stateLabel.Selectable = true

			lnetLabel := widget.NewLabel("")
			lnetLabel.Selectable = true

			recordArea := container.New(
				&layout.NetRecordsGrid{},
				adapterLabel,
				ipLabel,
				macLabel,
				linkTypeLabel,
				stateLabel,
				lnetLabel,
			)
			editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
			return container.NewBorder(nil, nil, nil, editBtn, recordArea)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			row := obj.(*fyne.Container)
			recordArea := row.Objects[0].(*fyne.Container)
			adapterLabel := recordArea.Objects[0].(*widget.Label)
			ipLabel := recordArea.Objects[1].(*widget.Label)
			macLabel := recordArea.Objects[2].(*widget.Label)
			linkTypeLabel := recordArea.Objects[3].(*widget.Label)
			stateLabel := recordArea.Objects[4].(*widget.Label)
			lnetLabel := recordArea.Objects[5].(*widget.Label)
			netInfo, lnetMap := v.state.GetNetInterfaceRecord(v.nodeList.Text, id)
			if netInfo != nil {
				adapterLabel.SetText(netInfo.Name)
				ipLabel.SetText(netInfo.IPv4)
				macLabel.SetText(netInfo.MAC)
				linkTypeLabel.SetText(netInfo.LinkType)
				stateLabel.SetText(netInfo.State)
				if nid, ok := lnetMap[netInfo.Name]; ok {
					lnetLabel.SetText(nid)
				}
			}
			editBtn := row.Objects[1].(*widget.Button)
			editBtn.OnTapped = func() {
				f := dialog.NewForm(
					"Net Config",
					"Save", "Cancel",
					makeNetConfigFormItems(netInfo, lnetMap),
					func(b bool) {},
					w)
				f.Resize(fyne.NewSize(350, 500))
				f.Show()
			}
		},
	)
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
					return
				}
				v.header.Show()
				v.records.Refresh()
			})
		}()
	})
	inputArea := container.NewGridWithColumns(2, v.nodeList, v.searchBtn)
	content := container.NewBorder(
		container.NewVBox(
			inputArea,
			widget.NewSeparator(),
			v.header,
		),
		nil,       // bottom
		nil,       // left
		nil,       // right
		v.records, // fill content space
	)
	return content
}

func makeNetConfigFormItems(netInfo *state.NetInterface, lnetMap map[string]string) []*widget.FormItem {
	items := make([]*widget.FormItem, 0)
	items = append(items, widget.NewFormItem("Interface", widget.NewLabel(netInfo.Name)))
	items = append(items, widget.NewFormItem("Alt names", widget.NewLabel(strings.Join(netInfo.AltNames, ","))))
	items = append(items, widget.NewFormItem("IP address", &widget.Entry{Text: netInfo.IPv4, MultiLine: false}))
	items = append(items, widget.NewFormItem("Mac address", widget.NewLabel(netInfo.MAC)))
	items = append(items, widget.NewFormItem("State", widget.NewLabel(netInfo.State)))
	items = append(items, widget.NewFormItem("Flags", widget.NewLabel(strings.Join(netInfo.Flags, ","))))
	items = append(items, widget.NewFormItem("MTU", widget.NewLabel(strconv.Itoa(netInfo.MTU))))
	nid := lnetMap[netInfo.Name]
	var (
		ip      string
		netType string
		idx     string
	)
	if len(nid) != 0 {
		arr := strings.Split(nid, "@")
		ip = arr[0]
		if strings.HasPrefix(arr[1], "tcp") {
			netType = "tcp"
			idx = strings.TrimPrefix(arr[1], "tcp")
		} else if strings.HasPrefix(arr[1], "o2ib") {
			netType = "o2ib"
			idx = strings.TrimPrefix(arr[1], "o2ib")
		}
	}
	ipEntry := widget.Entry{Text: ip, MultiLine: false}
	idxEntry := widget.Entry{Text: idx, MultiLine: false}
	ntSelect := widget.NewSelectEntry([]string{"tcp", "o2ib"})
	ntSelect.Text = netType
	nidArea := container.New(&layout.NIDAreaGrid{}, &ipEntry, ntSelect, &idxEntry)
	items = append(items, widget.NewFormItem("NID", nidArea))
	return items
}
