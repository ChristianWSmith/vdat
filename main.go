package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func newTabContent() fyne.CanvasObject {
	restMethod := widget.NewSelect([]string{"GET", "PUT", "POST", "DELETE", "PATH", "HEAD", "OPTIONS", "TRACE", "CONNECT"}, nil)
	restMethod.SetSelectedIndex(0)
	url := widget.NewEntry()
	sendButton := widget.NewButton("SEND", nil)
	controls := container.NewBorder(nil, nil, restMethod, sendButton, url)

	headers := widget.NewMultiLineEntry()
	body := widget.NewMultiLineEntry()
	responseCode := widget.NewEntry()
	responseCode.SetText("<response code>")
	responseCode.Disable()
	response := widget.NewMultiLineEntry()
	response.SetText("<response body>")
	response.Disable()

	headersAndBody := container.NewVSplit(headers, body)
	responseCodeAndResponse := container.NewBorder(responseCode, nil, nil, nil, response)
	requestAndResponse := container.NewHSplit(headersAndBody, responseCodeAndResponse)

	content := container.NewBorder(controls, nil, nil, nil, requestAndResponse)

	return content
}

func main() {
	vdatApp := app.New()
	vdatWindow := vdatApp.NewWindow("vdat")

	tabs := container.NewAppTabs()

	tabTitle := widget.NewEntry()

	tabs.OnSelected = func(ti *container.TabItem) {
		tabTitle.SetText(tabs.Selected().Text)
	}

	tabTitle.OnChanged = func(s string) {
		if tabs.Selected() != nil {
			tabs.Selected().Text = s
			tabs.Refresh()
		}
	}
	saveButton := widget.NewButton("SAVE", nil)
	newTabButton := widget.NewButton("NEW", func() {
		newTab := container.NewTabItem("untitled", newTabContent())
		tabs.Append(newTab)
		tabs.Select(newTab)
	})
	closeTabButton := widget.NewButton("CLOSE", func() {
		if len(tabs.Items) >= 2 {
			tabs.RemoveIndex(tabs.SelectedIndex())
		}
	})
	tabControlButtons := container.NewHBox(saveButton, newTabButton, closeTabButton)
	tabControls := container.NewBorder(nil, nil, nil, tabControlButtons, tabTitle)

	tabsWithControls := container.NewBorder(tabControls, nil, nil, nil, tabs)

	fileTree := widget.NewLabel("TODO: File Tree")

	vdatContent := container.NewHSplit(fileTree, tabsWithControls)
	vdatContent.SetOffset(0.25)

	vdatWindow.SetContent(vdatContent)

	vdatWindow.Resize(fyne.NewSize(800, 450))
	newTabButton.OnTapped()
	vdatWindow.ShowAndRun()
}
