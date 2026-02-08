package overview

import (
	"bytes"
	"context"
	"sync"

	"git.rpjosh.de/RPJosh/workout/internal/api/components/leaflet"
	"git.rpjosh.de/RPJosh/workout/internal/api/workout/shared"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/database/dbstruct"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

// GetTableData returns workout data for the overview table based
// on the provided search values.
//
// "IncludeDetails" states, whether detailed information should be fetched
// for every workout
func (api *Api) GetTableData(includeDeatails bool, filter *shared.WorkoutFilter) (*TableData, errors.Error) {
	rtc := &TableData{
		Filter: filter,
	}

	// Get filtered workouts
	sel := api.R().Db.Struct.QuerySlice(&rtc.WorkoutData)
	sel.Where().Column(models.Workout_UserId, "=", api.R().User.Id).Add()

	// Apply filter values
	if err := shared.ApplyFilter(filter, sel); err != nil {
		return nil, err
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
			downsampled := api.Shared.DownsamplePoints(workout, 20, shared.DownSampleConstraints{
				MaxPointDistance: 2000,
			})
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

// getWorkout returns the workout without any details
func (api *Api) getWorkout(id int) (rtc *models.Workout, err errors.Error) {
	sel := api.R().Db.Struct.Query(&rtc)
	sel.Where().Column(models.Workout_Id, "=", id).Add()
	sel.Where().Column(models.Workout_UserId, "=", api.R().User.Id).Add()

	if err := sel.Run(); err != nil {
		return nil, err.GetResponse().Log("Failed to query workout", err, api)
	}

	return rtc, nil
}
