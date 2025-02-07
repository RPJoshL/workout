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
	s := time.Date(2024, time.April, 1, 2, 0, 0, 0, time.Now().Location())

	// Basic test
	steps := []models.Steps{
		{Start: s, End: s.Add(20 * time.Minute), Count: 10},
		{Start: s.Add(20 * time.Minute), End: s.Add(40 * time.Minute), Count: 20},
	}
	expectSteps(t, api, "Basic insert", 1, steps, 2, 0)

	// Add another (non overlapping) steps
	steps = []models.Steps{
		{Start: s.Add(60 * time.Minute), End: s.Add(80 * time.Minute), Count: 20},
		{Start: s.Add(100 * time.Minute), End: s.Add(120 * time.Minute), Count: 20},
	}
	expectSteps(t, api, "Insert with existing data", 2, steps, 2, 0)

	// Add steps again → no steps should be inserted
	expectSteps(t, api, "Duplicate insert", 3, steps, 0, 2)

	// Start time overlap
	steps = []models.Steps{
		{Start: s.Add(10 * time.Minute), End: s.Add(30 * time.Minute)},
	}
	expectSteps(t, api, "Start time overlap", 4, steps, 0, 1)

	// End time overlap
	steps = []models.Steps{
		{Start: s.Add(-10 * time.Minute), End: s.Add(10 * time.Minute)},
		{Start: s.Add(140 * time.Minute), End: s.Add(160 * time.Minute)},
	}
	expectSteps(t, api, "End time overlap", 5, steps, 1, 1)

	// Insert before an existing step value inside db
	steps = []models.Steps{
		{Start: s.Add(120 * time.Minute), End: s.Add(140 * time.Minute)},
	}
	expectSteps(t, api, "Insert step before existing", 6, steps, 1, 0)
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
