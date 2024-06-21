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
	case models.TYPE_SURFEN:
		return i.Surfing(class)
	case models.TYPE_SAILING:
		return i.Sailing(class)
	case models.TYPE_SNOWBOARDING:
		return i.Snowboarding(class)
	case models.TYPE_SWIMMING:
		return i.Swimming(class)
	case models.TYPE_CYCLING:
		return i.Cycling(class)
	case models.TYPE_SKATEBOARDING:
		return i.Skateboarding(class)
	case models.TYPE_VOLLEYBALL:
		return i.Volleyball(class)
	}

	return i.Dumbells(class)
}
