package view

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type View interface {
	CreateView(w fyne.Window) fyne.CanvasObject
}

func showProgressing(win fyne.Window, message string, spinnerWidth float32) *widget.PopUp {

	msg := widget.NewLabel(message)
	spinner := widget.NewProgressBarInfinite()
	spinner.Start()

	spinnerLayout := container.NewWithoutLayout(spinner)
	spinner.Resize(fyne.NewSize(spinnerWidth, 12))
	spinner.Move(fyne.NewPos(50, 20))

	card := container.NewVBox(
		spinnerLayout,
		container.New(layout.NewCenterLayout(), msg),
	)

	popup := widget.NewModalPopUp(card, win.Canvas())
	popup.Resize(fyne.NewSize(spinnerWidth+100, 80))
	popup.Show()

	return popup
}
