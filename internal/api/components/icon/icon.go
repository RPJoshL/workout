package icons

import (
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"github.com/a-h/templ"
)

// Icon holds various SVG icons
type Icons struct{}

// GetWorkoutSymbolById returns the SVG icon for the provided
// workout type
func (i *Icons) GetWorkoutSymbolById(id int, class string) templ.Component {
	switch id {
	case models.TYPE_HIKING:
		return i.Hike(class)
	case models.TYPE_RUNNING:
		return i.Running(class)
	}

	return i.Dumbells(class)
}
