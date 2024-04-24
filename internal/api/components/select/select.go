package selectbox

import (
	"git.rpjosh.de/RPJosh/workout/internal/translator"
	"github.com/a-h/templ"
)

type Settings struct {

	// If the popup sholuld also be visible on hover
	PopupVisibleOnHoover bool

	// Send a request to the application with the selected value and the key
	HtmxOnClick string

	// Function (javascript) to execute when one element was selected
	OnClick func(value string, id string) templ.ComponentScript

	// Function (javascript) to execute after the request completed
	HtmxAfterRequest templ.ComponentScript

	// Index of options to select by default (starting by 1)
	SelectDefault []int

	// Hint to display when no value is selected
	Hint string

	// Allow the selection of multiple options
	MultiOption bool

	// Name of the value used for forms. This has to be unique at least for THIS select instance.
	Name string
}

type Option struct {

	// Display value of the option
	Display string

	// Instead of displaying only a text value, you can also provide
	// a custom component that is rendered inside the selection list and the
	// dropdown.
	// If you are using the class constructs [icon-text.text / icon-text.icon] styling
	// is automatically addedd
	DisplayComponent *templ.Component

	// Raw value for forms / API interactions
	Value string

	// If the element should be hidden by default.
	// You can controle this behavour with the attribute 'data-hidden'
	Hidden bool
}

type SelectBox struct {
	T *translator.Translator
}

func (s Settings) getOnClick(value string, id string) templ.ComponentScript {
	if s.OnClick == nil {
		return templ.ComponentScript{}
	} else {
		return s.OnClick(value, id)
	}
}

func (s Settings) getHint(t *translator.Translator) string {
	if s.Hint == "" {
		return t.Get("c.select.noValue")
	} else {
		return s.Hint
	}
}

func (s Settings) getCheckboxType() string {
	if s.MultiOption {
		return "checkbox"
	} else {
		return "radio"
	}
}

func (s Settings) getSelectType() string {
	if s.MultiOption {
		return "multi"
	} else {
		return "single"
	}
}

func (o Option) getComponent() templ.Component {
	return *o.DisplayComponent
}
