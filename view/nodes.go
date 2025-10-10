package view

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type nodesUI struct {
	list               *widget.List
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
	return &nodesUI{}
}

func (n *nodesUI) CreateView(w fyne.Window) fyne.CanvasObject {

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

	// n.list = widget.NewList(
	// 	func() int { return 0 },
	// 	func() fyne.CanvasObject { return nil },
	// 	func(lii widget.ListItemID, co fyne.CanvasObject) {},
	// )

	content := container.NewBorder(
		container.NewVBox(
			inputArea,
			widget.NewSeparator(),
		),
		btnBar, // bottom
		nil,    // left
		nil,    // right
		nil,    // fill content space, TODO "n.list"
	)
	return content
}

func (n *nodesUI) Cleanup() {}
