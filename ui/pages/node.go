package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Record struct {
	IP       string
	User     string
	Password string
	Checked  bool
}

func NodeScreen(_ fyne.Window) fyne.CanvasObject {
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("IP adress")

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("ssh user name")

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("ssh user password")

	records := []Record{}

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
				records[id].Checked = checked
			}
			checkbox.SetChecked(records[id].Checked)
			ipLabel.SetText(records[id].IP)
			userLabel.SetText(records[id].User)
			passLabel.SetText(records[id].Password)
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

		records = append(records, Record{IP: ip, User: user, Password: pass})
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
