package button

import (
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
	"github.com/a-h/templ"
)

type Button struct{}

type Options struct {
	// Random and unique ID of this button
	id string

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
	// Weather this button is used in a form as "reset"
	IsReset bool

	// Action to perform on right click (desktop) / long click (mobile).
	// No parameters are used to call this function
	OnRightClick templ.ComponentScript
}

func (o *Options) getHref() templ.SafeURL {
	// Don't do anything on click when no target was provided
	if o.Href == "" {
		return templ.FailedSanitizationURL
	}

	return templ.URL(o.Href)
}

func (o *Options) getTarget() string {
	if o.NewTab {
		return "_blank"
	}

	return "_self"
}

func (o *Options) getId() string {
	if o.id == "" {
		o.id, _ = utils.GenerateRandomString(12)
		o.id = "button-" + o.id
	}

	return o.id
}
