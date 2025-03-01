// shared contains generic methods for workout processing that can
// be accessed across all submodules without an import cycle
package shared

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/dbutils"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
)

const TYPE_CACHE_DIRECTORY = "./cache/types.json"

type Shared struct {
	router.ApiRequest
}

// Global types that are fetched once at startup
var WorkoutTypes []models.WorkoutType

// GetWorkoutTypeName returns the name of the workout based on the
// users langauge
func (s Shared) GetWorkoutTypeName(typ models.WorkoutType) string {

	// Fallback for unknown type
	if typ.Id == models.TYPE_UNKNOWN {
		return s.R().Tr.Get("workout.unknown")
	}

	switch s.R().Tr.Language {
	case translator.German:
		return typ.NameDe
	default:
		return typ.NameEn
	}
}

func InitializeTypes(db *dbutils.Db, isDevMode bool) {
	// Get workout types from the database once at startup
	if !isDevMode || !loadTypesFromCache() {
		if err := db.Struct.QuerySlice(&WorkoutTypes).Run(); err != nil {
			panic(fmt.Sprintf("Failed to query workout types from db: %s", err))
		} else {
			if content, err := json.Marshal(WorkoutTypes); err == nil {
				if err := os.WriteFile(TYPE_CACHE_DIRECTORY, content, 0644); err != nil {
					logger.Error("Failed to write workout types into cache: %s", err)
				}
			}
		}
	}
}

// loadTypesFromCache tries to load all available types from the local cache
// in order to improve startup times in development mode.
// This function returns whether the cache was consumed
func loadTypesFromCache() bool {
	stats, err := os.Stat(TYPE_CACHE_DIRECTORY)
	if err != nil {
		logger.Debug("Failed to open %q: %s", TYPE_CACHE_DIRECTORY, err)
		return false
	}

	// Only use cached data for maximum 2 days
	if stats.ModTime().After(time.Now().Add(-2 * 24 * time.Hour)) {
		content, err := os.ReadFile(TYPE_CACHE_DIRECTORY)
		if err != nil {
			logger.Debug("Failed to read cached file contents of %q: %s", TYPE_CACHE_DIRECTORY, content)
			return false
		}

		if err := json.Unmarshal(content, &WorkoutTypes); err != nil {
			logger.Warning("Failed to unmarshal types from cache: %s", err)
			return false
		} else {
			logger.Trace("Consumed workout types from local cache")
			return true
		}
	} else {
		logger.Trace("Not using types cache. Modification date is too old")
		return false
	}
}
