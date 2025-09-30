package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	data "github.com/luo2pei4/ltool/ui/menu"
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
	a := app.NewWithID("edmund.luo.ltool")
	topWindow = a.NewWindow("ltool")
	content := container.NewStack()
	setPage := func(t data.Page) {
		content.Objects = []fyne.CanvasObject{t.View(topWindow)}
		content.Refresh()
	}

	feature := container.NewBorder(nil, nil, nil, nil, content)
	split := container.NewHSplit(makeNav(setPage), feature)
	split.Offset = 0.25
	topWindow.SetContent(split)

	topWindow.Resize(fyne.NewSize(1024, 768))
	topWindow.ShowAndRun()
}

func makeNav(setPage func(page data.Page)) fyne.CanvasObject {
	a := fyne.CurrentApp()

	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			return data.ItemsIndex[uid]
		},
		IsBranch: func(uid string) bool {
			children, ok := data.ItemsIndex[uid]
			return ok && len(children) > 0
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Collection Widgets")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			i, ok := data.Items[uid]
			if !ok {
				fyne.LogError("Missing tutorial panel: "+uid, nil)
				return
			}
			obj.(*widget.Label).SetText(i.Title)
		},
		OnSelected: func(uid string) {
			if i, ok := data.Items[uid]; ok {
				for _, f := range data.OnChangeFuncs {
					f()
				}
				data.OnChangeFuncs = nil // Loading a page registers a new cleanup.
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
