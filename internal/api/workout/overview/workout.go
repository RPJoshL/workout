package overview

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/api/components/leaflet"
	"git.rpjosh.de/RPJosh/workout/internal/api/workout/shared"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/database/dbstruct"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

var (
	ErrCityToShort = errors.NewError("#workout.cityNameToShort", 400)
	ErrTime        = errors.NewError("Invalid date format provided: %q", 400)
)

// GetTableData returns workout data for the overview table based
// on the provided search values.
//
// "IncludeDetails" states, whether detailed information should be fetched
// for every workout
func (api *Api) GetTableData(includeDeatails bool, filter *shared.WorkoutFilter) (*TableData, errors.Error) {
	rtc := &TableData{}

	// Validate operator
	if err := filter.ValidateFilterOperator(); err != nil {
		return nil, err
	}

	// Get filtered workouts
	sel := api.R().Db.Struct.QuerySlice(&rtc.WorkoutData)
	sel.Where().Column(models.Workout_UserId, "=", api.R().User.Id).Add()

	// Apply filter values
	sel.Where().Column(models.Workout_TypeId, "IN", filter.Activities).IfNotZero()
	sel.Where().Column(`(SELECT tag_id FROM workout_tags tt WHERE tt.workout_id = workout.id)`, "IN", filter.Tags).IfNotZero()
	sel.Where().Column(models.Workout_Distance, filter.KmOperator, filter.Km*1000).IfNotZero()
	sel.Where().Column(models.Workout_Duration, filter.DurationOperator, filter.Duration*60).IfNotZero()

	// Exclude hidden tags. If the tag is specified explicitly, we always want to show it
	if !filter.ShowHiddenTags {
		placeholders := []any{-1}
		operators := "?"
		for _, tag := range filter.Tags {
			operators += ", ?"
			placeholders = append(placeholders, tag)
		}

		sel.Where().Custom(fmt.Sprintf(
			`(SELECT COUNT(*) FROM workout_tags tt 
				 INNER JOIN tag t ON tt.tag_id = t.id
				 WHERE tt.workout_id = workout.id
				   AND t.exclude_default = 1
				   AND t.id NOT IN (%s)
				) = 0`, operators,
		), placeholders...).Add()
	}

	// Date range
	if filter.DateRange != "" {
		toIndex := strings.Index(filter.DateRange, " to ")

		// Only a single date was selected → search for whole day
		if toIndex == -1 {
			if t, err := time.Parse("02.01.2006", filter.DateRange); err != nil {
				return nil, ErrTime.Sprintf(filter.DateRange)
			} else {
				sel.Where().Column(models.Workout_Start, ">=", t).Add()
				sel.Where().Column(models.Workout_Start, "<=", t.AddDate(0, 0, 1)).Add()
			}
		} else {
			t1, err1 := time.Parse("02.01.2006", filter.DateRange[0:toIndex])
			t2, err2 := time.Parse("02.01.2006", filter.DateRange[strings.LastIndex(filter.DateRange, " ")+1:])
			if err1 != nil || err2 != nil {
				return nil, ErrTime.Sprintf(filter.DateRange)
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

	// Add order by
	sel.OrderBy("", models.Workout_Start, "DESC")

	exclude := []string{"*|workout.workout_details"}
	if includeDeatails {
		exclude = []string{}
	}
	if err := sel.Selector(dbstruct.ColumnSelector{ForeignKeyReference: true, PointedKeyReference: true, PointedKeyReferenceAsync: true, ExcludeColumns: exclude}).Run(); err != nil {
		return nil, err.GetResponse().Log("Failed to query workout", err.GetError(), api)
	}

	var mtx sync.Mutex
	var wg sync.WaitGroup

	// Get tags and types
	wg.Add(2)
	go func() {
		if err := api.R().Db.Struct.QuerySlice(&rtc.Types).Run(); err != nil {
			api.Logger().Error("Failed to query workout types: %s", err)
		}
		wg.Done()
	}()
	go func() {
		if err := api.R().Db.Struct.QuerySlice(&rtc.Tags).Run(); err != nil {
			api.Logger().Error("Failed to query workout tags: %s", err)
		}
		wg.Done()
	}()

	// Get downsampled workout data (if specified)
	if !includeDeatails {
		wg.Wait()
		return rtc, nil
	}
	for i := range rtc.WorkoutData {
		wg.Add(1)
		go func(workout *models.Workout) {
			defer wg.Done()

			// Get tooltip content
			buf := new(bytes.Buffer)
			_ = api.GetWorkoutPopup(workout).Render(context.Background(), buf)

			// Downsample points
			downsampled := api.Shared.DownsamplePoints(workout, 20, 2000)
			line := leaflet.Line{
				TooltipContent: buf.String(),
			}
			for _, d := range downsampled {
				line.Points = append(line.Points, leaflet.Point{
					Latitude:  d.Latitude,
					Longitude: d.Longitude,
				})
			}

			mtx.Lock()
			rtc.Lines = append(rtc.Lines, line)
			mtx.Unlock()
		}(&rtc.WorkoutData[i])
	}
	wg.Wait()

	return rtc, nil
}
