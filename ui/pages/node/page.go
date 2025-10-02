package node

import (
	"fmt"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type record struct {
	ip       string
	user     string
	password string
	status   string
	checked  bool
}

func NodeScreen(w fyne.Window) fyne.CanvasObject {
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("ip adress")

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("user name")

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("user password")
	passEntry.Password = false

	records := []record{}
	selectedStatsLabel := widget.NewLabel(makeSelectedStatsMsg(&records))

	// list define
	list := widget.NewList(
		func() int {
			return len(records)
		},
		func() fyne.CanvasObject {
			// UI template for each row
			checkbox := widget.NewCheck("", nil)
			ipLabel := widget.NewLabel("")
			userInput := widget.NewEntry()
			passInput := widget.NewPasswordEntry()
			passInput.Password = false

			inputArea := container.NewGridWithColumns(3, ipLabel, userInput, passInput)
			row := container.NewBorder(nil, nil, checkbox, nil, inputArea)
			return row
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			row := obj.(*fyne.Container)
			checkbox := row.Objects[1].(*widget.Check)
			inputArea := row.Objects[0].(*fyne.Container)
			ipLabel := inputArea.Objects[0].(*widget.Label)
			userInput := inputArea.Objects[1].(*widget.Entry)
			passInput := inputArea.Objects[2].(*widget.Entry)
			passInput.Password = false

			// check
			checkbox.OnChanged = func(checked bool) {
				records[id].checked = checked
				selectedStatsLabel.SetText(makeSelectedStatsMsg(&records))
			}
			checkbox.SetChecked(records[id].checked)

			// show ip address
			ipLabel.SetText(records[id].ip)

			// modify user
			userInput.OnChanged = func(user string) {
				records[id].user = user
			}
			userInput.SetText(records[id].user)

			// modify password
			passInput.OnChanged = func(pass string) {
				records[id].password = pass
			}
			passInput.SetText(records[id].password)
		},
	)

	selectAllBtn := widget.NewButton("Select All", func() {
		for i := range records {
			records[i].checked = true
		}
		selectedStatsLabel.SetText(makeSelectedStatsMsg(&records))
		list.Refresh()
	})
	unselectAllBtn := widget.NewButton("Unselect All", func() {
		for i := range records {
			records[i].checked = false
		}
		selectedStatsLabel.SetText(makeSelectedStatsMsg(&records))
		list.Refresh()
	})
	deleteBtn := widget.NewButton("Delete", func() {
		if len(records) == 0 {
			list.Refresh()
			return
		}
		checkedRec := 0
		for _, rec := range records {
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
					newRecs := []record{}
					for _, rec := range records {
						if !rec.checked {
							newRecs = append(newRecs, rec)
						}
					}
					records = newRecs
					selectedStatsLabel.SetText(makeSelectedStatsMsg(&records))
					list.Refresh()
				}, w,
			)
		}
	})
	saveBtn := widget.NewButton("Save", func() {
		for _, rec := range records {
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

		addNode(&records, ip, user, pass)
		sort.SliceStable(records, func(i, j int) bool {
			return records[i].ip < records[j].ip
		})
		selectedStatsLabel.SetText(makeSelectedStatsMsg(&records))
		list.Refresh()
		// show bottom widgets
		updateBtnBarVisibility(btnBar, len(records))

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
	updateBtnBarVisibility(btnBar, len(records))
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
