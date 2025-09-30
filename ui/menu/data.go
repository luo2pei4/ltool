package memu

import (
	"fyne.io/fyne/v2"
	"github.com/luo2pei4/ltool/ui/pages"
)

var OnChangeFuncs []func()

type Page struct {
	Title string
	Intro string
	View  func(w fyne.Window) fyne.CanvasObject
}

var (
	Items = map[string]Page{
		"lustre": {"Lustre", "", pages.NodeScreen},
		"node":   {"Node", "", pages.NodeScreen},
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
