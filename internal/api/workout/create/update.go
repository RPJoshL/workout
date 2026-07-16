package create

import (
	"math"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"git.rpjosh.de/RPJosh/workout/internal/api/workout/shared"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/parser"
	"git.rpjosh.de/RPJosh/workout/pkg/database"
	"git.rpjosh.de/RPJosh/workout/pkg/database/dbstruct"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"github.com/guregu/null/v5"
)

var (
	ErrFieldRequired      = errors.NewError("Field %q is required", 400)
	ErrWorkoutsTooFarAway = errors.NewError("#workout.tooFarAwayTime", 400)
	ErrWorkoutsSame       = errors.NewError("#workout.chooseNotSame", 400)

	// Maximum allowed duration in seconds how far two workouts can be
	// separated by each other
	MergeAllowedTimeOffset int64 = 16 * 60 * 60
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

	sel := a.R().Db.Struct.Update(&workout).Selector(dbstruct.ColumnSelector{IncludeColumns: FormUpdateFields, PointedKeyReference: true})
	if err := sel.Run(); err != nil {
		return err.GetResponse().Log("Failed to update workout", err, a)
	}

	return nil
}

// MergeWorkouts combines two separate workouts into a single one.
// The time between the workouts are counted as a break
func (a *Api) MergeWorkouts(id1, id2 int) errors.Error {
	// Workouts cannot be the same
	if id1 == id2 {
		return ErrWorkoutsSame
	}

	// Get workout headers
	workouts := []models.Workout{}
	sel := a.R().Db.Struct.QuerySlice(&workouts).Selector(dbstruct.ColumnSelector{
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
		workouts[0], workouts[1] = workouts[1], workouts[0]
	}

	// Max allowed offset between workouts
	if diff > MergeAllowedTimeOffset {
		return ErrWorkoutsTooFarAway
	}

	// Merge workouts
	newWorkout := a.mergeHeaders(workouts[0], &workouts[1])
	a.Logger().Info("Merging workout %d into %d", workouts[1].Id, workouts[0].Id)

	// Create transaction
	trans, err := a.R().Db.NewTransaction()
	if err != nil {
		return errors.InternalError().Log("Failed to create transaction", err, a)
	}

	// Update the first workout with the combined header
	selI := trans.Struct.Update(newWorkout).Selector(dbstruct.ColumnSelector{
		PointedKeyReference: true, ExcludeColumns: []string{models.Workout_WorkoutDetails, models.Workout_CityLocation},
	})
	if err := selI.NoTransaction().Run(); err != nil {
		trans.RollbackTransactionLog()
		return errors.InternalError().Log("Failed to update header of first workout (%d)", err, a, newWorkout.Id)
	}

	// Get the maximum part number of the first workout
	partId := 0
	if err := trans.QueryForValue(&partId, `SELECT MAX(part) FROM workout_details WHERE workout_id = ?`, newWorkout.Id); err != nil {
		return errors.InternalError().Log("Failed to select max part number for workout ID %d", err, a, newWorkout.Id)
	}

	// Update all points of the second workout
	_, err = trans.Db.Exec(`
		UPDATE workout_details SET
			workout_id = ?,
			duration = duration + 1 + ?,
			distance = distance + ?,
			part = ?
		WHERE workout_id = ?`,
		newWorkout.Id, workouts[0].Duration, workouts[0].Distance, partId+1, workouts[1].Id,
	)
	if err != nil {
		trans.RollbackTransactionLog()
		return errors.InternalError().Log("Failed to modify workout_id", err, a)
	}

	// Remove the second workout
	_, err = trans.Db.Exec(`DELETE FROM workout where id = ?`, workouts[1].Id)
	if err != nil {
		trans.RollbackTransactionLog()
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
//
//nolint:all
func (a *Api) mergeHeaders(w1 models.Workout, w2 *models.Workout) *models.Workout {
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

	return &w1
}

// DownsampleWorkout downsample the workout with the provided IDs to the provided level.
// This function is optimized for handling a large amount of workouts
func (a *Api) DownsampleWorkout(ids []int, level models.SamplingLevel) errors.Error {
	if level != models.SamplingLevelDownsampled {
		return errors.BadRequest("Unsupported downsampling interval")
	}

	// Get all IDs which belong to the user and are not downsampled yet
	workouts := []models.Workout{}
	sel := a.R().Db.Struct.QuerySlice(&workouts).Selector(dbstruct.ColumnSelector{IncludeColumns: []string{models.Workout_Id}})
	sel.Where().Column(models.Workout_Id, "IN", ids).Add()
	sel.Where().Column(models.Workout_UserId, "=", a.R().User.Id).Add()
	sel.Where().Column(models.Workout_SamplingLevel, "IN", []models.SamplingLevel{
		models.SamplingLevelDefault, models.SamplingLevelDetailed,
	}).Add()

	if err := sel.Run(); err != nil {
		return err.GetResponse().Log("Failed to query workouts for downsampling", err, a)
	}

	if len(workouts) == 0 {
		return ErrWorkoutNotFoundUnsampled
	}

	// Downsample a single workout directly
	if len(workouts) == 1 {
		if err := a.downsampleWorkout(workouts[0].Id); err != nil {
			return errors.InternalError().Log("Failed to downsample workout", err, a)
		}

		return nil
	}

	wg := sync.WaitGroup{}
	errs := atomic.Int32{}

	// Start workers
	workers := int(math.Min(float64(len(workouts)), 5))
	workerChan := make(chan *models.Workout, workers)

	for range workers {
		go a.downsamplingWorker(workerChan, &wg, &errs)
	}
	for i := range workouts {
		wg.Add(1)
		workerChan <- &workouts[i]
	}

	wg.Wait()
	close(workerChan)

	if len(workouts) == int(errs.Load()) {
		return errors.InternalError().Log("Failed to downsample all workout", nil, a)
	} else if errs.Load() > 0 {
		return errors.NewError(a.R().Tr.Sprintf("workout.downsamplingFailedPartial", len(workouts)-int(errs.Load()), len(workouts)), http.StatusInternalServerError)
	}

	return nil
}

func (a *Api) downsamplingWorker(workerChan chan *models.Workout, wg *sync.WaitGroup, errs *atomic.Int32) {
	for {
		workout, ok := <-workerChan
		if !ok {
			return
		}

		if err := a.downsampleWorkout(workout.Id); err != nil {
			a.Logger().Error("Failed to downsample workout %d: %s", workout.Id, err)
			errs.Add(1)
		}

		wg.Done()
	}
}

func (a *Api) downsampleWorkout(id int) error {
	details := []models.WorkoutDetails{}
	sel := a.R().Db.Struct.QuerySlice(&details)
	sel.Where().Column(models.WorkoutDetails_WorkoutId, "=", id).Add()

	if err := sel.Run(); err != nil {
		return errors.Wrap(err.GetError(), "query workout details")
	}

	downsampled := a.Shared.DownsamplePoints(&models.Workout{WorkoutDetails: details}, 26, shared.DownSampleConstraints{
		MaxDuration:      30,
		ConstraintDriven: true,
	})

	if len(downsampled) == len(details) {
		return nil
	}
	if len(downsampled) > len(details) {
		return errors.New("got more downsampled points than original points")
	}

	trans, err := a.R().Db.NewTransaction()
	if err != nil {
		return errors.Wrap(err, "create transaction")
	}

	// Get all workout points that were not selected for downsampling.
	// We can do this as the original points were not modified
	ids := make([]any, 0, len(details)-len(downsampled))

outer:
	for _, d := range details {
		for _, ds := range downsampled {
			if d.Id == ds.Id {
				continue outer
			}
		}

		ids = append(ids, d.Id)
	}

	if len(ids) == 0 {
		return nil
	}

	delStatement := `DELETE FROM workout_details WHERE id IN (?` + strings.Repeat(",?", len(ids)-1) + `)`
	if _, err := trans.Db.Exec(delStatement, ids...); err != nil {
		return errors.Wrap(err, "delete old workout details")
	}

	updStatement := trans.Struct.Update(&models.Workout{
		Id:            id,
		SamplingLevel: int(models.SamplingLevelDownsampled),
	}).Selector(dbstruct.ColumnSelector{IncludeColumns: []string{models.Workout_SamplingLevel}})
	if err := updStatement.NoTransaction().Run(); err != nil {
		return errors.Wrap(err, "update workout sampling level")
	}

	if err := trans.CommitTransaction(); err != nil {
		return errors.Wrap(err, "commit transaction")
	}

	return nil
}

func (a *Api) reprocessWorkout(id int) errors.Error {
	workout := models.Workout{}
	sel := a.R().Db.Struct.Query(&workout)
	sel.Where().Column(models.Workout_Id, "=", id).Add()
	sel.Where().Column(models.Workout_UserId, "=", a.R().User.Id).Add()
	sel.Selector(dbstruct.ColumnSelector{
		PointedKeyReference: true,
	})

	if err := sel.Run(); err != nil {
		if err.Type() == database.NoRows {
			return errors.NotFound()
		}

		return errors.InternalError().Log("Failed to query workout", err, a)
	}

	postProcessor := parser.NewPostProcessor(parser.PostProcessingOptions{
		// We expect that the user knows what he is doing and expects reprocessing
		UseSpeedDeviceData: false,
	})
	postProcessor.PostProcess(&workout)

	trans, err := a.R().Db.NewTransaction()
	if err != nil {
		return errors.InternalError().Log("Failed to start transaction", err, a)
	}

	stmt := "DELETE wd, wm FROM workout_details wd LEFT JOIN workout_metric wm ON wm.workout_id = wd.workout_id WHERE wd.workout_id = ?"
	if _, err := trans.Db.GetDb().Exec(stmt, workout.Id); err != nil {
		trans.RollbackTransactionLog()
		return errors.InternalError().Log("Failed to delete existing workout details", err, a)
	}

	workout.ToDB()

	if _, err := trans.Struct.InsertSlice(&workout.WorkoutMetric).Run(); err != nil {
		trans.RollbackTransactionLog()
		return errors.InternalError().Log("Failed to insert workout metrics", err, a)
	}

	if _, err := trans.Struct.InsertSlice(&workout.WorkoutDetails).Run(); err != nil {
		trans.RollbackTransactionLog()
		return errors.InternalError().Log("Failed to update workout after reprocessing", err, a)
	}

	if err := trans.CommitTransaction(); err != nil {
		trans.RollbackTransactionLog()
		return errors.InternalError().Log("Failed to commit transaction", err, a)
	}

	return nil
}
