package view

import "fyne.io/fyne/v2"

type View interface {
	CreateView(w fyne.Window) fyne.CanvasObject
}
