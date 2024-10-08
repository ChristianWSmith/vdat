package main

import "net/http"

const APP_NAME = "vdat"

const BODY_TYPE_FORM = "FORM"
const BODY_TYPE_RAW = "RAW"
const BODY_TYPE_NONE = "NONE"

var REST_METHODS = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace}

var VALID_RUNES = []rune{
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	' ', '-', '_', '.', '~', '!', '#', '$', '&', '(', ')', '*', '+', ',', '/', ':', ';', '=', '?', '@', '[', ']'}

const HEADERS_PLACEHOLDER = "# comment\nheader1 <tab> value1\nheader2 <tab> value2"
const PARAMS_PLACEHOLDER = "# comment\nparam1=value1\nparam2=value2"
const BODY_CONTENT_PLACEHOLDER_TYPE_NONE = ""
const BODY_CONTENT_PLACEHOLDER_TYPE_FORM = "# comment\nbody1=value1\nbody2=value2"
const BODY_CONTENT_PLACEHOLDER_TYPE_RAW = "{\n    \"body1\": \"value1\",\n    \"body2\": \"value2\"\n}"
const RESPONSE_STATUS_PLACEHOLDER = "<response status>"
const RESPONSE_BODY_PLACEHOLDER = "<response body>"
const RESPONSE_TIME_PLACEHOLDER = "<response time>"
const URL_PLACEHOLDER = "<url>"
const TITLE_PLACEHOLDER = "<title>"

const TABS_PARAMS = "Params"
const TABS_HEADERS = "Headers"
const TABS_BODY = "Body"

const TITLE_DEFAULT = "untitled"

const SSL_ENABLED_TEXT = "SSL"
const SEND_BUTTON_TEXT = "SEND"
const SAVE_BUTTON_TEXT = "SAVE"
const IMPORT_BUTTON_TEXT = "IMPORT FROM CURL"
const NEW_BUTTON_TEXT = "NEW"
const CLOSE_BUTTON_TEXT = "CLOSE"
const OK_BUTTON_TEXT = "OK"
const YES_BUTTON_TEXT = "YES"
const NO_BUTTON_TEXT = "NO"
