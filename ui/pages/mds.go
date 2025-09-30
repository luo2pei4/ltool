package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func MDSScreen(_ fyne.Window) fyne.CanvasObject {
	return container.NewStack(widget.NewLabel("mds"))
}
