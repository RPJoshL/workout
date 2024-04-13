package components

import "git.rpjosh.de/RPJosh/workout/internal/translator"

// Components is a base struct that embeds all available components for building a
// static side
type Components struct {
}

func NewComponents(t *translator.Translator) *Components {
	return &Components{}
}
