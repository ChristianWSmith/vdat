package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	vdatApp := app.New()
	vdatWindow := vdatApp.NewWindow("vdat")

	tabs := container.NewAppTabs()

	tabTitle := widget.NewEntry()

	tabs.OnSelected = func(ti *container.TabItem) {
		tabTitle.SetText(tabs.Selected().Text)
	}

	tabTitle.OnChanged = func(s string) {
		tabs.Selected().Text = s
		tabs.Refresh()
	}

	newTabButton := widget.NewButton("NEW", func() {
		newTab := container.NewTabItem("untitled", widget.NewLabel(""))
		tabs.Append(newTab)
	})
	closeTabButton := widget.NewButton("CLOSE", func() {
		tabs.RemoveIndex(tabs.SelectedIndex())
	})
	tabControlButtons := container.NewHBox(newTabButton, closeTabButton)
	tabControls := container.NewBorder(nil, nil, nil, tabControlButtons, tabTitle)

	tabsWithControls := container.NewBorder(tabControls, nil, nil, nil, tabs)

	vdatWindow.SetContent(tabsWithControls)

	vdatWindow.Resize(fyne.NewSize(800, 450))
	vdatWindow.ShowAndRun()
}
