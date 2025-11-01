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
	nodeList  *widget.SelectEntry // management ip address list
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
			return len(v.state.Details)
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

			detail := &v.state.Details[id]

			row := obj.(*fyne.Container)
			recordArea := row.Objects[0].(*fyne.Container)
			adapterLabel := recordArea.Objects[0].(*widget.Label)
			ipLabel := recordArea.Objects[1].(*widget.Label)
			macLabel := recordArea.Objects[2].(*widget.Label)
			linkTypeLabel := recordArea.Objects[3].(*widget.Label)
			stateLabel := recordArea.Objects[4].(*widget.Label)
			lnetLabel := recordArea.Objects[5].(*widget.Label)

			adapterLabel.SetText(detail.Name)
			ipLabel.SetText(detail.IPv4)
			macLabel.SetText(detail.MAC)
			linkTypeLabel.SetText(detail.LinkType)
			stateLabel.SetText(detail.State)
			lnetLabel.SetText(detail.NID)

			editBtn := row.Objects[1].(*widget.Button)
			editBtn.OnTapped = func() {
				v.showDetailDialog(w, id)
			}
		},
	)
	v.searchBtn = widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		popup := showProgressing(w, "Searching, please wait...", 400)
		go func() {
			conn := v.state.SSHCon[v.nodeList.Text]
			err := v.state.LoadInterfaceDetail(conn.IPAddress, conn.User, conn.Password)
			fyne.Do(func() {
				if popup != nil {
					popup.Hide()
				}
				if err != nil {
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

func (v *NetMainUI) showDetailDialog(w fyne.Window, id int) {

	managementIP := v.nodeList.Text
	detail := &v.state.Details[id]

	items := make([]*widget.FormItem, 0)
	items = append(items, widget.NewFormItem("Interface", widget.NewLabel(detail.Name)))
	items = append(items, widget.NewFormItem("Alt names", widget.NewLabel(detail.AltNames)))

	ipEntry := &widget.Entry{Text: detail.IPv4, MultiLine: false}
	maskSelect := widget.NewSelectEntry(state.IPv4MaskCIDRList)
	if detail.IPv4 == managementIP {
		items = append(items, widget.NewFormItem("IPv4", widget.NewLabel(managementIP+"/"+strconv.Itoa(detail.Mask))))
	} else {
		maskSelect.Text = strconv.Itoa(detail.Mask)
		ipArea := container.New(&layout.IPAddressAreaGrid{}, ipEntry, maskSelect)
		items = append(items, widget.NewFormItem("IPv4", ipArea))
	}
	gwEntry := &widget.Entry{Text: detail.Gateway, MultiLine: false}
	items = append(items, widget.NewFormItem("Gateway", gwEntry))
	items = append(items, widget.NewFormItem("Mac", widget.NewLabel(detail.MAC)))
	items = append(items, widget.NewFormItem("State", widget.NewLabel(detail.State)))
	items = append(items, widget.NewFormItem("Flags", widget.NewLabel(detail.Flags)))
	items = append(items, widget.NewFormItem("MTU", widget.NewLabel(strconv.Itoa(detail.MTU))))

	nidIPEntry := &widget.Entry{Text: detail.NIDIP, MultiLine: false}
	idxEntry := &widget.Entry{Text: detail.SuffixIdx, MultiLine: false}
	ntSelect := widget.NewSelectEntry([]string{"tcp", "o2ib"})
	ntSelect.Text = detail.NetType
	nidArea := container.New(&layout.NIDAreaGrid{}, nidIPEntry, ntSelect, idxEntry)
	items = append(items, widget.NewFormItem("NID", nidArea))

	f := dialog.NewForm(
		"Net Config",
		"Save", "Cancel",
		items,
		func(ok bool) {
			if ok {
				isManagementInterface := detail.IPv4 == managementIP
				if isManagementInterface {
					return
				}
				if strings.TrimSpace(ipEntry.Text) == "" && detail.IPv4 != "" {
					return
				} else {
					detail.IPv4 = ipEntry.Text
					detail.Mask, _ = strconv.Atoi(maskSelect.Text)
					detail.Gateway = gwEntry.Text
				}

				popup := showProgressing(w, "Saving, please wait...", 400)
				go func() {
					conn := v.state.SSHCon[managementIP]
					err := detail.SetIPv4(conn.IPAddress, conn.User, conn.Password)
					if err == nil {
						err = v.state.LoadInterfaceDetail(conn.IPAddress, conn.User, conn.Password)
					}
					fyne.Do(func() {
						if popup != nil {
							popup.Hide()
						}
						if err != nil {
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
			}
		}, w,
	)
	f.Resize(fyne.NewSize(350, 500))
	f.Show()
}
