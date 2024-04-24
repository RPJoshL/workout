package components

import (
	"git.rpjosh.de/RPJosh/workout/internal/api/components/button"
	"git.rpjosh.de/RPJosh/workout/internal/api/components/filedrop"
	icons "git.rpjosh.de/RPJosh/workout/internal/api/components/icon"
	"git.rpjosh.de/RPJosh/workout/internal/api/components/markdown"
	selectbox "git.rpjosh.de/RPJosh/workout/internal/api/components/select"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
)

// Components is a base struct that embeds all available components for building a
// static side
type Components struct {
	Select   *selectbox.SelectBox
	Icons    *icons.Icons
	Button   *button.Button
	Markdown *markdown.Markdown
	FileDrop *filedrop.FileDrop
}

func NewComponents(t *translator.Translator) *Components {
	return &Components{
		Icons:    &icons.Icons{},
		Button:   &button.Button{},
		Select:   &selectbox.SelectBox{T: t},
		Markdown: &markdown.Markdown{Icons: &icons.Icons{}},
		FileDrop: &filedrop.FileDrop{Icons: &icons.Icons{}},
	}
}
