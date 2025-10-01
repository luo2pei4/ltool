package node

import (
	"image/color"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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
	newRec   bool
}

func NodeScreen(w fyne.Window) fyne.CanvasObject {
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("ip adress")

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("user name")

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("user password")

	records := []record{}

	// list define
	list := widget.NewList(
		func() int {
			return len(records)
		},
		func() fyne.CanvasObject {
			// UI template for each row
			bg := canvas.NewRectangle(color.Transparent)
			checkbox := widget.NewCheck("", nil)
			ipLabel := widget.NewLabel("")
			userLabel := widget.NewLabel("")
			statusLabel := widget.NewLabel("")
			row := container.NewGridWithColumns(4, checkbox, ipLabel, userLabel, statusLabel)
			return container.NewStack(bg, row)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			rec := records[id]
			// update data
			c := obj.(*fyne.Container)
			bg := c.Objects[0].(*canvas.Rectangle)
			row := c.Objects[1].(*fyne.Container)
			checkbox := row.Objects[0].(*widget.Check)
			ipLabel := row.Objects[1].(*widget.Label)
			userLabel := row.Objects[2].(*widget.Label)
			statusLabel := row.Objects[3].(*widget.Label)

			checkbox.OnChanged = func(checked bool) {
				records[id].checked = checked
			}
			checkbox.SetChecked(records[id].checked)
			ipLabel.SetText(records[id].ip)
			userLabel.SetText(records[id].user)
			statusLabel.SetText(records[id].status)
			if rec.newRec {
				bg.FillColor = color.RGBA{R: 52, G: 135, B: 255, A: 255} // 浅蓝
			} else {
				bg.FillColor = color.Transparent
			}
		},
	)

	selectAllBtn := widget.NewButton("Select All", func() {
		for i := range records {
			records[i].checked = true
		}
		list.Refresh()
	})
	unselectAllBtn := widget.NewButton("Unselect All", func() {
		for i := range records {
			records[i].checked = false
		}
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
					list.Refresh()
				}, w,
			)
		}
	})
	btnBar := container.NewHBox(selectAllBtn, unselectAllBtn, deleteBtn)

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
		list.Refresh()
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
