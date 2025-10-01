package memu

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/luo2pei4/ltool/ui/pages"
	"github.com/luo2pei4/ltool/ui/pages/node"
)

var OnChangeFuncs []func()

type Page struct {
	Title string
	Intro string
	View  func(w fyne.Window) fyne.CanvasObject
}

var (
	Items = map[string]Page{
		"lustre": {"Lustre", "", firstScreen},
		"node":   {"Node", "", node.NodeScreen},
		"lnet":   {"LNet", "", pages.LNetScreen},
		"fs":     {"Filesystem", "", pages.FilesystemScreen},
		"mgs":    {"MGS", "", pages.MGSScreen},
		"mds":    {"MDS", "", pages.MDSScreen},
		"oss":    {"OSS", "", pages.OSSScreen},
	}
	ItemsIndex = map[string][]string{
		"":       {"lustre"},
		"lustre": {"node", "lnet", "fs", "mgs", "mds", "oss"},
	}
)

func firstScreen(_ fyne.Window) fyne.CanvasObject {
	return widget.NewLabel("")
}
