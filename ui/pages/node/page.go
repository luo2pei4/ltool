package node

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type record struct {
	ip       string
	user     string
	password string
	checked  bool
}

func NodeScreen(_ fyne.Window) fyne.CanvasObject {
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
			checkbox := widget.NewCheck("", nil)
			ipLabel := widget.NewLabel("")
			userLabel := widget.NewLabel("")
			passLabel := widget.NewLabel("")
			return container.NewHBox(checkbox, ipLabel, userLabel, passLabel)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			// update data
			row := obj.(*fyne.Container)
			checkbox := row.Objects[0].(*widget.Check)
			ipLabel := row.Objects[1].(*widget.Label)
			userLabel := row.Objects[2].(*widget.Label)
			passLabel := row.Objects[3].(*widget.Label)

			checkbox.OnChanged = func(checked bool) {
				records[id].checked = checked
			}
			checkbox.SetChecked(records[id].checked)
			ipLabel.SetText(records[id].ip)
			userLabel.SetText(records[id].user)
			passLabel.SetText(records[id].password)
		},
	)

	// add bottun
	addBtn := widget.NewButton("+", func() {
		ip := ipEntry.Text
		user := userEntry.Text
		pass := passEntry.Text

		if ip == "" || user == "" || pass == "" {
			return
		}

		records = append(records, record{ip: ip, user: user, password: pass})
		list.Refresh()

		// clean input
		ipEntry.SetText("")
		userEntry.SetText("")
		passEntry.SetText("")
	})

	inputForm := container.NewGridWithColumns(4, ipEntry, userEntry, passEntry, addBtn)

	// layout
	content := container.NewVBox(
		inputForm,
		widget.NewSeparator(),
		container.NewStack(list),
	)

	return content
}
