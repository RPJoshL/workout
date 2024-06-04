// shared contains generic methods for workout processing that can
// be accessed across all sub modules without an import cycle
package shared

import (
	"fmt"

	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/database"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
)

type Shared struct {
	router.ApiRequest
}

// Global types that are fetched once at startup
var WorkoutTypes []models.WorkoutType

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

func InitializeTypes(db *database.DatabaseUtils) {
	// Get workout types from the database once at startup
	if err := db.Struct.QuerySlice(&WorkoutTypes).Run(); err != nil {
		panic(fmt.Sprintf("Failed to query workout types from db: %s", err))
	}
}
