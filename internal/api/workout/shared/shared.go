// shared contains generic methods for workout processing that can
// be accessed across all sub modules without an import cycle
package shared

import (
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
)

type Shared struct {
	router.ApiRequest
}

// GetWorkoutTypeName returns the name of the workout based on the
// users langauge
func (s Shared) GetWorkoutTypeName(typ models.WorkoutType) string {
	switch s.R().Tr.Language {
	case translator.German:
		return typ.NameDe
	default:
		return typ.NameEn
	}
}
