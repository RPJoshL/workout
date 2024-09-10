package create

import (
	"math"

	"git.rpjosh.de/RPJosh/workout/internal/database"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"github.com/guregu/null/v5"
)

var (
	ErrFieldRequired      = errors.NewError("Field %q is required", 400)
	ErrWorkoutsTooFarAway = errors.NewError("#workout.tooFarAwayTime", 400)
	ErrWorkoutsSame       = errors.NewError("#workout.chooseNotSame", 400)

	// Maximum allowed duration in seconds how far two workouts can be
	// seperated by each other
	MergeAllowedTimeOffset int64 = 12 * 60 * 60
)

// Fields of a workout that can be updated from the form
var FormUpdateFields = []string{
	models.Workout_TypeId,
	models.Workout_WorkoutTags,
	models.Workout_Name,
	models.Workout_Note,
	models.Workout_City,
}

// UpdateWorkout updates a single workout with the provided data
func (a *Api) UpdateWorkout(id int, data *WorkoutCreateUpdate) errors.Error {

	// Get and validate tags
	workoutTags, e := a.validateTags(data.Tags)
	if e != nil {
		return e
	}

	// Validate type
	if err := a.validateType(data.Type); err != nil {
		return err
	}
	if data.Type == 0 {
		return ErrFieldRequired.Sprintf("type")
	}

	// Get existing workout
	existingWorkout, err := a.getExistingWorkout(id)
	if err != nil {
		return err
	}

	// Patch workout
	workout := models.Workout{
		Id:          id,
		TypeId:      data.Type,
		WorkoutTags: workoutTags,
		Name:        data.Name,
		Note:        null.NewString(data.Note, data.Note != ""),
		City:        data.City,
	}

	// Update default name if old workout type is unknown
	if existingWorkout.TypeId == models.TYPE_HIKING && workout.Name == a.R().Tr.Get("workout.unknown") {
		workout.Name = a.getTypeName(workout.TypeId)
	}

	sel := a.R().Db.Struct.Update(&workout).Selector(database.ColumnSelector{IncludeColumns: FormUpdateFields, PointedKeyReference: true})
	if err := sel.Run(); err != nil {
		return err.GetResponse().Log("Failed to update workout", err, a)
	}

	return nil
}

// MergeWorkouts combines two seperate workouts into a single one.
// The time between the workouts are counted as a break
func (a *Api) MergeWorkouts(id1, id2 int) errors.Error {

	// Workouts cannot be the same
	if id1 == id2 {
		return ErrWorkoutsSame
	}

	// Get workout headers
	workouts := []models.Workout{}
	sel := a.R().Db.Struct.QuerySlice(&workouts).Selector(database.ColumnSelector{
		PointedKeyReference: true, ExcludeColumns: []string{models.Workout_WorkoutDetails},
	})
	sel.Where().Column(models.Workout_Id, "IN", []int{id1, id2}).Add()
	sel.Where().Column(models.Workout_UserId, "=", a.R().User.Id).Add()
	if err := sel.Run(); err != nil {
		return err.GetResponse().Log("Failed to query workouts", err, a)
	}

	// Workouts must exist
	if len(workouts) != 2 {
		return ErrWorkoutNotFound
	}

	// Time Difference between end and start time
	diff := workouts[0].End.Unix() - workouts[1].Start.Unix()

	// Sort hierarchically
	if diff < 0 {
		diff *= -1
	} else {
		tmpWorkout := workouts[0]
		workouts[0] = workouts[1]
		workouts[1] = tmpWorkout
	}

	// Max allowed offset between workouts
	if diff > MergeAllowedTimeOffset {
		return ErrWorkoutsTooFarAway
	}

	// Merge workouts
	newWorkout := a.mergeHeaders(workouts[0], workouts[1])
	a.Logger().Info("Merging workout %d into %d", workouts[1].Id, workouts[0].Id)

	// Create transaction
	trans, err := a.R().Db.NewTransaction()
	if err != nil {
		return errors.InternalError().Log("Failed to create transaction", err, a)
	}

	// Update the first workout with the combined header
	selI := a.R().Db.Struct.Update(&newWorkout).Selector(database.ColumnSelector{
		PointedKeyReference: true, ExcludeColumns: []string{models.Workout_WorkoutDetails, models.Workout_CityLocation},
	})
	if err := selI.Run(); err != nil {
		trans.RollbackTransaction()
		return errors.InternalError().Log("Failed to update header of first workout (%d)", err, a, newWorkout.Id)
	}

	// Update all points of the second workout
	_, err = trans.Db.Exec(`
		UPDATE workout_details SET
			workout_id = ?,
			duration = duration + 1 + ?,
			distance = distance + ?
		WHERE workout_id = ?`,
		newWorkout.Id, workouts[0].Duration, workouts[0].Distance, workouts[1].Id,
	)
	if err != nil {
		trans.RollbackTransaction()
		return errors.InternalError().Log("Failed to modify workout_id", err, a)
	}

	// Remove the second workout
	_, err = trans.Db.Exec(`DELETE FROM workout where id = ?`, workouts[1].Id)
	if err != nil {
		trans.RollbackTransaction()
		return errors.InternalError().Log("Failed to delete second workout", err, a)
	}

	// Commit transaction
	if err := trans.CommitTransaction(); err != nil {
		return errors.InternalError().Log("Failed to commit transaction", err, a)
	}

	return nil
}

// mergeHeaders combines the second workout header with the first one.
// The ID of the first one will be kept
func (a *Api) mergeHeaders(w1 models.Workout, w2 models.Workout) models.Workout {
	w1.End = w2.End

	// Recalculate average heart rate. We need correct durations for this in w1
	if w1.HeartRateAv.Valid && w2.HeartRateAv.Valid {
		weighted1 := float64(w1.HeartRateAv.Int64) * float64(w1.Duration)
		weighted2 := float64(w2.HeartRateAv.Int64) * float64(w2.Duration)
		sum := float64(w1.Duration) + float64(w2.Duration)
		w1.HeartRateAv.Int64 = int64(math.Round((weighted1 + weighted2) / sum))
	}

	// Sum up simple values
	w1.Duration += w2.Duration
	w1.Calories += w2.Calories
	w1.CaloriesDefault += w2.CaloriesDefault
	w1.Distance += w2.Distance
	w1.ElevationUp += w2.ElevationUp
	w1.ElevationDown += w2.ElevationDown
	w1.Pai += w2.Pai
	w1.Steps.Int64 += w2.Steps.Int64
	w1.Steps.Valid = w1.Steps.Valid || w2.Steps.Valid

	// Append or replace note text
	if w2.Note.Valid {
		if w1.Note.Valid {
			w1.Note.String += "\n" + w2.Note.String
		} else {
			w1.Note = w2.Note
		}
	}

	// Apply maximum heart rate
	if w2.HeartRateMax.Int64 > w1.HeartRateMax.Int64 {
		w1.HeartRateMax = null.IntFrom(w2.HeartRateMax.Int64)
	}

	// Merge tags
outer:
	for _, t2 := range w2.WorkoutTags {
		// Check if tag is already contained
		for _, t1 := range w1.WorkoutTags {
			if t1.TagId.Id == t2.TagId.Id {
				continue outer
			}
		}

		t2.WorkoutId = w1.Id
		w1.WorkoutTags = append(w1.WorkoutTags, t2)
	}

	// Recalculate things
	speedAv := float64(w1.Duration) / (float64(w1.Distance) / 1000.0)
	w1.SpeedAv = int(math.Round(speedAv))

	return w1
}
