package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/luo2pei4/ltool/ui/pages"
	"github.com/luo2pei4/ltool/ui/pages/lnet"
)

type Page struct {
	Title string
	Intro string
	View  func(w fyne.Window) fyne.CanvasObject
}

var (
	MenuItems = map[string]Page{
		"lustre": {"Lustre", "", firstScreen},
		"lnet":   {"LNet", "", lnet.LNetScreen},
		"fs":     {"Filesystem", "", pages.FilesystemScreen},
		"mgs":    {"MGS", "", pages.MGSScreen},
		"mds":    {"MDS", "", pages.MDSScreen},
		"oss":    {"OSS", "", pages.OSSScreen},
	}
	MenuItemsIndex = map[string][]string{
		"":       {"lustre"},
		"lustre": {"node", "lnet", "fs", "mgs", "mds", "oss"},
	}
	OnChangedFunc = map[string]func(){}
)

func firstScreen(_ fyne.Window) fyne.CanvasObject {
	return widget.NewLabel("")
}
