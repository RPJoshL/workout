package metric

import (
	"testing"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/tests"
)

// TestStoreSteps tests the storing of overlapping steps
func TestStoreSteps(t *testing.T) {
	api := &Api{}
	tests.InjectRequestData(api, t)

	// Start time used for all tests
	s := time.Date(2024, time.April, 1, 14, 0, 0, 0, time.Now().Location())

	// Basic test
	steps := []models.Steps{
		{Start: s, End: s.Add(20 * time.Minute), Count: 10},
		{Start: s.Add(20 * time.Minute), End: s.Add(40 * time.Minute), Count: 20},
	}
	expectSteps(t, api, "Basic insert", 1, steps, 2, 0)
	expectStepsPai(t, api, 0, s)

	// Add another (non overlapping) steps
	steps = []models.Steps{
		{Start: s.Add(60 * time.Minute), End: s.Add(80 * time.Minute), Count: 20},
		{Start: s.Add(100 * time.Minute), End: s.Add(120 * time.Minute), Count: 20},
	}
	expectSteps(t, api, "Insert with existing data", 2, steps, 2, 0)
	expectStepsPai(t, api, 0, s)

	// Add steps again → no steps should be inserted
	expectSteps(t, api, "Duplicate insert", 3, steps, 0, 2)

	// Start time overlap
	steps = []models.Steps{
		{Start: s.Add(10 * time.Minute), End: s.Add(30 * time.Minute), Count: 20},
	}
	expectSteps(t, api, "Start time overlap", 4, steps, 0, 1)

	// End time overlap
	steps = []models.Steps{
		{Start: s.Add(-10 * time.Minute), End: s.Add(10 * time.Minute), Count: 20},
		{Start: s.Add(140 * time.Minute), End: s.Add(160 * time.Minute), Count: 20},
	}
	expectSteps(t, api, "End time overlap", 5, steps, 1, 1)

	// Insert before an existing step value inside db
	steps = []models.Steps{
		{Start: s.Add(120 * time.Minute), End: s.Add(140 * time.Minute), Count: 20},
	}
	expectSteps(t, api, "Insert step before existing", 6, steps, 1, 0)
}

func TestPaiStoreSteps(t *testing.T) {
	api := &Api{}
	tests.InjectRequestData(api, t)

	// Start time used for all tests
	s := time.Date(2024, time.April, 1, 14, 0, 0, 0, time.Now().Location())

	steps := []models.Steps{
		{Start: s.Add(0 * time.Minute), End: s.Add(40 * time.Minute), Count: 10_200},
	}
	expectSteps(t, api, "Insert step for PAI test", 6, steps, 1, 0)
	expectStepsPai(t, api, 2, s)

	// No increment => we should still only get one row (the other got deleted due to overlap)
	steps = []models.Steps{
		{Start: s.Add(100 * time.Minute), End: s.Add(120 * time.Minute), Count: 10},
	}
	expectSteps(t, api, "Insert step for PAI test", 6, steps, 1, 0)
	expectStepsPai(t, api, 2, s)

	steps = []models.Steps{
		{Start: s.Add(200 * time.Minute), End: s.Add(220 * time.Minute), Count: 10_200},
	}
	expectSteps(t, api, "Insert step for PAI test", 6, steps, 1, 0)
	expectStepsPai(t, api, 5, s)

	steps = []models.Steps{
		{Start: s.Add(300 * time.Minute), End: s.Add(320 * time.Minute), Count: 10_200},
	}
	expectSteps(t, api, "Insert step for PAI test", 6, steps, 1, 0)
	expectStepsPai(t, api, 10, s)

	// Inserted for next day
	nextDay := time.Date(2024, time.April, 2, 6, 0, 0, 0, time.Now().Location())
	steps = []models.Steps{
		{Start: nextDay.Add(0 * time.Minute), End: nextDay.Add(40 * time.Minute), Count: 10_200},
	}
	expectSteps(t, api, "Insert step for PAI test", 6, steps, 1, 0)
	expectStepsPai(t, api, 2, nextDay)
	expectStepsPai(t, api, 10, s)
}

func expectSteps(t *testing.T, api *Api, action string, id int, input []models.Steps, saved, dropped int) {
	t.Helper()

	if res, err := api.StoreSteps(input); err != nil {
		t.Errorf("StepStore (%d: %s) - received error: %s", id, action, err)
	} else if res.StoredCount != saved {
		t.Errorf("StepStore (%d: %s) - Expected %d saved steps. Got %d", id, action, saved, res.StoredCount)
	} else if res.DroppedCount != dropped {
		t.Errorf("StepStore (%d: %s) - Expected %d dropped steps. Got %d", id, action, dropped, res.DroppedCount)
	}
}

func expectStepsPai(t *testing.T, api *Api, pai int, baseDate time.Time) {
	t.Helper()

	sql := `
		SELECT p.pai
		FROM steps_pai p
		INNER JOIN v_year_day_user_offset glob ON glob.id = p.id AND glob.user_id = ?
		WHERE glob.start <= ?
		  AND glob.end >= ?
		  AND p.user_id = ?
	`

	var got int
	if err := api.R().Db.QueryForValue(&got, sql, api.R().User.Id, baseDate, baseDate, api.R().User.Id); err != nil {
		t.Errorf("StepsPAI - received error: %s", err)
	} else if got != pai {
		t.Errorf("StepsPAI - Expected %d PAI. Got %d", pai, got)
	}
}
