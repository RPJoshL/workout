package details

import (
	"testing"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/tests"
	"git.rpjosh.de/RPJosh/workout/pkg/assert"
)

func TestModifyLocation(t *testing.T) {
	api := &Api{}
	tests.InjectRequestData(api, t)

	// Dummy workout entity
	workout := models.Workout{
		Id:     1001,
		UserId: api.R().User.Id,
		TypeId: models.TYPE_RUNNING,
		Start:  time.Now(),
		End:    time.Now().Add(1 * time.Hour),
	}
	_, err := api.R().Db.Struct.Insert(&workout).Run()
	assert.NoError(t, err)

	// 3 workout details with almost identical location IDs
	details := []models.WorkoutDetails{
		{WorkoutId: workout.Id, Latitude: 48.12345, Longitude: 11.12345, Part: 0},
		{WorkoutId: workout.Id, Latitude: 48.12346, Longitude: 11.12346, Part: 0},
		{WorkoutId: workout.Id, Latitude: 48.12345, Longitude: 11.12347, Part: 0},
	}
	_, err = api.R().Db.Struct.InsertSlice(&details).Run()
	assert.NoError(t, err)

	// Location update should work
	errApi := api.PatchWorkoutLocation(workout.Id, 48.99999, 11.99999)
	assert.NoError(t, errApi)

	// 3 other workouts with different lat/lon
	workout2 := models.Workout{Id: 1002, UserId: api.R().User.Id, TypeId: models.TYPE_RUNNING, Start: time.Now(), End: time.Now().Add(1 * time.Hour)}
	_, err = api.R().Db.Struct.Insert(&workout2).Run()
	assert.NoError(t, err)
	details2 := []models.WorkoutDetails{
		{WorkoutId: workout2.Id, Latitude: 47.1, Longitude: 10.1, Part: 0},
		{WorkoutId: workout2.Id, Latitude: 48.2, Longitude: 11.2, Part: 0},
		{WorkoutId: workout2.Id, Latitude: 49.3, Longitude: 12.3, Part: 0},
	}
	_, err = api.R().Db.Struct.InsertSlice(&details2).Run()
	assert.NoError(t, err)

	// Location update should fail
	errApi = api.PatchWorkoutLocation(workout2.Id, 48.88888, 11.88888)
	assert.ErrorIs(t, errApi, ErrLocationUpdateNotAllowed)
}
