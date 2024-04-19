package button

import (
	"github.com/a-h/templ"
)

type Button struct{}

type Options struct {

	// Link to open when the button was pressed
	Href string

	// Open the target in a new tab
	NewTab bool

	// Display an SVG image in front of the text
	Image templ.Component

	// Only display the image without any text
	OnlyImage bool

	// Don't fill the button, but make the text colored
	Outlined bool

	// Action to perform on button click
	OnClick templ.ComponentScript

	// Weather this button is used in a form as "submit"
	IsSubmit bool
}

func (o Options) getHref() templ.SafeURL {
	// Don't do anything on click when no target was provided
	if o.Href == "" {
		return templ.FailedSanitizationURL
	}

	return templ.URL(o.Href)
}

func (o Options) getTarget() string {
	if o.NewTab {
		return "_blank"
	}

	return "_self"
}
