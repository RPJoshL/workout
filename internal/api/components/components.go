package components

import (
	"git.rpjosh.de/RPJosh/workout/internal/api/components/button"
	icons "git.rpjosh.de/RPJosh/workout/internal/api/components/icon"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
)

// Components is a base struct that embeds all available components for building a
// static side
type Components struct {
	Icons  *icons.Icons
	Button *button.Button
}

func NewComponents(t *translator.Translator) *Components {
	return &Components{
		Icons:  &icons.Icons{},
		Button: &button.Button{},
	}
}
