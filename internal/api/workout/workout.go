package workout

import (
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

var (
	ErrWorkoutNotFound = errors.NewError("#workout.notFound", 404)
)

// Delete deletes the workout by the provided ID
func (a *Api) Delete(id int) errors.Error {

	// Check if the workout exists within user context
	workout := &models.Workout{}
	sel := a.R().Db.Struct.Query(workout).Where().Column(models.Workout_UserId, "=", a.R().User.Id).Add()
	sel.Where().Column(models.Workout_Id, "=", id).Add()
	if count, err := sel.Count(); err != nil {
		return err.GetResponse().Log("Failed to count workouts", err, a)
	} else if count != 1 {
		return ErrWorkoutNotFound
	}

	// Delete it
	if _, err := a.R().Db.Db.Exec("DELETE FROM workout WHERE id = ?", id); err != nil {
		return errors.InternalError().Log("Failed to delete workout", err, a)
	}

	return nil
}
