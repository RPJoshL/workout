package metric

import (
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
)

// StoreStepsResult contains the number of pushed steps
type StoreStepsResult struct {

	// Number of stored steps
	StoredCount int `json:"storedCount"`

	// Number of steps that weren't pushed to the database
	// (because they probably already exist)
	DroppedCount int `json:"droppedCount"`
}

// StoreSteps stores the provided steps within the database.
// Duplicate data is dropped and returned to the response
func (a *Api) StoreSteps(steps []models.Steps) (rtc StoreStepsResult, err errors.Error) {
	origCount := len(steps)

	// Nothing to do
	if len(steps) == 0 {
		return StoreStepsResult{}, nil
	}

	// Select all steps that do overlap with the provided one → stripe them away
	selectStmt := "1 = 0"
	placeholder := []any{}
	for i, step := range steps {
		// We plot every step value to full minutes
		steps[i].Start = step.Start.Truncate(time.Minute)
		steps[i].End = step.End.Truncate(time.Minute)

		// Add user ID to steps
		steps[i].UserId = a.R().User.Id
		steps[i].Id = 0

		selectStmt += " OR (? >= start AND ? <= end) OR (? >= start AND ? <= end) "
		placeholder = append(placeholder, steps[i].Start, steps[i].Start, steps[i].End, steps[i].End)
	}
	// Query overlapping steps
	overlappingSteps := []models.Steps{}
	sel := a.R().Db.Struct.QuerySlice(&overlappingSteps).Where().Custom(
		selectStmt, placeholder...,
	).Add()
	if e := sel.Run(); e != nil {
		return rtc, errors.InternalError().Log("Failed to select overlapping steps", e, a)
	}

	// Exclude overlapping steps
	for _, step := range overlappingSteps {
		// Find matching steps and remove them
		for i := 0; i < len(steps); i++ {
			if steps[i].Count <= 0 {
				// Drop negative step counts
				steps = utils.Remove(&steps, i)
				i--
			} else if steps[i].Start.Equal(step.Start) && steps[i].End.Equal(step.End) {
				// Equal
				steps = utils.Remove(&steps, i)
				i--
			} else if steps[i].Start.After(step.Start) && steps[i].Start.Before(step.End) {
				// Start time
				steps = utils.Remove(&steps, i)
				i--
			} else if steps[i].End.After(step.Start) && steps[i].End.Before(step.End) {
				// End time
				steps = utils.Remove(&steps, i)
				i--
			}
		}
	}

	// Insert steps (if no data)
	if len(steps) > 0 {
		if _, e := a.R().Db.Struct.InsertSlice(&steps).Run(); e != nil {
			return rtc, errors.InternalError().Log("Failed to insert steps", e, a)
		}
	}

	return StoreStepsResult{
		StoredCount:  len(steps),
		DroppedCount: origCount - len(steps),
	}, nil
}

// GetStepsSince returns the total number of steps that the user made
// since the provided time
func (a *Api) GetStepsSince(startDate time.Time) (rtc int, err errors.Error) {
	dbError := a.R().Db.QueryForValue(&rtc,
		`SELECT NVL(SUM(count), 0)
		 FROM steps
		 WHERE user_id = ?
		   AND start >= ?
		`, a.R().User.Id, startDate,
	)

	if dbError != nil {
		return 0, errors.InternalError().Log("Failed to query step count: %s", dbError, a)
	}

	return rtc, nil
}
