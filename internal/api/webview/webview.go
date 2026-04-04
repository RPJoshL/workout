package webview

type Action string

const (
	// Changes the theme of the application
	NoAction                 Action = ""
	ActionOpenNativeSettings Action = "openNativeSettings"
	ActionLogout             Action = "logout"
)
