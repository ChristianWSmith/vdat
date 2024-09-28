package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/yosssi/gohtml"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func checkDirExists(dir string) (bool, error) {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false, nil // Directory does not exist
	}
	if err != nil {
		return false, err // Some other error occurred
	}
	return info.IsDir(), nil // Return true if it is a directory
}

func getVdatDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var vdatDir string
	var docDir string

	switch runtime.GOOS {
	case "linux":
		docDir = os.Getenv("XDG_DOCUMENTS_DIR")
		if docDir == "" {
			docDir = filepath.Join(home, "Documents")
		}
	default:
		docDir = filepath.Join(home, "Documents")
	}
	vdatDir = filepath.Join(docDir, APP_NAME)
	err = os.MkdirAll(vdatDir, os.ModePerm)
	return vdatDir, err
}

func fileTreeNodeName(id widget.TreeNodeID) string {
	return filepath.Base(id)
}

func fileTreeIsBranch(id widget.TreeNodeID) bool {
	fileInfo, err := os.Stat(id)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func fileTreeLoadChildren(id widget.TreeNodeID) (children []widget.TreeNodeID) {
	files, err := os.ReadDir(id)
	if err != nil {
		return
	}
	for _, file := range files {
		childPath := filepath.Join(id, file.Name())
		children = append(children, childPath)
	}
	return
}

func smartFormat(content []byte) string {

	var rawJson json.RawMessage
	err := json.Unmarshal(content, &rawJson)
	if err == nil {
		var prettyJson []byte
		prettyJson, err = json.MarshalIndent(&rawJson, "", "  ")
		if err != nil {
			return string(content)
		}
		return string(prettyJson)
	}
	return gohtml.Format(string(content))
}

func errorPopUp(canvas fyne.Canvas, err error) {
	modalContent := container.NewVBox(widget.NewLabel(err.Error()))
	popUp := widget.NewModalPopUp(modalContent, canvas)
	okButton := widget.NewButton(OK_BUTTON_TEXT, func() { popUp.Hide() })
	modalContent.Add(okButton)
	popUp.Show()
}

func getStringPopUp(canvas fyne.Canvas, message string) <-chan string {
	entry := widget.NewEntry()
	modalContent := container.NewVBox(widget.NewLabel(message), entry)
	popUp := widget.NewModalPopUp(modalContent, canvas)

	resultCh := make(chan string) // Channel to capture the result

	okButton := widget.NewButton(OK_BUTTON_TEXT, func() {
		resultCh <- entry.Text // Send the entry text to the channel
		popUp.Hide()           // Hide the popup
	})

	modalContent.Add(okButton)
	popUp.Show()

	return resultCh // Return the channel
}

func confirmationPopup(canvas fyne.Canvas, message string) <-chan bool {
	buttons := container.NewHBox()
	modalContent := container.NewBorder(nil, buttons, nil, nil, widget.NewLabel(message))
	popUp := widget.NewModalPopUp(modalContent, canvas)

	resultCh := make(chan bool) // Channel to capture the result

	yesButton := widget.NewButton(YES_BUTTON_TEXT, func() {
		resultCh <- true // Send the entry text to the channel
		popUp.Hide()     // Hide the popup
	})
	noButton := widget.NewButton(NO_BUTTON_TEXT, func() {
		resultCh <- false // Send the entry text to the channel
		popUp.Hide()      // Hide the popup
	})
	buttons.Add(layout.NewSpacer())
	buttons.Add(yesButton)
	buttons.Add(noButton)
	buttons.Add(layout.NewSpacer())

	modalContent.Add(buttons)
	popUp.Show()

	return resultCh // Return the channel
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

type SaveCallback func(string, string) error
type LoadCallback func(string) (string, error)
type VdatRequest struct {
	Headers     string `json:"Headers"`
	Params      string `json:"Params"`
	BodyContent string `json:"BodyContent"`
	BodyType    string `json:"BodyType"`
	Url         string `json:"Url"`
	Title       string `json:"Title"`
	RestMethod  string `json:"RestMethod"`
}

func makeNewTabContent(canvas fyne.Canvas) (fyne.CanvasObject, SaveCallback, LoadCallback) {
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

	saveCallback := func(dirname string, title string) error {
		vdatRequest := VdatRequest{
			Headers:     headers.Text,
			Params:      params.Text,
			BodyContent: bodyContent.Text,
			BodyType:    bodyType.Selected,
			Url:         url.Text,
			Title:       title,
			RestMethod:  restMethod.Selected,
		}

		// Create a file to save the struct
		filename := filepath.Join(dirname, fmt.Sprint(restMethod.Selected, " - ", title))
		file, err := os.Create(filename)
		defer file.Close()
		if err != nil {
			return err
		}

		// Serialize the struct to JSON
		encoder := json.NewEncoder(file)
		err = encoder.Encode(vdatRequest)
		return err
	}

	loadCallback := func(filename string) (string, error) {
		file, err := os.Open(filename)
		if err != nil {
			return "", err
		}
		defer file.Close()

		// Create an instance of the struct to load data into
		vdatRequest := VdatRequest{}

		// Create a JSON decoder and decode the file content into the struct
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&vdatRequest)
		if err != nil {
			return "", err
		}

		headers.SetText(vdatRequest.Headers)
		params.SetText(vdatRequest.Params)
		bodyContent.SetText(vdatRequest.BodyContent)
		bodyType.SetSelected(vdatRequest.BodyType)
		url.SetText(vdatRequest.Url)
		restMethod.SetSelected(vdatRequest.RestMethod)

		return vdatRequest.Title, nil
	}

	return content, saveCallback, loadCallback
}

func main() {
	vdatApp := app.New()
	vdatWindow := vdatApp.NewWindow(APP_NAME)
	tabs := container.NewAppTabs()
	tabTitle := widget.NewEntry()
	tabSaveCallbacks := make(map[*container.TabItem]SaveCallback)

	tree := widget.NewTree(
		func(id widget.TreeNodeID) (children []widget.TreeNodeID) {
			return fileTreeLoadChildren(id)
		},
		func(id widget.TreeNodeID) bool {
			return fileTreeIsBranch(id)
		},
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(fileTreeNodeName(id))
		},
	)

	var err error
	tree.Root, err = getVdatDir()
	tree.Select(tree.Root)
	if err != nil {
		panic("No vdat directory.")
	}

	treeSelected := tree.Root
	selectedFolder := tree.Root

	tree.OnSelected = func(uid widget.TreeNodeID) {
		treeSelected = uid
		isDir, err := checkDirExists(treeSelected)
		if err != nil {
			errorPopUp(vdatWindow.Canvas(), err)
		}
		if isDir {
			selectedFolder = treeSelected
		} else {
			selectedFolder = filepath.Dir(treeSelected)
			newTabContent, saveCallback, loadCallback := makeNewTabContent(vdatWindow.Canvas())
			title, err := loadCallback(uid)
			if err != nil {
				errorPopUp(vdatWindow.Canvas(), errors.New(fmt.Sprint("Failed to load file: ", uid)))
				return
			}
			newTab := container.NewTabItem(title, newTabContent)
			tabSaveCallbacks[newTab] = saveCallback
			tabs.Append(newTab)
			tabs.Select(newTab)
		}
	}

	// Create a scrollable container for the tree
	fileTree := container.NewScroll(tree)

	deleteButton := widget.NewButton("DELETE", func() {
		resultCh := confirmationPopup(vdatWindow.Canvas(), fmt.Sprint("Are you sure you want to delete: ", treeSelected))
		go func() {
			result := <-resultCh
			if result {
				err := os.RemoveAll(treeSelected)
				if err != nil {
					errorPopUp(vdatWindow.Canvas(), errors.New(fmt.Sprint("Failed to delete: ", treeSelected)))
					return
				}
				tree.RefreshItem(selectedFolder)
				tree.Select(filepath.Dir(treeSelected))
			}
		}()
	})

	newFolderButton := widget.NewButton("NEW FOLDER", func() {
		resultCh := getStringPopUp(vdatWindow.Canvas(), "New Folder Name")
		go func() {
			result := <-resultCh
			newFolder := filepath.Join(selectedFolder, result)

			err = os.MkdirAll(newFolder, os.ModePerm)
			if err != nil {
				errorPopUp(vdatWindow.Canvas(), err)
				return
			}
			tree.RefreshItem(selectedFolder)
		}()
	})
	fileControls := container.NewBorder(nil, nil, nil, deleteButton, newFolderButton)
	filePane := container.NewBorder(fileControls, nil, nil, nil, fileTree)

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
	saveButton := widget.NewButton(SAVE_BUTTON_TEXT, func() {
		err := tabSaveCallbacks[tabs.Selected()](selectedFolder, tabTitle.Text)
		if err != nil {
			errorPopUp(vdatWindow.Canvas(), errors.New(fmt.Sprint("Save failed for: ", tabTitle)))
			return
		}
		tree.RefreshItem(selectedFolder)

	})
	newTabButton := widget.NewButton(NEW_BUTTON_TEXT, func() {
		newTabContent, saveCallback, _ := makeNewTabContent(vdatWindow.Canvas())
		newTab := container.NewTabItem(TITLE_DEFAULT, newTabContent)
		tabSaveCallbacks[newTab] = saveCallback
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

	vdatContent := container.NewHSplit(filePane, tabsWithControls)
	vdatContent.SetOffset(0.25)

	vdatWindow.SetContent(vdatContent)

	defer glfw.Terminate()
	monitor := glfw.GetPrimaryMonitor()
	if monitor == nil {
		panic("No monitor.")
	}
	mode := monitor.GetVideoMode()
	if mode == nil {
		panic("No video mode.")
	}

	vdatWindow.Resize(fyne.NewSize(float32(mode.Width*2/3), float32(mode.Height*2/3)))
	vdatWindow.Canvas().Refresh(vdatContent)
	newTabButton.OnTapped()
	vdatWindow.ShowAndRun()
}
