package main

import (
	"fmt"
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/luo2pei4/ltool/pkg/dblayer"
	"github.com/luo2pei4/ltool/view"
)

type forcedVariant struct {
	fyne.Theme
	variant fyne.ThemeVariant
}

func (f *forcedVariant) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return f.Theme.Color(name, f.variant)
}

const preferenceCurrentPage = "currentPage"

var topWindow fyne.Window

func main() {

	// init database layer
	if err := dblayer.Init("sqlite", "./ltool.db"); err != nil {
		fmt.Printf("initialize database instance failed, %v\n", err)
		os.Exit(1)
	}

	a := app.NewWithID("lustre.gui.tool")
	topWindow = a.NewWindow("ltool")
	page := container.NewStack()
	setContent := func(navi view.Navi) {
		v := navi.Content()
		page.Objects = []fyne.CanvasObject{v.CreateView(topWindow)}
		page.Refresh()
	}

	content := container.NewBorder(nil, nil, nil, nil, page)
	split := container.NewHSplit(makeNav(setContent), content)
	split.Offset = 0.2
	topWindow.SetContent(split)

	topWindow.Resize(fyne.NewSize(1024, 768))
	topWindow.ShowAndRun()
}

func makeNav(setContent func(v view.Navi)) fyne.CanvasObject {
	a := fyne.CurrentApp()

	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			return view.NaviItemsIndex[uid]
		},
		IsBranch: func(uid string) bool {
			children, ok := view.NaviItemsIndex[uid]
			return ok && len(children) > 0
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Collection Widgets")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			i, ok := view.NaviItems[uid]
			if !ok {
				fyne.LogError("Missing tutorial panel: "+uid, nil)
				return
			}
			obj.(*widget.Label).SetText(i.Title)
		},
		OnSelected: func(uid string) {
			if navi, ok := view.NaviItems[uid]; ok {
				a.Preferences().SetString(preferenceCurrentPage, uid)
				setContent(navi)
			}
		},
	}

	themes := container.NewGridWithColumns(2,
		widget.NewButton("Dark", func() {
			a.Settings().SetTheme(&forcedVariant{Theme: theme.DefaultTheme(), variant: theme.VariantDark})
		}),
		widget.NewButton("Light", func() {
			a.Settings().SetTheme(&forcedVariant{Theme: theme.DefaultTheme(), variant: theme.VariantLight})
		}),
	)

	return container.NewBorder(nil, themes, nil, nil, tree)
}
