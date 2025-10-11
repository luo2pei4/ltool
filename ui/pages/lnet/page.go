package lnet

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func LNetScreen(w fyne.Window) fyne.CanvasObject {
	nodeListBox := widget.NewSelectEntry([]string{})
	ipList, err := getNodesList()
	if err != nil {
		fmt.Printf("load node ip failed, %v\n", err)
	} else {
		nodeListBox.SetOptions(ipList)
	}
	searchBtn := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		fmt.Printf("select: %s\n", nodeListBox.Text)
	})
	inputArea := container.NewGridWithColumns(2, nodeListBox, searchBtn)
	content := container.NewBorder(
		container.NewVBox(
			inputArea,
			widget.NewSeparator(),
		),
		nil, // bottom
		nil, // left
		nil, // right
		nil, // fill content space
	)
	return content
}
