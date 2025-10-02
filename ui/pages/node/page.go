package node

import (
	"fmt"
	"image/color"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func NodeScreen(w fyne.Window) fyne.CanvasObject {
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("ip adress")

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("user name")

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("user password")
	passEntry.Password = false

	nodes := &nodes{
		records:  []node{},
		ipCh:     make(chan string, 100),
		statusCh: make(chan string, 100),
	}

	selectedStatsLabel := widget.NewLabel(nodes.makeSelectedStatsMsg())

	// list define
	list := widget.NewList(
		func() int {
			return len(nodes.records)
		},
		func() fyne.CanvasObject {
			// UI template for each row
			bg := canvas.NewRectangle(color.Transparent)
			checkbox := widget.NewCheck("", nil)
			ipLabel := widget.NewLabel("")
			userInput := widget.NewEntry()
			passInput := widget.NewPasswordEntry()
			passInput.Password = false

			inputArea := container.NewGridWithColumns(3, container.NewStack(bg, ipLabel), userInput, passInput)
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

			// check
			checkbox.OnChanged = func(checked bool) {
				nodes.records[id].checked = checked
				selectedStatsLabel.SetText(nodes.makeSelectedStatsMsg())
			}
			checkbox.SetChecked(nodes.records[id].checked)

			// show ip address
			ipLabel.SetText(nodes.records[id].ip)

			// modify user
			userInput.OnChanged = func(user string) {
				nodes.records[id].user = user
				nodes.records[id].changed = true
				selectedStatsLabel.SetText(nodes.makeSelectedStatsMsg())
			}
			userInput.SetText(nodes.records[id].user)

			// modify password
			passInput.OnChanged = func(pass string) {
				nodes.records[id].password = pass
				nodes.records[id].changed = true
				selectedStatsLabel.SetText(nodes.makeSelectedStatsMsg())
			}
			passInput.SetText(nodes.records[id].password)

			if nodes.records[id].newRec {
				bg.FillColor = color.RGBA{R: 34, G: 177, B: 76, A: 255} // light green
			} else if nodes.records[id].changed {
				bg.FillColor = color.RGBA{R: 50, G: 130, B: 246, A: 255} // light blue
			} else {
				bg.FillColor = color.Transparent
			}
		},
	)

	selectAllBtn := widget.NewButton("Select All", func() {
		for i := range nodes.records {
			nodes.records[i].checked = true
		}
		selectedStatsLabel.SetText(nodes.makeSelectedStatsMsg())
		list.Refresh()
	})
	unselectAllBtn := widget.NewButton("Unselect All", func() {
		for i := range nodes.records {
			nodes.records[i].checked = false
		}
		selectedStatsLabel.SetText(nodes.makeSelectedStatsMsg())
		list.Refresh()
	})
	deleteBtn := widget.NewButton("Delete", func() {
		if len(nodes.records) == 0 {
			list.Refresh()
			return
		}
		checkedRec := 0
		for _, rec := range nodes.records {
			if rec.checked {
				checkedRec++
			}
		}
		if checkedRec > 0 {
			dialog.ShowCustomConfirm(
				"Delete confirm",
				"Yes", "No",
				widget.NewLabel("Are you sure you want to delete the selected records?"),
				func(confirm bool) {
					if !confirm {
						return
					}
					newRecs := []node{}
					for _, rec := range nodes.records {
						if !rec.checked {
							newRecs = append(newRecs, rec)
						}
					}
					nodes.records = newRecs
					selectedStatsLabel.SetText(nodes.makeSelectedStatsMsg())
					list.Refresh()
				}, w,
			)
		}
	})
	saveBtn := widget.NewButton("Save", func() {
		for _, rec := range nodes.records {
			fmt.Printf("IP: %s, user: %s, password: %s\n", rec.ip, rec.user, rec.password)
		}
		list.Refresh()
	})
	btnBar := container.NewBorder(
		nil,
		nil,
		container.NewHBox(selectAllBtn, unselectAllBtn, deleteBtn),
		saveBtn,
		container.NewCenter(selectedStatsLabel),
	)

	// add bottun
	addBtn := widget.NewButton("+", func() {

		ip := ipEntry.Text
		user := userEntry.Text
		pass := passEntry.Text

		switch {
		case ip == "":
			w.Canvas().Focus(ipEntry)
			return
		case user == "":
			w.Canvas().Focus(userEntry)
			return
		case pass == "":
			w.Canvas().Focus(passEntry)
			return
		default:
		}

		if err := validateIP(ip); err != nil {
			dialog.ShowCustom("Warning", "Close", widget.NewLabel(err.Error()), w)
			w.Canvas().Focus(ipEntry)
			return
		}

		nodes.addNode(ip, user, pass)
		sort.SliceStable(nodes.records, func(i, j int) bool {
			return nodes.records[i].ip < nodes.records[j].ip
		})
		selectedStatsLabel.SetText(nodes.makeSelectedStatsMsg())
		list.Refresh()
		// show bottom widgets
		updateBtnBarVisibility(btnBar, len(nodes.records))

		// set focus on ip entry
		w.Canvas().Focus(ipEntry)
	})
	inputForm := container.NewGridWithColumns(4, ipEntry, userEntry, passEntry, addBtn)

	// layout
	content := container.NewBorder(
		container.NewVBox(
			inputForm,
			widget.NewSeparator(),
		),
		btnBar, // bottom
		nil,    // left
		nil,    // right
		list,   // fill content space
	)
	updateBtnBarVisibility(btnBar, len(nodes.records))
	return content
}

func updateBtnBarVisibility(btnBar fyne.CanvasObject, count int) {
	if count > 0 {
		btnBar.Show()
	} else {
		btnBar.Hide()
	}
	btnBar.Refresh()
}
