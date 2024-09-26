package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("vdat")
	l := widget.NewLabel("Hello World!")
	l.Alignment = fyne.TextAlignCenter

	b := container.NewBorder(nil, nil, nil, nil, l)

	w.SetContent(b)
	w.ShowAndRun()
}
