package workout

import (
	"git.rpjosh.de/RPJosh/go-webserver/errors"
	"git.rpjosh.de/RPJosh/workout/internal/database"
	"git.rpjosh.de/RPJosh/workout/internal/models"
)

var (
	ErrWorkoutNotFound = errors.NewError("#workout.notFound", 404)
)

func (a *Api) GetWorkoutNewEditData(existingWorkout int) (*workoutNewEditData, errors.Error) {
	rtc := &workoutNewEditData{}

	// Query existing workout data
	if existingWorkout > 0 {
		err := a.R().Db.Struct.Query(&rtc.existingWorkout).Selector(database.ColumnSelector{
			// We don't need any workout details in edit dialog
			ExcludeColumns: []string{models.Workout_WorkoutDetails},
		}).Run()

		if err.Type() == database.NoRows {
			return nil, ErrWorkoutNotFound
		} else {
			return nil, err.GetResponse().Log("Failed to query existing workout", err, a)
		}
	}

	// Query available types and tags
	if err := a.R().Db.Struct.QuerySlice(&rtc.workoutTypes).Run(); err != nil {
		return nil, err.GetResponse().Log("Failed to query workout typs", err, a)
	}
	if err := a.R().Db.Struct.QuerySlice(&rtc.tags).Run(); err != nil {
		return nil, err.GetResponse().Log("Failed to query workout tags", err, a)
	}

	return rtc, nil
}
