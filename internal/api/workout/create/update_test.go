package create

import (
	"testing"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/dbutils"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/tests"
	"git.rpjosh.de/RPJosh/workout/pkg/database/dbstruct"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"github.com/google/go-cmp/cmp"
	"github.com/guregu/null/v5"
)

// TestMergeTooFar tests the merging of two workouts with
// a time difference > 12 hours
func TestMergeTooFar(t *testing.T) {
	api := &Api{}
	tests.InjectRequestData(api, t)

	// Insert dummy workouts
	testData := []models.Workout{
		{
			Id:     1,
			Start:  time.Now(),
			End:    time.Now().Add(10 * time.Minute),
			UserId: api.R().User.Id,
			TypeId: models.TYPE_CYCLING,
		},
		{
			Id:     2,
			Start:  time.Now().Add(-24 * time.Hour),
			End:    time.Now().Add(5 * time.Minute).Add(-24 * time.Hour),
			UserId: api.R().User.Id,
			TypeId: models.TYPE_CYCLING,
		},
	}
	if _, err := api.R().Db.Struct.InsertSlice(&testData).Run(); err != nil {
		t.Fatalf("Failed to insert test data: %s", err)
	}

	errMerge := api.MergeWorkouts(testData[0].Id, testData[1].Id)
	if errors.IsNot(errMerge, ErrWorkoutsTooFarAway) {
		t.Errorf("Incorrect or no error found. Expected 'ErrWorkoutsTooFarAway'. Got %s", errMerge)
	}
}

// TestMerge tests the merging of two different workouts into a single one
func TestMerge(t *testing.T) {
	api := &Api{}
	tests.InjectRequestData(api, t)

	// Insert dummy workouts
	createDumyTag(1, t, api.R().Db)
	createDumyTag(2, t, api.R().Db)
	createDumyTag(3, t, api.R().Db)
	testData := []models.Workout{
		{
			Id:              2,
			Start:           time.Now(),
			End:             time.Now().Add(1 * time.Minute),
			Duration:        60,
			City:            "First",
			Calories:        150,
			CaloriesDefault: 5,
			ElevationUp:     1,
			ElevationDown:   1,
			Distance:        1500,
			TypeId:          1,
			Pai:             2,
			HeartRateMax:    null.IntFrom(150),
			HeartRateAv:     null.IntFrom(200),
			WorkoutTags: []models.WorkoutTags{
				{
					WorkoutId: 2,
					TagId: &models.Tag{
						Id: 1,
					},
				},
				{
					WorkoutId: 2,
					TagId: &models.Tag{
						Id: 3,
					},
				},
			},
			Note:   null.StringFrom("From1"),
			UserId: api.R().User.Id,
			WorkoutDetails: []models.WorkoutDetails{
				{
					Id:        1,
					WorkoutId: 2,
					Time:      time.Now(),
					Duration:  0,
					Distance:  0,
				},
				{
					Id:        2,
					WorkoutId: 2,
					Time:      time.Now(),
					Duration:  60,
					Distance:  1500,
				},
			},
		},
		{
			Id:              1,
			Start:           time.Now().Add(10 * time.Minute),
			End:             time.Now().Add(12 * time.Minute),
			Duration:        120,
			City:            "Second",
			Calories:        200,
			CaloriesDefault: 10,
			ElevationUp:     2,
			ElevationDown:   2,
			Distance:        500,
			TypeId:          2,
			Pai:             3,
			HeartRateMax:    null.IntFrom(160),
			HeartRateAv:     null.IntFrom(100),
			WorkoutTags: []models.WorkoutTags{
				{
					WorkoutId: 1,
					TagId: &models.Tag{
						Id: 1,
					},
				},
				{
					WorkoutId: 1,
					TagId: &models.Tag{
						Id: 2,
					},
				},
			},
			Note:   null.StringFrom("From2"),
			UserId: api.R().User.Id,
			WorkoutDetails: []models.WorkoutDetails{
				{
					Id:        3,
					WorkoutId: 1,
					Time:      time.Now(),
					Duration:  0,
					Distance:  0,
				},
				{
					Id:        4,
					WorkoutId: 1,
					Time:      time.Now(),
					Duration:  120,
					Distance:  500,
				},
			},
		},
	}
	if _, err := api.R().Db.Struct.InsertSlice(&testData).Selector(dbstruct.ColumnSelector{PointedKeyReference: true}).Run(); err != nil {
		t.Fatalf("Failed to insert test data: %s", err)
	}

	// Merge workouts
	errMerge := api.MergeWorkouts(testData[0].Id, testData[1].Id)
	if errMerge != nil {
		t.Fatalf("Failed to merge workouts: %s", errMerge)
	}

	// Expected (merged) workout
	expcted := models.Workout{
		Id:              2,
		Start:           testData[0].Start,
		End:             testData[1].End,
		Duration:        180,
		Calories:        350,
		CaloriesDefault: 15,
		ElevationUp:     3,
		ElevationDown:   3,
		Distance:        2000,
		SpeedAv:         90,
		Pai:             5,
		HeartRateMax:    null.IntFrom(160),
		HeartRateAv:     null.IntFrom(133),
		UserId:          api.R().User.Id,

		// Use values from the first workout
		City:   "First",
		TypeId: 1,

		// Merging of tags
		WorkoutTags: []models.WorkoutTags{
			{
				WorkoutId: 2,
				TagId:     &models.Tag{Id: 1},
			},
			{
				WorkoutId: 2,
				TagId:     &models.Tag{Id: 2},
			},
			{
				WorkoutId: 2,
				TagId:     &models.Tag{Id: 3},
			},
		},
		Note: null.StringFrom(`From1
From2`),

		// Merging of details
		WorkoutDetails: []models.WorkoutDetails{
			{Id: 1, Duration: 0, Distance: 0},
			{Id: 2, Duration: 60, Distance: 1500},
			{Id: 3, Duration: 61, Distance: 1500, Part: 1},
			{Id: 4, Duration: 181, Distance: 2000, Part: 1},
		},
	}

	// Merged workout within DB
	var got models.Workout
	if err := api.R().Db.Struct.Query(&got).Selector(dbstruct.ColumnSelector{PointedKeyReference: true}).Run(); err != nil {
		t.Fatalf("Failed to query created workout: %s", err)
	}

	if diff := cmp.Diff(expcted, got, cmp.Comparer(func(x, y time.Time) bool {
		return x.Unix() == y.Unix()
	}), cmp.Comparer(func(x, y models.WorkoutDetails) bool {
		return y.Id == x.Id && y.Duration == x.Duration && y.Distance == x.Distance
	})); diff != "" {
		t.Errorf("Mismatch of merge (-want +got):\n%s", diff)
	}

	// logger.Debug("%d - %d", expcted.Start.Unix(), got.Start.Unix())
}

func createDumyTag(tagId int, t *testing.T, db *dbutils.Db) {
	tag := models.Tag{
		Id: tagId,
	}

	if _, err := db.Struct.Insert(&tag).Run(); err != nil {
		t.Fatalf("Failed to insert Tag %d", tagId)
	}
}
