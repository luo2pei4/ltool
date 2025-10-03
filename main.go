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
	"github.com/luo2pei4/ltool/ui"
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
		fmt.Printf("initialize database instance failed, %v", err)
		os.Exit(1)
	}

	a := app.NewWithID("lustre.gui.tool")
	topWindow = a.NewWindow("ltool")
	page := container.NewStack()
	setPage := func(p ui.Page) {
		page.Objects = []fyne.CanvasObject{p.View(topWindow)}
		page.Refresh()
	}

	content := container.NewBorder(nil, nil, nil, nil, page)
	split := container.NewHSplit(makeNav(setPage), content)
	split.Offset = 0.25
	topWindow.SetContent(split)

	topWindow.Resize(fyne.NewSize(1024, 768))
	topWindow.ShowAndRun()
}

func makeNav(setPage func(page ui.Page)) fyne.CanvasObject {
	a := fyne.CurrentApp()

	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			return ui.MenuItemsIndex[uid]
		},
		IsBranch: func(uid string) bool {
			children, ok := ui.MenuItemsIndex[uid]
			return ok && len(children) > 0
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Collection Widgets")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			i, ok := ui.MenuItems[uid]
			if !ok {
				fyne.LogError("Missing tutorial panel: "+uid, nil)
				return
			}
			obj.(*widget.Label).SetText(i.Title)
		},
		OnSelected: func(uid string) {
			for id, f := range ui.OnChangedFunc {
				if id == uid {
					continue
				}
				f()
			}
			if i, ok := ui.MenuItems[uid]; ok {
				a.Preferences().SetString(preferenceCurrentPage, uid)
				setPage(i)
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
