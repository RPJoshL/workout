package shared

import (
	"fmt"
	"strings"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/database/dbstruct"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

var (
	ErrOperator    = errors.NewError("Invalid comparison operator provided: %q", 400)
	ErrCityToShort = errors.NewError("#workout.cityNameToShort", 400)
	ErrTime        = errors.NewError("Invalid date format provided: %q", 400)
)

// WorkoutFilter contains filter conditions for fetching workouts
type WorkoutFilter struct {
	Activities []int `query:"types"`
	Tags       []int `query:"tags"`

	Km         int    `query:"km"`
	KmOperator string `query:"kmOperator"`

	Duration         int    `query:"duration"`
	DurationOperator string `query:"durationOperator"`

	Radius         int    `query:"radius"`
	RadiusOperator string `query:"radiusOperator"`

	City           int    `query:"city"`
	DateRange      string `query:"dateRange"`
	ShowHiddenTags bool   `query:"showHiddenTags"`
}

// ValidateFilterOperator checks if the provided filter operator
// is valid. It's needed to avoid SQL injections
func (f *WorkoutFilter) ValidateFilterOperator() errors.Error {
	valsTocheck := []string{f.DurationOperator, f.KmOperator, f.RadiusOperator}
	for _, v := range valsTocheck {
		if v != "" && v != "=" && v != ">" && v != ">=" && v != "<" && v != "<=" && v != "<>" {
			return ErrOperator.Sprintf(v)
		}
	}

	return nil
}

func ApplyFilter(filter *WorkoutFilter, sel *dbstruct.Query) errors.Error {
	if err := filter.ValidateFilterOperator(); err != nil {
		return err
	}

	sel.Where().Column(models.Workout_TypeId, "IN", filter.Activities).IfNotZero()
	sel.Where().Column(`(SELECT tag_id FROM workout_tags tt WHERE tt.workout_id = workout.id)`, "IN", filter.Tags).IfNotZero()
	sel.Where().Column(models.Workout_Distance, filter.KmOperator, filter.Km*1000).IfNotZero()
	sel.Where().Column(models.Workout_Duration, filter.DurationOperator, filter.Duration*60).IfNotZero()

	// Exclude hidden tags. If the tag is specified explicitly, we always want to show it
	if !filter.ShowHiddenTags {
		placeholders := []any{-1}
		var operators strings.Builder
		operators.WriteString("?")

		for _, tag := range filter.Tags {
			operators.WriteString(", ?")
			placeholders = append(placeholders, tag)
		}

		sel.Where().Custom(fmt.Sprintf(
			`(SELECT COUNT(*) FROM workout_tags tt 
				 INNER JOIN tag t ON tt.tag_id = t.id
				 WHERE tt.workout_id = workout.id
				   AND t.exclude_default = 1
				   AND t.id NOT IN (%s)
				) = 0`, operators.String(),
		), placeholders...).Add()
	}

	// Date range
	if filter.DateRange != "" {
		toIndex := strings.Index(filter.DateRange, " to ")

		// Only a single date was selected → search for whole day
		if toIndex == -1 {
			if t, err := time.Parse("02.01.2006", filter.DateRange); err != nil {
				return ErrTime.Sprintf(filter.DateRange)
			} else {
				sel.Where().Column(models.Workout_Start, ">=", t).Add()
				sel.Where().Column(models.Workout_Start, "<=", t.AddDate(0, 0, 1)).Add()
			}
		} else {
			t1, err1 := time.Parse("02.01.2006", filter.DateRange[0:toIndex])
			t2, err2 := time.Parse("02.01.2006", filter.DateRange[strings.LastIndex(filter.DateRange, " ")+1:])
			if err1 != nil || err2 != nil {
				return ErrTime.Sprintf(filter.DateRange)
			}

			sel.Where().Column(models.Workout_Start, ">=", t1).Add()
			// Because it's parsed at 00:00, we add a day
			sel.Where().Column(models.Workout_Start, "<=", t2.AddDate(0, 0, 1)).Add()
		}
	}

	// Radius to geonames location
	if filter.Radius > 0 && filter.RadiusOperator != "" && filter.City > 0 {
		sel.CustomJoin(`
			INNER JOIN geonames g ON g.geonameid = ?
		`, filter.City)

		// @TODO use polygon bound to improve query performance.
		// We would need to fetch the cities' location. Is that worth?
		sel.Where().Custom("ST_Distance_Sphere(workout.city_location, g.location) "+filter.RadiusOperator+" ?", filter.Radius*1000).Add()
	}

	return nil
}
