package view

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	logger "github.com/luo2pei4/ltool/pkg/log"
	"github.com/luo2pei4/ltool/pkg/utils"
	"github.com/luo2pei4/ltool/view/state"
)

type NodesUI struct {
	state          *state.NodesState
	records        *widget.List
	ipEntry        *widget.Entry
	userEntry      *widget.Entry
	passEntry      *widget.Entry
	addBtn         *widget.Button
	selectAllBtn   *widget.Button
	unselectAllBtn *widget.Button
	deleteBtn      *widget.Button
	statusBtn      *widget.Button
	saveBtn        *widget.Button
	statsLabel     *widget.Label
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
	n.addBtn = widget.NewButton("+", func() {
		ip := n.ipEntry.Text
		user := n.userEntry.Text
		pass := n.passEntry.Text
		switch {
		case ip == "":
			w.Canvas().Focus(n.ipEntry)
			return
		case user == "":
			w.Canvas().Focus(n.userEntry)
			return
		case pass == "":
			w.Canvas().Focus(n.passEntry)
			return
		default:
		}
		if err := utils.ValidateIPv4(ip); err != nil {
			dialog.ShowCustom("Warning", "Close", widget.NewLabel(err.Error()), w)
			// set focus on ip entry
			w.Canvas().Focus(n.ipEntry)
			return
		}
		// add node
		n.state.AddNode(ip, user, pass)
		// refresh records list
		n.records.Refresh()
		// set focus on ip entry
		w.Canvas().Focus(n.ipEntry)
	})
	inputArea := container.NewGridWithColumns(4, n.ipEntry, n.userEntry, n.passEntry, n.addBtn)

	n.selectAllBtn = widget.NewButton("Select All", func() {
		n.state.SelectAllRecords()
		n.records.Refresh()
		n.updateStatsMsg()
	})
	n.unselectAllBtn = widget.NewButton("Unselect All", func() {
		n.state.UnselectAllRecords()
		n.records.Refresh()
		n.updateStatsMsg()
	})
	n.deleteBtn = widget.NewButton("Delete", func() {
		if cnt := n.state.GetCheckedRecordsCount(); cnt == 0 {
			return
		}
		dialog.ShowCustomConfirm(
			"Delete confirm",
			"Yes", "No",
			widget.NewLabel("Are you sure you want to delete the selected records?"),
			func(confirm bool) {
				if !confirm {
					return
				}
				if err := n.state.DeleteRecords(); err != nil {
					dialog.ShowCustom("Error", "Close", widget.NewLabel(err.Error()), w)
					return
				}
				if err := n.state.LoadAllRecords(); err != nil {
					dialog.ShowCustom("Error", "Close", widget.NewLabel(fmt.Sprintf("reload nodes failed, %v", err)), w)
					return
				}
				n.updateStatsMsg()
				n.records.Refresh()
			}, w,
		)
	})
	n.statusBtn = widget.NewButton("Status", func() {
		popup := showProgressing(w, "Checking, please wait...", 400)
		go func() {
			n.state.CheckNodesStatus()
			fyne.Do(func() {
				if popup != nil {
					popup.Hide()
				}
				n.records.Refresh()
			})
		}()
	})
	n.saveBtn = widget.NewButton("Save", func() {
		if err := n.state.SaveRecords(); err != nil {
			dialog.ShowCustom("Error", "Close", widget.NewLabel(err.Error()), w)
			return
		}
		if err := n.state.LoadAllRecords(); err != nil {
			dialog.ShowCustom("Error", "Close", widget.NewLabel(fmt.Sprintf("reload nodes failed, %v", err)), w)
			return
		}
		n.updateStatsMsg()
		n.records.Refresh()
	})
	n.statsLabel = widget.NewLabel("")
	btnBar := container.NewBorder(
		nil,
		nil,
		container.NewHBox(n.selectAllBtn, n.unselectAllBtn, n.deleteBtn),
		container.NewHBox(n.statusBtn, n.saveBtn),
		container.NewCenter(n.statsLabel),
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
			passInput.Resize(fyne.NewSize(150, 25))
			statuscc := container.NewCenter(canvas.NewText("", color.Black))
			archcc := container.NewCenter(widget.NewLabel(""))
			kernelcc := container.NewCenter(widget.NewLabel(""))
			inputArea := container.NewGridWithColumns(6,
				container.NewStack(bg, ipLabel),
				userInput,
				passInput,
				statuscc,
				archcc,
				kernelcc,
			)
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
			statustext := statuscc.Objects[0].(*canvas.Text)
			archcc := inputArea.Objects[4].(*fyne.Container)
			archLabel := archcc.Objects[0].(*widget.Label)
			kernelcc := inputArea.Objects[5].(*fyne.Container)
			kernelLabel := kernelcc.Objects[0].(*widget.Label)

			node := n.state.GetNodeRecord(id)

			// check
			checkbox.OnChanged = func(checked bool) {
				n.state.CheckedRecord(id, checked)
				n.updateStatsMsg()
			}
			checkbox.SetChecked(node.Checked)

			// display ip address
			ipLabel.SetText(node.IP)
			// set background color
			bg.FillColor = n.state.GetFillColor(id)

			// change user
			userInput.OnChanged = func(user string) {
				n.state.ChangeUser(id, user)
				bg.FillColor = n.state.GetFillColor(id)
				n.updateStatsMsg()
			}
			userInput.SetText(node.User)

			// change password
			passInput.OnChanged = func(pass string) {
				n.state.ChangePassword(id, pass)
				bg.FillColor = n.state.GetFillColor(id)
				n.updateStatsMsg()
			}
			passInput.SetText(node.Password)

			statustext.Text = node.Status
			statustext.Color = n.state.GetStatusColor(node.Status)
			archLabel.SetText(node.Arch)
			kernelLabel.SetText(node.Kernel)
		},
	)

	if err := n.state.LoadAllRecords(); err != nil {
		logger.Errorf("loading all records failed, %v\n", err)
	} else if len(n.state.Records) > 0 {
		n.records.Refresh()
		n.updateStatsMsg()
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

func (n *NodesUI) updateStatsMsg() {
	n.statsLabel.SetText(n.state.MakeStatsMsg())
}
