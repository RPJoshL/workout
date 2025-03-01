package section

import (
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
	"github.com/a-h/templ"
)

const BUTTON_KEY_ATTRIBUTE = "data-section-button-key"

// Section renders a collapsible box with a tile and
// various action. You have to provide a content
// to display inside the section
type Section struct{}

type Options struct {
	// Title to render on the top of the collapsible box
	Title string

	// Buttons to display next to the expand button
	Buttons []Button

	// Internal ID to identify this section
	id string
}

func (o *Options) getId() string {
	if o.id == "" {
		o.id, _ = utils.GenerateRandomString(12)
		o.id = "o" + o.id
	}

	return o.id
}

type Button struct {
	// Unique key of this button. This property will be set as a value
	// for the HTML attribute BUTTON_KEY_ATTRIBUTE
	Key string

	// Element to display as a content of the button
	Element templ.Component
}
