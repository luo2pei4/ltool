package view

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type Navi struct {
	Title   string
	Content func() View
}

type firstScreen struct{}

func createFirstScreen() View {
	return &firstScreen{}
}

var (
	NaviItems = map[string]Navi{
		"lustre": {"Lustre", createFirstScreen},
		"node":   {"Node", NewNodesUI},
	}
	NaviItemsIndex = map[string][]string{
		"":       {"lustre"},
		"lustre": {"node"},
	}
)

func (f *firstScreen) CreateView(w fyne.Window) fyne.CanvasObject {
	return widget.NewLabel("")
}
