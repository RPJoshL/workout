package statistics

import (
	"testing"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/api/workout/shared"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/tests"
	"git.rpjosh.de/RPJosh/workout/pkg/assert"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/guregu/null/v5"
)

func TestWorkoutSumAvg(t *testing.T) {
	api := &Api{}
	tests.InjectRequestData(api, t)

	data := []models.Workout{
		// One workout a day
		{
			Start:           mockDate(0, 0),
			Duration:        20,
			Distance:        20000,
			Pai:             5,
			HeartRateAv:     null.IntFrom(120),
			SpeedAv:         300,
			Calories:        100,
			CaloriesDefault: 20,
			TypeId:          models.TYPE_CYCLING,
		},
		// Two workouts of the same type
		{
			Start:           mockDate(1, 0),
			Duration:        20,
			Distance:        20000,
			Pai:             10,
			HeartRateAv:     null.IntFrom(120),
			SpeedAv:         300,
			Calories:        120,
			CaloriesDefault: 20,
			TypeId:          models.TYPE_CYCLING,
		},
		{
			Start:           mockDate(1, 2),
			Duration:        40,
			Distance:        40000,
			Pai:             20,
			HeartRateAv:     null.IntFrom(150),
			SpeedAv:         600,
			Calories:        240,
			CaloriesDefault: 40,
			TypeId:          models.TYPE_CYCLING,
		},
		// Two workouts of different type
		{
			Start:           mockDate(2, 0),
			Duration:        20,
			Distance:        20000,
			Pai:             10,
			HeartRateAv:     null.IntFrom(120),
			SpeedAv:         300,
			Calories:        120,
			CaloriesDefault: 20,
			TypeId:          models.TYPE_CYCLING,
		},
		{
			Start:           mockDate(2, 2),
			Duration:        40,
			Distance:        40000,
			Pai:             20,
			HeartRateAv:     null.IntFrom(150),
			SpeedAv:         600,
			Calories:        240,
			CaloriesDefault: 40,
			TypeId:          models.TYPE_HIKING,
		},
	}

	// Insert test data
	for i := range data {
		data[i].UserId = tests.DefaultUserID
	}
	_, dbErr := api.R().Db.Struct.InsertSlice(&data).Run()
	assert.NoError(t, dbErr)

	expectedSum := []workoutData{
		{
			statisticsRow: statisticsRow{
				Start: mockDateAbsolute(0, 2, 0, 0),
				End:   mockDateAbsolute(1, 1, 59, 59),
				Label: "10.09",
			},
			Distance: map[int]float64{
				-1:                  20000,
				models.TYPE_CYCLING: 20000,
			},
			Calories: map[int]float64{
				-1:                  80,
				models.TYPE_CYCLING: 80,
			},
			Duration: map[int]float64{
				-1:                  20,
				models.TYPE_CYCLING: 20,
			},
			PAI: map[int]float64{
				-1:                  5,
				models.TYPE_CYCLING: 5,
			},
			Count: map[int]float64{
				-1:                  1,
				models.TYPE_CYCLING: 1,
			},
			Heartrate: map[int]float64{
				-1:                  120,
				models.TYPE_CYCLING: 120,
			},
			Speed: map[int]float64{
				-1:                  12,
				models.TYPE_CYCLING: 12,
			},
		},
		{
			statisticsRow: statisticsRow{
				Start: mockDateAbsolute(1, 2, 0, 0),
				End:   mockDateAbsolute(2, 1, 59, 59),
				Label: "11.09",
			},
			Distance: map[int]float64{
				-1:                  60_000,
				models.TYPE_CYCLING: 60_000,
			},
			Calories: map[int]float64{
				-1:                  300,
				models.TYPE_CYCLING: 300,
			},
			Duration: map[int]float64{
				-1:                  60,
				models.TYPE_CYCLING: 60,
			},
			PAI: map[int]float64{
				-1:                  30,
				models.TYPE_CYCLING: 30,
			},
			Count: map[int]float64{
				-1:                  2,
				models.TYPE_CYCLING: 2,
			},
			Heartrate: map[int]float64{
				-1:                  135,
				models.TYPE_CYCLING: 135,
			},
			Speed: map[int]float64{
				-1:                  8, // 8 are correct instead of 9 km/h. We want to use the pace
				models.TYPE_CYCLING: 8, // 8 are correct instead of 9 km/h. We want to use the pace
			},
		},
		{
			statisticsRow: statisticsRow{
				Start: mockDateAbsolute(2, 2, 0, 0),
				End:   mockDateAbsolute(3, 1, 59, 59),
				Label: "12.09",
			},
			Distance: map[int]float64{
				-1:                  60_000,
				models.TYPE_CYCLING: 20_000,
				models.TYPE_HIKING:  40_000,
			},
			Calories: map[int]float64{
				-1:                  300,
				models.TYPE_CYCLING: 100,
				models.TYPE_HIKING:  200,
			},
			Duration: map[int]float64{
				-1:                  60,
				models.TYPE_CYCLING: 20,
				models.TYPE_HIKING:  40,
			},
			PAI: map[int]float64{
				-1:                  30,
				models.TYPE_CYCLING: 10,
				models.TYPE_HIKING:  20,
			},
			Count: map[int]float64{
				-1:                  2,
				models.TYPE_CYCLING: 1,
				models.TYPE_HIKING:  1,
			},
			Heartrate: map[int]float64{
				-1:                  135,
				models.TYPE_CYCLING: 120,
				models.TYPE_HIKING:  150,
			},
			Speed: map[int]float64{
				-1:                  8, // 8 are correct instead of 9 km/h. We want to use the pace
				models.TYPE_CYCLING: 12,
				models.TYPE_HIKING:  6,
			},
		},
	}

	expectedAvg := []workoutData{
		{
			statisticsRow: statisticsRow{
				Start: mockDateAbsolute(0, 2, 0, 0),
				End:   mockDateAbsolute(1, 1, 59, 59),
				Label: "10.09",
			},
			Distance: map[int]float64{
				-1:                  20000,
				models.TYPE_CYCLING: 20000,
			},
			Calories: map[int]float64{
				-1:                  80,
				models.TYPE_CYCLING: 80,
			},
			Duration: map[int]float64{
				-1:                  20,
				models.TYPE_CYCLING: 20,
			},
			PAI: map[int]float64{
				-1:                  5,
				models.TYPE_CYCLING: 5,
			},
			Count: map[int]float64{
				-1:                  1,
				models.TYPE_CYCLING: 1,
			},
			Heartrate: map[int]float64{
				-1:                  120,
				models.TYPE_CYCLING: 120,
			},
			Speed: map[int]float64{
				-1:                  12,
				models.TYPE_CYCLING: 12,
			},
		},
		{
			statisticsRow: statisticsRow{
				Start: mockDateAbsolute(1, 2, 0, 0),
				End:   mockDateAbsolute(2, 1, 59, 59),
				Label: "11.09",
			},
			Distance: map[int]float64{
				-1:                  30_000,
				models.TYPE_CYCLING: 30_000,
			},
			Calories: map[int]float64{
				-1:                  150,
				models.TYPE_CYCLING: 150,
			},
			Duration: map[int]float64{
				-1:                  30,
				models.TYPE_CYCLING: 30,
			},
			PAI: map[int]float64{
				-1:                  15,
				models.TYPE_CYCLING: 15,
			},
			Count: map[int]float64{
				-1:                  2,
				models.TYPE_CYCLING: 2,
			},
			Heartrate: map[int]float64{
				-1:                  135,
				models.TYPE_CYCLING: 135,
			},
			Speed: map[int]float64{
				-1:                  8, // 8 are correct instead of 9 km/h. We want to use the pace
				models.TYPE_CYCLING: 8, // 8 are correct instead of 9 km/h. We want to use the pace
			},
		},
		{
			statisticsRow: statisticsRow{
				Start: mockDateAbsolute(2, 2, 0, 0),
				End:   mockDateAbsolute(3, 1, 59, 59),
				Label: "12.09",
			},
			Distance: map[int]float64{
				-1:                  30_000,
				models.TYPE_CYCLING: 20_000,
				models.TYPE_HIKING:  40_000,
			},
			Calories: map[int]float64{
				-1:                  150,
				models.TYPE_CYCLING: 100,
				models.TYPE_HIKING:  200,
			},
			Duration: map[int]float64{
				-1:                  30,
				models.TYPE_CYCLING: 20,
				models.TYPE_HIKING:  40,
			},
			PAI: map[int]float64{
				-1:                  15,
				models.TYPE_CYCLING: 10,
				models.TYPE_HIKING:  20,
			},
			Count: map[int]float64{
				-1:                  2,
				models.TYPE_CYCLING: 1,
				models.TYPE_HIKING:  1,
			},
			Heartrate: map[int]float64{
				-1:                  135,
				models.TYPE_CYCLING: 120,
				models.TYPE_HIKING:  150,
			},
			Speed: map[int]float64{
				-1:                  8, // 8 are correct instead of 9 km/h. We want to use the pace
				models.TYPE_CYCLING: 12,
				models.TYPE_HIKING:  6,
			},
		},
	}

	compareOptions := []cmp.Option{
		cmp.AllowUnexported(workoutData{}),
		cmpopts.IgnoreFields(statisticsRow{}, "ID"),
		cmpopts.IgnoreFields(workoutData{}, "rowCnt"),
	}

	sum, err := api.getWorkoutData(mockDate(1, 0), SamplingDay, AggregateFunctionSum, 3, &shared.WorkoutFilter{})
	assert.NoError(t, err)
	assert.EqualStruct(t, "Sum", expectedSum, sum, compareOptions...)

	avg, err := api.getWorkoutData(mockDate(1, 0), SamplingDay, AggregateFunctionAvg, 3, &shared.WorkoutFilter{})
	assert.NoError(t, err)
	assert.EqualStruct(t, "Avg", expectedAvg, avg, compareOptions...)
}

func mockDate(day, hour int) time.Time {
	base := time.Date(2025, time.September, 10, 10, 0, 0, 0, time.UTC)
	base = base.Add(time.Hour * time.Duration((24*day)+hour))

	return base
}
func mockDateAbsolute(day, hour, min, sec int) time.Time {
	base := time.Date(2025, time.September, 10+day, hour, min, sec, 0, time.UTC)

	return base
}
