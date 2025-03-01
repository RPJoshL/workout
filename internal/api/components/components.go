package components

import (
	"git.rpjosh.de/RPJosh/workout/internal/api/components/button"
	"git.rpjosh.de/RPJosh/workout/internal/api/components/chart"
	"git.rpjosh.de/RPJosh/workout/internal/api/components/filedrop"
	icons "git.rpjosh.de/RPJosh/workout/internal/api/components/icon"
	"git.rpjosh.de/RPJosh/workout/internal/api/components/leaflet"
	"git.rpjosh.de/RPJosh/workout/internal/api/components/markdown"
	"git.rpjosh.de/RPJosh/workout/internal/api/components/section"
	selectbox "git.rpjosh.de/RPJosh/workout/internal/api/components/select"
	"git.rpjosh.de/RPJosh/workout/internal/api/components/table"
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
	Map      *leaflet.Map
	Chart    *chart.Chart
	Section  *section.Section
	Table    *table.Table
}

func NewComponents(t *translator.Translator) *Components {
	return &Components{
		Icons:    &icons.Icons{},
		Button:   &button.Button{},
		Select:   &selectbox.SelectBox{T: t},
		Markdown: &markdown.Markdown{Icons: &icons.Icons{}},
		FileDrop: &filedrop.FileDrop{Icons: &icons.Icons{}},
		Map:      &leaflet.Map{T: t},
		Chart:    &chart.Chart{},
		Section:  &section.Section{},
		Table:    &table.Table{Tr: t, Icons: &icons.Icons{}},
	}
}
