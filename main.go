package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	menu "github.com/luo2pei4/ltool/ui"
)

type forcedVariant struct {
	fyne.Theme
	variant fyne.ThemeVariant
}

func (f *forcedVariant) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return f.Theme.Color(name, f.variant)
}

const preferenceCurrentTutorial = "currentTutorial"

var topWindow fyne.Window

func main() {
	a := app.NewWithID("lustre.gui.tool")
	topWindow = a.NewWindow("ltool")
	page := container.NewStack()
	setPage := func(t menu.Page) {
		page.Objects = []fyne.CanvasObject{t.View(topWindow)}
		page.Refresh()
	}

	content := container.NewBorder(nil, nil, nil, nil, page)
	split := container.NewHSplit(makeNav(setPage), content)
	split.Offset = 0.25
	topWindow.SetContent(split)

	topWindow.Resize(fyne.NewSize(1024, 768))
	topWindow.ShowAndRun()
}

func makeNav(setPage func(page menu.Page)) fyne.CanvasObject {
	a := fyne.CurrentApp()

	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			return menu.ItemsIndex[uid]
		},
		IsBranch: func(uid string) bool {
			children, ok := menu.ItemsIndex[uid]
			return ok && len(children) > 0
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Collection Widgets")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			i, ok := menu.Items[uid]
			if !ok {
				fyne.LogError("Missing tutorial panel: "+uid, nil)
				return
			}
			obj.(*widget.Label).SetText(i.Title)
		},
		OnSelected: func(uid string) {
			if i, ok := menu.Items[uid]; ok {
				for _, f := range menu.OnChangeFuncs {
					f()
				}
				menu.OnChangeFuncs = nil // Loading a page registers a new cleanup.
				a.Preferences().SetString(preferenceCurrentTutorial, uid)
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
