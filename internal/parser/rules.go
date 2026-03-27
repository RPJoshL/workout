package parser

import (
	"slices"

	"git.rpjosh.de/RPJosh/workout/internal/dbutils"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"github.com/RPJoshL/go-logger"
)

// ApplyRules applies all automatation rules to the provided workout.
// Only tags are applied within this function
func ApplyRules(workout *models.Workout, db *dbutils.Db) error {
	if db == nil {
		logger.Debug("No database provided in apply rules")
		return nil
	}

	// Select all rules which do match within the location and duration
	rules := []models.RuleTagging{}
	sel := db.Struct.QuerySlice(&rules)
	sel.Where().Column(models.RuleTagging_UserId, "=", workout.UserId).Add()

	// Duration
	sel.Where().Custom(`rule_tagging.duration_min IS NULL OR ? >= rule_tagging.duration_min`, workout.Duration).Add()
	sel.Where().Custom(`rule_tagging.duration_max IS NULL OR ? <= rule_tagging.duration_max`, workout.Duration).Add()

	// Location
	if len(workout.WorkoutDetails) > 0 {
		startLocation := workout.WorkoutDetails[0]
		endLocation := workout.WorkoutDetails[len(workout.WorkoutDetails)-1]
		sel.CustomJoin(`LEFT JOIN workout.area_circle start ON start.id = rule_tagging.start_location`)
		sel.CustomJoin(`LEFT JOIN workout.area_circle end ON end.id = rule_tagging.end_location`)
		sel.Where().Custom(`
			start.id IS NULL OR ST_Distance_Sphere(start.center, point(?, ?)) <= start.radius
		`, startLocation.Longitude, startLocation.Latitude).Add()
		sel.Where().Custom(`
			end.id IS NULL OR ST_Distance_Sphere(end.center, point(?, ?)) <= end.radius
		`, endLocation.Longitude, endLocation.Latitude).Add()
	}

	if err := sel.Run(); err != nil {
		return err
	}

	// Add all tags if they aren't present already
	for _, rule := range rules {
		doesContain := slices.ContainsFunc(workout.WorkoutTags, func(w models.WorkoutTags) bool {
			return rule.TagId == w.TagId.Id
		})

		if !doesContain {
			workout.WorkoutTags = append(workout.WorkoutTags, models.WorkoutTags{
				TagId: &models.Tag{Id: rule.TagId},
			})
		}
	}

	return nil
}
