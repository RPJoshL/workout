package parser

import (
	"testing"

	"git.rpjosh.de/RPJosh/go-ddl-parser"
	"git.rpjosh.de/RPJosh/workout/internal/dbutils"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/tests"
	"git.rpjosh.de/RPJosh/workout/pkg/assert"
	"github.com/google/go-cmp/cmp"
	"github.com/guregu/null/v5"
)

// TestRules tests the applaying of tagging rules
func TestRules(t *testing.T) {
	tags := []models.Tag{
		{Id: 1, Name: "1"},
		{Id: 2, Name: "2"},
	}

	allTests := map[string]struct {
		expected []int
		rules    []models.RuleTagging
		workout  models.Workout
	}{
		"durationMinMatch": {
			[]int{1},
			[]models.RuleTagging{
				{TagId: 1, DurationMin: null.IntFrom(100)},
			},
			models.Workout{
				Duration: 200,
			},
		},
		"tagMergingNoDuplicate": {
			[]int{1, 2},
			[]models.RuleTagging{
				{TagId: 1, DurationMin: null.IntFrom(100)},
				{TagId: 2, DurationMin: null.IntFrom(130)},
			},
			models.Workout{
				Duration: 200,
				WorkoutTags: []models.WorkoutTags{
					{TagId: &models.Tag{Id: 1}},
				},
			},
		},
		"durationMaxMatch": {
			[]int{1},
			[]models.RuleTagging{
				{TagId: 1, DurationMax: null.IntFrom(100)},
			},
			models.Workout{
				Duration: 50,
			},
		},
		"durationBetweenMatch": {
			[]int{1},
			[]models.RuleTagging{
				{TagId: 1, DurationMin: null.IntFrom(40), DurationMax: null.IntFrom(100)},
			},
			models.Workout{
				Duration: 60,
			},
		},
		"durationBetweenNoMatch": {
			[]int{},
			[]models.RuleTagging{
				{TagId: 1, DurationMin: null.IntFrom(40), DurationMax: null.IntFrom(100)},
			},
			models.Workout{
				Duration: 200,
			},
		},
		"startLocationWithinRadius": {
			[]int{1},
			[]models.RuleTagging{
				{TagId: 1, StartLocation: &models.AreaCircle{
					Radius: 200, Center: ddl.Location{Latitude: 48.66, Longitude: 10.84},
				}},
			},
			models.Workout{
				WorkoutDetails: []models.WorkoutDetails{
					{Latitude: 48.656, Longitude: 10.836},
					{Latitude: 48.64, Longitude: 10.85},
				},
			},
		},
		"startLocationOutOfRadius": {
			[]int{1},
			[]models.RuleTagging{
				{TagId: 1, StartLocation: &models.AreaCircle{
					Radius: 200, Center: ddl.Location{Latitude: 48.66, Longitude: 10.84},
				}},
			},
			models.Workout{
				WorkoutDetails: []models.WorkoutDetails{
					{Latitude: 48.64, Longitude: 10.85},
					{Latitude: 48.63, Longitude: 10.85},
				},
			},
		},
	}

	for name, test := range allTests {
		t.Run(name, func(t *testing.T) {
			db := dbutils.NewByDb(tests.GetDbConnection(t))
			tests.CreateDefaultUser(db)

			// Insert tags
			_, err := db.Struct.InsertSlice(&tags).Run()
			assert.NoErrorf(t, err, "Failed to insert tags")

			// Insert automation rules
			for i := range test.rules {
				test.rules[i].UserId = tests.DefaultUserID
			}
			_, err = db.Struct.InsertSlice(&test.rules).Run()
			assert.NoErrorf(t, err, "Failed to insert testing rules")

			// Fill dummy data into workout
			test.workout.UserId = tests.DefaultUserID

			// Call function
			errR := ApplyRules(&test.workout, db)
			assert.NoErrorf(t, errR, "Failed to call applyRules")

			// Expect same tags
			gotTagIds := []int{}
			for _, tag := range test.workout.WorkoutTags {
				gotTagIds = append(gotTagIds, tag.TagId.Id)
			}

			if diff := cmp.Diff(test.expected, gotTagIds); diff != "" {
				t.Error("MIssmatch of workout tags (-want +got):\n" + diff)
			}
		})
	}
}
