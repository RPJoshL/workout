package metric

import (
	"testing"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/tests"
	"git.rpjosh.de/RPJosh/workout/pkg/assert"
	"github.com/RPJoshL/go-logger"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestDummy(t *testing.T) {
	api := &Api{}
	tests.InjectRequestData(api, t)

	count := 0
	api.R().Db.QueryForValue(&count, `SELECT COUNT(*) FROM year_day WHERE start = ?`, dummyDate(10, 0, 0))
	logger.Info("Count = %d", count)
}

type dummyPai struct {
	start time.Time
	end   time.Time

	value int
}

func insertDummyPai(t *testing.T, r *router.Request, steps, workout []dummyPai) {
	t.Helper()

	for _, data := range workout {
		end := data.end
		if end.IsZero() {
			end = data.start.Add(5 * time.Minute)
		}

		_, err := r.Db.Struct.Insert(&models.Workout{
			Start:  data.start,
			End:    end,
			Pai:    data.value,
			UserId: r.User.Id,
			TypeId: 1,
		}).Run()
		assert.NoError(t, err)
	}
	for _, data := range steps {
		end := data.end
		if end.IsZero() {
			end = data.start.Add(5 * time.Minute)
		}

		_, err := r.Db.Struct.Insert(&models.Steps{
			Start:  data.start,
			End:    end,
			Count:  data.value,
			UserId: r.User.Id,
		}).Run()
		assert.NoError(t, err)
	}
}

func TestGetWeeklyPaiScore(t *testing.T) {
	startUnix := 19946

	allTests := map[string]struct {
		workout []dummyPai
		steps   []dummyPai

		expected []PaiDay
	}{
		"workouts": {
			workout: []dummyPai{
				// Pre values
				{start: dummyDate(7, 10, 0), value: 10},
				{start: dummyDate(9, 10, 0), value: 12},
				// Normale values
				{start: dummyDate(10, 10, 0), value: 15},
				{start: dummyDate(15, 10, 0), value: 11},
				{start: dummyDate(16, 10, 0), value: 13},
			},
			expected: []PaiDay{
				{Value: 37, Earned: 15, DayIndex: startUnix, WeekdayIndex: 5},     // 10
				{Value: 37, Earned: 0, DayIndex: startUnix + 1, WeekdayIndex: 6},  // 11
				{Value: 37, Earned: 0, DayIndex: startUnix + 2, WeekdayIndex: 0},  // 12
				{Value: 37, Earned: 0, DayIndex: startUnix + 3, WeekdayIndex: 1},  // 13
				{Value: 27, Earned: 0, DayIndex: startUnix + 4, WeekdayIndex: 2},  // 14
				{Value: 38, Earned: 11, DayIndex: startUnix + 5, WeekdayIndex: 3}, // 15
				{Value: 39, Earned: 13, DayIndex: startUnix + 6, WeekdayIndex: 4}, // 16
			},
		},
		"with steps": {
			workout: []dummyPai{
				{start: dummyDate(10, 10, 0), value: 15},
			},
			steps: []dummyPai{
				{start: dummyDate(7, 10, 0), value: 10_000},
				{start: dummyDate(7, 15, 0), value: 15_000},
				{start: dummyDate(7, 19, 30), end: dummyDate(8, 10, 0), value: 6_000},
				// Normale values
				{start: dummyDate(10, 10, 0), value: 12_000},
				{start: dummyDate(16, 10, 0), value: 34_000},
			},
			expected: []PaiDay{
				{Value: 27, Earned: 17, DayIndex: startUnix, WeekdayIndex: 5},     // 10
				{Value: 27, Earned: 0, DayIndex: startUnix + 1, WeekdayIndex: 6},  // 11
				{Value: 27, Earned: 0, DayIndex: startUnix + 2, WeekdayIndex: 0},  // 12
				{Value: 27, Earned: 0, DayIndex: startUnix + 3, WeekdayIndex: 1},  // 13
				{Value: 17, Earned: 0, DayIndex: startUnix + 4, WeekdayIndex: 2},  // 14
				{Value: 17, Earned: 0, DayIndex: startUnix + 5, WeekdayIndex: 3},  // 15
				{Value: 27, Earned: 10, DayIndex: startUnix + 6, WeekdayIndex: 4}, // 16
			},
		},
		// Dummy user has +2
		"timezones": {
			workout: []dummyPai{
				{start: dummyDate(12, 1, 0), value: 15}, // => 11
			},
			steps: []dummyPai{
				{start: dummyDate(13, 1, 0), value: 11_000}, // => 12
			},
			expected: []PaiDay{
				{Value: 0, Earned: 0, DayIndex: startUnix, WeekdayIndex: 5},       // 10
				{Value: 15, Earned: 15, DayIndex: startUnix + 1, WeekdayIndex: 6}, // 11
				{Value: 17, Earned: 2, DayIndex: startUnix + 2, WeekdayIndex: 0},  // 12
				{Value: 17, Earned: 0, DayIndex: startUnix + 3, WeekdayIndex: 1},  // 13
				{Value: 17, Earned: 0, DayIndex: startUnix + 4, WeekdayIndex: 2},  // 14
				{Value: 17, Earned: 0, DayIndex: startUnix + 5, WeekdayIndex: 3},  // 15
				{Value: 17, Earned: 0, DayIndex: startUnix + 6, WeekdayIndex: 4},  // 16
			},
		},
	}

	for name, test := range allTests {
		t.Run(name, func(t *testing.T) {
			api := &Api{}
			tests.InjectRequestData(api, t)

			// Insert all dummy data
			insertDummyPai(t, api.R(), test.steps, test.workout)

			// Cache steps
			err := api.cacheStepsPAI(dummyDate(0, 0, 0), dummyDate(17, 0, 0), api.R().User.Id)
			assert.NoError(t, err)

			values, err := api.GetWeeklyPaiScore(
				dummyDate(10, 0, 0),
				dummyDate(17, 0, 0),
			)
			assert.NoError(t, err)
			assert.EqualStruct(
				t, "weekkly PAI score", test.expected, values,
				cmpopts.IgnoreFields(PaiDay{}, "WeekdayShort"),
			)
		})
	}
}

func TestSumPaiScore(t *testing.T) {
	api := &Api{}
	tests.InjectRequestData(api, t)

	insertDummyPai(t, api.R(), []dummyPai{
		{start: dummyDate(9, 10, 0), value: 24_000}, // Not included
		{start: dummyDate(16, 10, 0), value: 34_000},
	}, []dummyPai{
		{start: dummyDate(9, 12, 0), value: 11}, // Not included
		{start: dummyDate(10, 12, 0), value: 1},
		{start: dummyDate(14, 12, 0), value: 15},
	})

	// Cache steps
	err := api.cacheStepsPAI(dummyDate(0, 0, 0), dummyDate(17, 0, 0), api.R().User.Id)
	assert.NoError(t, err)

	expected := 10 + 1 + 15

	got, err := api.GetSumOfPai(
		dummyDate(10, 0, 0),
		dummyDate(17, 0, 0),
	)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func dummyDate(day, hour, minute int) time.Time {
	return time.Date(2024, time.August, day, hour, minute, 0, 0, time.UTC)
}
