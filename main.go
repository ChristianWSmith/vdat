package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var VALID_HEADER_CHARACTERS = []rune{
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	'-', '_', '.', '~', '!', '#', '$', '&', '(', ')', '*', '+', ',', '/', ':', ';', '=', '?', '@', '[', ']'}

func errorPopUp(canvas fyne.Canvas, err error) {
	modalContent := container.NewVBox(widget.NewLabel(err.Error()))
	popUp := widget.NewModalPopUp(modalContent, canvas)
	okButton := widget.NewButton("Ok", func() { popUp.Hide() })
	modalContent.Add(okButton)
	popUp.Show()
}

func containsRune(slice []rune, element rune) bool {
	for _, item := range slice {
		if item == element {
			return true
		}
	}
	return false
}

func validRunes(value string) bool {
	for _, r := range value {
		if !containsRune(VALID_HEADER_CHARACTERS, r) {
			return false
		}
	}
	return true
}

func newTabContent(canvas fyne.Canvas) fyne.CanvasObject {
	headers := widget.NewMultiLineEntry()
	headers.TextStyle.Monospace = true
	headers.SetPlaceHolder("header1 <tab> value1\nheader2 <tab> value2")
	params := widget.NewMultiLineEntry()
	params.TextStyle.Monospace = true
	params.SetPlaceHolder("param1=value1\nparam2=value2")
	bodyContent := widget.NewMultiLineEntry()
	bodyContent.TextStyle.Monospace = true
	bodyContent.SetPlaceHolder("{\n    \"body\": \"value\"\n}")
	bodyType := widget.NewSelect([]string{"FORM", "MULTIPART FORM", "RAW", "NONE"}, func(value string) {
		if value == "NONE" {
			bodyContent.Disable()
		} else {
			if value == "FORM" {
				bodyContent.SetPlaceHolder("body1=value1\nbody2=value2")
			} else if value == "MULTIPART FORM" {
				// TODO: figure out proper placeholder
				postData :=
					`--xxx
Content-Disposition: form-data; name="field1"

value1
--xxx
Content-Disposition: form-data; name="field2"

value2
--xxx
Content-Disposition: form-data; name="file"; filename="file"
Content-Type: application/octet-stream
Content-Transfer-Encoding: binary

binary data
--xxx--
`
				bodyContent.SetPlaceHolder(postData)
			} else if value == "RAW" {
				bodyContent.SetPlaceHolder("{\n    \"body1\": \"value1\"\n    \"body2\": \"value2\"\n}")
			}
		}
	})
	bodyType.SetSelectedIndex(0)
	bodyPane := container.NewBorder(bodyType, nil, nil, nil, bodyContent)
	responseStatus := widget.NewEntry()
	responseStatus.TextStyle.Monospace = true
	responseStatus.SetPlaceHolder("<response status>")
	responseBody := widget.NewMultiLineEntry()
	responseBody.TextStyle.Monospace = true
	responseBody.SetPlaceHolder("<response body>")
	responseBody.Wrapping = fyne.TextWrapWord

	restMethod := widget.NewSelect([]string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace}, nil)
	restMethod.SetSelectedIndex(0)
	url := widget.NewEntry()
	url.SetPlaceHolder("https://www.website.com/path/to/endpoint")
	sendButton := widget.NewButton("SEND", func() {
		urlText := url.Text
		paramsText := []string{}
		for _, line := range strings.Split(params.Text, "\n") {
			if line == "" {
				continue
			}
			keyValue := strings.Split(line, "=")
			if len(keyValue) == 2 && keyValue[0] != "" && validRunes(keyValue[0]) && validRunes(keyValue[1]) {
				paramsText = append(paramsText, keyValue[0]+"="+keyValue[1])
			} else {
				errorPopUp(canvas, errors.New(fmt.Sprint("Error with param: ", keyValue)))
			}
		}
		if len(paramsText) != 0 {
			urlText = urlText + "?" + strings.Join(paramsText, "&")
		}

		var body io.Reader
		if bodyType.Selected == "NONE" {
			body = strings.NewReader(string(""))
		} else if bodyType.Selected == "FORM" {
			if len(bodyContent.Text) != 0 {
				body = strings.NewReader(strings.Join(strings.Split(bodyContent.Text, "\n"), "&"))
			}
		} else if bodyType.Selected == "MULTIPART FORM" {
			body = strings.NewReader(bodyContent.Text)
		} else if bodyType.Selected == "RAW" {
			body = strings.NewReader(bodyContent.Text)
		}
		req, err := http.NewRequest(restMethod.Selected, urlText, body)
		if err != nil {
			errorPopUp(canvas, err)
			return
		}
		if bodyType.Selected == "FORM" {
			err = req.ParseForm()
			if err != nil {
				errorPopUp(canvas, err)
				return
			}
		} else if bodyType.Selected == "MULTIPART FORM" {
			err = req.ParseMultipartForm(math.MaxInt64)
			if err != nil {
				errorPopUp(canvas, err)
				return
			}
		}

		for _, line := range strings.Split(headers.Text, "\n") {
			if line == "" {
				continue
			}
			keyValue := strings.Split(line, "\t")
			if len(keyValue) == 2 && keyValue[0] != "" && validRunes(keyValue[0]) && validRunes(keyValue[1]) {
				req.Header.Set(keyValue[0], keyValue[1])
			} else {
				errorPopUp(canvas, errors.New(fmt.Sprint("Error with header: ", keyValue)))
			}
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			errorPopUp(canvas, err)
			return
		}
		defer resp.Body.Close()

		// Read and print the response body
		responseBodyContent, err := io.ReadAll(resp.Body)
		if err != nil {
			errorPopUp(canvas, err)
			return
		}
		responseStatus.SetText(resp.Status)
		responseBody.SetText(string(responseBodyContent))
	})
	controls := container.NewBorder(nil, nil, restMethod, sendButton, url)

	requestPane := container.NewAppTabs(
		container.NewTabItem("Params", params),
		container.NewTabItem("Headers", headers),
		container.NewTabItem("Body", bodyPane))
	responsePane := container.NewBorder(responseStatus, nil, nil, nil, responseBody)
	requestAndResponse := container.NewHSplit(requestPane, responsePane)

	content := container.NewBorder(controls, nil, nil, nil, requestAndResponse)

	return content
}

func main() {
	vdatApp := app.New()
	vdatWindow := vdatApp.NewWindow("vdat")

	tabs := container.NewAppTabs()

	tabTitle := widget.NewEntry()
	tabTitle.SetPlaceHolder("<request title>")

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
		newTab := container.NewTabItem("untitled", newTabContent(vdatWindow.Canvas()))
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
