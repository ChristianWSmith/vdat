package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func smartFormat(content []byte) string {
	// TODO: use the following to intelligently format response,
	// possibly also format body that we're sending out whenever we save and/or hit send
	// http.DetectContentType(responseBodyContent)
	return string(content)
}

func errorPopUp(canvas fyne.Canvas, err error) {
	modalContent := container.NewVBox(widget.NewLabel(err.Error()))
	popUp := widget.NewModalPopUp(modalContent, canvas)
	okButton := widget.NewButton(OK_BUTTON_TEXT, func() { popUp.Hide() })
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
		if !containsRune(VALID_RUNES, r) {
			return false
		}
	}
	return true
}

func newTabContent(canvas fyne.Canvas) fyne.CanvasObject {
	headers := widget.NewMultiLineEntry()
	headers.TextStyle.Monospace = true
	headers.SetPlaceHolder(HEADERS_PLACEHOLDER)
	params := widget.NewMultiLineEntry()
	params.TextStyle.Monospace = true
	params.SetPlaceHolder(PARAMS_PLACEHOLDER)
	bodyContent := widget.NewMultiLineEntry()
	bodyContent.TextStyle.Monospace = true
	bodyType := widget.NewSelect([]string{BODY_TYPE_FORM, BODY_TYPE_RAW, BODY_TYPE_NONE}, func(value string) {
		if value == BODY_TYPE_NONE {
			bodyContent.Disable()
			bodyContent.SetPlaceHolder(BODY_CONTENT_PLACEHOLDER_TYPE_NONE)
		} else if value == BODY_TYPE_FORM {
			bodyContent.Enable()
			bodyContent.SetPlaceHolder(BODY_CONTENT_PLACEHOLDER_TYPE_FORM)
		} else if value == BODY_TYPE_RAW {
			bodyContent.Enable()
			bodyContent.SetPlaceHolder(BODY_CONTENT_PLACEHOLDER_TYPE_RAW)
		}

	})
	bodyType.SetSelectedIndex(0)
	bodyPane := container.NewBorder(bodyType, nil, nil, nil, bodyContent)
	responseStatus := widget.NewEntry()
	responseStatus.TextStyle.Monospace = true
	responseStatus.SetPlaceHolder(RESPONSE_STATUS_PLACEHOLDER)
	responseBody := widget.NewMultiLineEntry()
	responseBody.TextStyle.Monospace = true
	responseBody.SetPlaceHolder(RESPONSE_BODY_PLACEHOLDER)
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
	url.SetPlaceHolder(URL_PLACEHOLDER)

	sendButton := widget.NewButton(SEND_BUTTON_TEXT, func() {
		// prepare url with params
		urlText := url.Text
		paramsText := []string{}
		for _, line := range strings.Split(params.Text, "\n") {
			if line == "" || line[0] == '#' {
				continue
			}
			key, value, found := strings.Cut(line, "=")
			if found {
				if key != "" && validRunes(key) && validRunes(value) {
					paramsText = append(paramsText, key+"="+value)
				} else {
					errorPopUp(canvas, errors.New(fmt.Sprint("Error with param entry: ", key, "=", value)))
					return
				}
			}
		}
		if len(paramsText) != 0 {
			urlText = urlText + "?" + strings.Join(paramsText, "&")
		}

		// prepare body
		var body io.Reader
		if bodyType.Selected == BODY_TYPE_NONE {
			body = strings.NewReader(string(""))
		} else if bodyType.Selected == BODY_TYPE_RAW {
			body = strings.NewReader(bodyContent.Text)
		} else if bodyType.Selected == BODY_TYPE_FORM {
			bodyText := []string{}
			for _, line := range strings.Split(bodyContent.Text, "\n") {
				if line == "" || line[0] == '#' {
					continue
				}
				key, value, found := strings.Cut(line, "=")
				if found {
					if key != "" && validRunes(key) && validRunes(value) {
						bodyText = append(paramsText, key+"="+value)
					} else {
						errorPopUp(canvas, errors.New(fmt.Sprint("Error with body entry: ", key, "=", value)))
						return
					}
				}
			}
			finalBodyText := ""
			if len(paramsText) != 0 {
				finalBodyText = strings.Join(bodyText, "&")
			}
			body = strings.NewReader(finalBodyText)

		}

		// create request
		req, err := http.NewRequest(restMethod.Selected, urlText, body)
		if err != nil {
			errorPopUp(canvas, err)
			return
		}

		// parse form if applicable
		if bodyType.Selected == BODY_TYPE_FORM {
			err = req.ParseForm()
			if err != nil {
				errorPopUp(canvas, err)
				return
			}
		}

		// set headers
		for _, line := range strings.Split(headers.Text, "\n") {
			if line == "" || line[0] == '#' {
				continue
			}
			key, value, found := strings.Cut(line, "\t")
			if found {
				if key != "" && validRunes(key) && validRunes(value) {
					req.Header.Set(key, value)
				} else {
					errorPopUp(canvas, errors.New(fmt.Sprint("Error with header entry: ", key, "=", value)))
					return
				}
			}
		}

		// send request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			errorPopUp(canvas, err)
			return
		}
		defer resp.Body.Close()

		// read response
		responseBodyContent, err := io.ReadAll(resp.Body)
		if err != nil {
			errorPopUp(canvas, err)
			return
		}

		// report response
		responseStatus.SetText(resp.Status)
		responseBody.SetText(smartFormat(responseBodyContent))
	})
	controls := container.NewBorder(nil, nil, restMethod, sendButton, url)

	requestPane := container.NewAppTabs(
		container.NewTabItem(TABS_PARAMS, params),
		container.NewTabItem(TABS_HEADERS, headers),
		container.NewTabItem(TABS_BODY, bodyPane))
	responsePane := container.NewBorder(responseStatus, nil, nil, nil, responseBody)
	requestAndResponse := container.NewHSplit(requestPane, responsePane)

	content := container.NewBorder(controls, nil, nil, nil, requestAndResponse)

	return content
}

func main() {
	vdatApp := app.New()
	vdatWindow := vdatApp.NewWindow(APP_NAME)

	tabs := container.NewAppTabs()

	tabTitle := widget.NewEntry()
	tabTitle.SetPlaceHolder(TITLE_PLACEHOLDER)

	tabs.OnSelected = func(ti *container.TabItem) {
		tabTitle.SetText(tabs.Selected().Text)
	}

	tabTitle.OnChanged = func(s string) {
		if tabs.Selected() != nil {
			tabs.Selected().Text = s
			tabs.Refresh()
		}
	}
	saveButton := widget.NewButton(SAVE_BUTTON_TEXT, nil)
	newTabButton := widget.NewButton(NEW_BUTTON_TEXT, func() {
		newTab := container.NewTabItem(TITLE_DEFAULT, newTabContent(vdatWindow.Canvas()))
		tabs.Append(newTab)
		tabs.Select(newTab)
	})
	closeTabButton := widget.NewButton(CLOSE_BUTTON_TEXT, func() {
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
