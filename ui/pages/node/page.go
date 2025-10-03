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

	ns := &nodes{
		records:     []node{},
		ipsCh:       make(chan []string, 1),
		statusChgCh: make(chan struct{}, 1),
	}

	// start online/offline status monitor
	go ns.startStatusMonitor()

	selectedStatsLabel := widget.NewLabel(ns.makeSelectedStatsMsg())

	// list define
	list := widget.NewList(
		func() int {
			return len(ns.records)
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
			checkbox.OnChanged = func(checked bool) {
				ns.records[id].checked = checked
				selectedStatsLabel.SetText(ns.makeSelectedStatsMsg())
			}
			checkbox.SetChecked(ns.records[id].checked)

			// show ip address
			ipLabel.SetText(ns.records[id].ip)

			// modify user
			userInput.OnChanged = func(user string) {
				ns.records[id].user = user
				ns.records[id].changed = true
				selectedStatsLabel.SetText(ns.makeSelectedStatsMsg())
			}
			userInput.SetText(ns.records[id].user)

			// modify password
			passInput.OnChanged = func(pass string) {
				ns.records[id].password = pass
				ns.records[id].changed = true
				selectedStatsLabel.SetText(ns.makeSelectedStatsMsg())
			}
			passInput.SetText(ns.records[id].password)

			if ns.records[id].status == "online" {
				ctext.Color = color.RGBA{R: 34, G: 177, B: 76, A: 255}
			} else {
				ctext.Color = color.RGBA{R: 235, G: 51, B: 36, A: 255}
			}
			ctext.Text = ns.records[id].status

			if ns.records[id].newRec {
				bg.FillColor = color.RGBA{R: 34, G: 177, B: 76, A: 255} // light green
			} else if ns.records[id].changed {
				bg.FillColor = color.RGBA{R: 50, G: 130, B: 246, A: 255} // light blue
			} else {
				bg.FillColor = color.Transparent
			}
		},
	)

	go func(n *nodes) {
		fmt.Println("start node status change receiver.")
		for {
			select {
			case <-n.statusChgCh:
				fyne.Do(func() {
					list.Refresh()
				})
			case <-nodePageDoneCh:
				fmt.Println("close node status change receiver.")
				return
			}
		}
	}(ns)

	selectAllBtn := widget.NewButton("Select All", func() {
		for i := range ns.records {
			ns.records[i].checked = true
		}
		selectedStatsLabel.SetText(ns.makeSelectedStatsMsg())
		list.Refresh()
	})
	unselectAllBtn := widget.NewButton("Unselect All", func() {
		for i := range ns.records {
			ns.records[i].checked = false
		}
		selectedStatsLabel.SetText(ns.makeSelectedStatsMsg())
		list.Refresh()
	})
	deleteBtn := widget.NewButton("Delete", func() {
		if len(ns.records) == 0 {
			list.Refresh()
			return
		}
		checkedRec := 0
		for _, rec := range ns.records {
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
					for _, rec := range ns.records {
						if !rec.checked {
							newRecs = append(newRecs, rec)
						}
					}
					ns.records = newRecs
					selectedStatsLabel.SetText(ns.makeSelectedStatsMsg())
					list.Refresh()
				}, w,
			)
		}
	})
	saveBtn := widget.NewButton("Save", func() {
		for _, rec := range ns.records {
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

		ns.addNode(ip, user, pass)
		sort.SliceStable(ns.records, func(i, j int) bool {
			return ns.records[i].ip < ns.records[j].ip
		})
		selectedStatsLabel.SetText(ns.makeSelectedStatsMsg())
		list.Refresh()
		// show bottom widgets
		updateBtnBarVisibility(btnBar, len(ns.records))

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
	updateBtnBarVisibility(btnBar, len(ns.records))
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
