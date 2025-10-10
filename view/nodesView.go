package view

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/luo2pei4/ltool/view/state"
)

type NodesUI struct {
	state              *state.NodesState
	records            *widget.List
	ipEntry            *widget.Entry
	userEntry          *widget.Entry
	passEntry          *widget.Entry
	addBtn             *widget.Button
	selectAllBtn       *widget.Button
	unselectAllBtn     *widget.Button
	deleteBtn          *widget.Button
	saveBtn            *widget.Button
	selectedStatsLabel *widget.Label
}

func NewNodesUI() View {
	return &NodesUI{
		state: &state.NodesState{},
	}
}

func (n *NodesUI) CreateView(w fyne.Window) fyne.CanvasObject {

	n.ipEntry = widget.NewEntry()
	n.ipEntry.SetPlaceHolder("ip adress")
	n.userEntry = widget.NewEntry()
	n.userEntry.SetPlaceHolder("user name")
	n.passEntry = widget.NewPasswordEntry()
	n.passEntry.SetPlaceHolder("user password")
	n.passEntry.Password = false
	n.addBtn = widget.NewButton("+", func() {})
	inputArea := container.NewGridWithColumns(4, n.ipEntry, n.userEntry, n.passEntry, n.addBtn)

	n.selectAllBtn = widget.NewButton("Select All", func() {})
	n.unselectAllBtn = widget.NewButton("Unselect All", func() {})
	n.deleteBtn = widget.NewButton("Delete", func() {})
	n.saveBtn = widget.NewButton("Save", func() {})
	n.selectedStatsLabel = widget.NewLabel("selected stats label")
	btnBar := container.NewBorder(
		nil,
		nil,
		container.NewHBox(n.selectAllBtn, n.unselectAllBtn, n.deleteBtn),
		n.saveBtn,
		container.NewCenter(n.selectedStatsLabel),
	)

	n.records = widget.NewList(
		func() int {
			return len(n.state.Records)
		},
		func() fyne.CanvasObject {
			// UI template for each row
			bg := canvas.NewRectangle(color.Transparent)
			checkbox := widget.NewCheck("", nil)
			ipLabel := widget.NewLabel("")
			userInput := widget.NewEntry()
			passInput := widget.NewPasswordEntry()
			passInput.Password = false
			statuscc := container.NewCenter(canvas.NewText("", color.Black))

			inputArea := container.NewGridWithColumns(4, container.NewStack(bg, ipLabel), userInput, passInput, statuscc)
			row := container.NewBorder(nil, nil, checkbox, nil, inputArea)
			return row
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {

			row := obj.(*fyne.Container)
			checkbox := row.Objects[1].(*widget.Check)
			inputArea := row.Objects[0].(*fyne.Container)

			stack := inputArea.Objects[0].(*fyne.Container)
			bg := stack.Objects[0].(*canvas.Rectangle)
			ipLabel := stack.Objects[1].(*widget.Label)

			userInput := inputArea.Objects[1].(*widget.Entry)
			passInput := inputArea.Objects[2].(*widget.Entry)
			passInput.Password = false

			statuscc := inputArea.Objects[3].(*fyne.Container)
			ctext := statuscc.Objects[0].(*canvas.Text)

			// check
			// checkbox.OnChanged = func(checked bool) {
			// 	n.state.Lock()
			// 	n.state.Records[id].Checked = checked
			// 	n.state.Unlock()
			// 	selectedStatsLabel.SetText(ns.makeSelectedStatsMsg())
			// }
			checkbox.SetChecked(n.state.Records[id].Checked)

			// show ip address
			ipLabel.SetText(n.state.Records[id].IP)

			// modify user
			// userInput.OnChanged = func(user string) {
			// 	ns.Lock()
			// 	init := false
			// 	if ns.records[id].user == user {
			// 		init = true
			// 	}
			// 	ns.records[id].user = user
			// 	if !init {
			// 		ns.records[id].changed = true
			// 	} else {
			// 		ns.records[id].changed = false
			// 	}
			// 	ns.Unlock()
			// 	selectedStatsLabel.SetText(ns.makeSelectedStatsMsg())
			// 	if ns.records[id].changed {
			// 		ns.statusChgCh <- struct{}{}
			// 	}
			// }
			userInput.SetText(n.state.Records[id].User)

			// modify password
			// passInput.OnChanged = func(pass string) {
			// 	ns.Lock()
			// 	init := false
			// 	if ns.records[id].password == pass {
			// 		init = true
			// 	}
			// 	ns.records[id].password = pass
			// 	if !init {
			// 		ns.records[id].changed = true
			// 	} else {
			// 		ns.records[id].changed = false
			// 	}
			// 	ns.Unlock()
			// 	selectedStatsLabel.SetText(ns.makeSelectedStatsMsg())
			// 	if ns.records[id].changed {
			// 		ns.statusChgCh <- struct{}{}
			// 	}
			// }
			passInput.SetText(n.state.Records[id].Password)

			if n.state.Records[id].Status == "online" {
				ctext.Color = color.RGBA{R: 34, G: 177, B: 76, A: 255}
			} else {
				ctext.Color = color.RGBA{R: 235, G: 51, B: 36, A: 255}
			}
			ctext.Text = n.state.Records[id].Status

			if n.state.Records[id].NewRec {
				bg.FillColor = color.RGBA{R: 34, G: 177, B: 76, A: 255} // light green
			} else if n.state.Records[id].Changed {
				bg.FillColor = color.RGBA{R: 50, G: 130, B: 246, A: 255} // light blue
			} else {
				bg.FillColor = color.Transparent
			}
		},
	)

	if err := n.state.LoadAllRecords(); err != nil {
		n.selectedStatsLabel.Text = err.Error()
	} else {
		n.records.Refresh()
	}

	content := container.NewBorder(
		container.NewVBox(
			inputArea,
			widget.NewSeparator(),
		),
		btnBar,    // bottom
		nil,       // left
		nil,       // right
		n.records, // fill content space, node records
	)
	return content
}

func (n *NodesUI) Cleanup() {}
