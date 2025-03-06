package form

import (
	"git.rpjosh.de/RPJosh/workout/internal/api/components/button"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
	"github.com/a-h/templ"
)

type FieldType int8

const (
	Text FieldType = iota
	Checkbox
	Color
)

type Form struct {
	Tr     *translator.Translator
	Button *button.Button
}

type Options struct {
	// Unique ID of the form within the page
	Id string

	// Url where to send a post request when submitting the form
	Url string

	// Additional attributes to apply to the form
	Attributes templ.Attributes

	// Function that returns the translation key to use for the button
	ButtonLabel func() string

	// Fields to display in a row
	Fields []Field
}

// Field is a single user input that is displayed as a row within the form
type Field struct {

	// Unique name of this field within the form. It will be used
	// as a key when psoting the values
	Name string

	Type FieldType

	// Translation key used to display label above / besides the form
	Label string

	// Render a custom component instead of the provided defaults
	CustomComponent templ.Component

	// Initial value to apply
	Value string

	// Wheather the input field should be displayed as hidden.
	// This is especially usefull when setting the ID of an entity
	// that cannot be changed by the user
	Hidden bool

	// Do not render this field in the DOM. This allows you to
	// dynamically exclude fields without modifying the fields array
	Exclude bool
}
