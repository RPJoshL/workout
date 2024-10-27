package create

import (
	"fmt"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/converter"
	"git.rpjosh.de/RPJosh/workout/internal/database"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/parser"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"github.com/guregu/null/v5"
)

var (
	ErrWorkoutNotFound = errors.NewError("#workout.notFound", 404)
	ErrTagsNotFound    = errors.NewError("#workout.tagsNotFound", 404)
	ErrTypeNotFound    = errors.NewError("#workout.typeNotFound", 404)
	ErrWorkoutExists   = errors.NewError("#workout.similarExists", 409)
)

func (a *Api) GetWorkoutNewEditData(existingWorkout int) (work *workoutNewEditData, e errors.Error) {
	rtc := &workoutNewEditData{}

	// Query existing workout data
	if existingWorkout > 0 {
		if rtc.existingWorkout, e = a.getExistingWorkout(existingWorkout); e != nil {
			return nil, e
		}
	}

	// Query available types and tags
	if err := a.R().Db.Struct.QuerySlice(&rtc.workoutTypes).Run(); err != nil {
		return nil, err.GetResponse().Log("Failed to query workout typs", err, a)
	}
	if err := a.R().Db.Struct.QuerySlice(&rtc.tags).Run(); err != nil {
		return nil, err.GetResponse().Log("Failed to query workout tags", err, a)
	}

	return rtc, nil
}

// CreateWorkoutByApi creates a new workout by the provided GPX file
// and returns the header of the created workout
func (a *Api) CreateWorkoutByApi(file models.GpxFile) (rtc *models.Workout, rtcE errors.Error) {

	if file.Type != 0 {
		// Validate type
		if err := a.validateType(file.Type); err != nil {
			return nil, err
		}
	} else if file.TypeName != "" {
		// Get workout type by name
		file.Type = models.GetWorkoutTypeByName(file.TypeName)
	}

	// Get PAI score of last week
	startDate := file.Points[0].Timestamp
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day()+1, 0, 0, 0, 0, startDate.Location())
	paiScoreWeek, errPai := a.Metric.GetSumOfPai(startDate, startDate.AddDate(0, 0, -7))
	if errPai != nil {
		return nil, errPai.GetResponse().Log("Failed to get PAI sum", errPai, a)
	}

	// Downsample
	workout, e := parser.Workout(&file, a.R().User.User, a.R().Db, paiScoreWeek)
	if e != nil {
		return nil, e.GetErrorStruct().Log("Failed to downsample workout / parse workout file", e, a)
	}

	// Check if workout already exists
	if exists, err := a.getDuplicates(workout); err != nil {
		return nil, err
	} else if len(exists) > 0 {
		return nil, ErrWorkoutExists.WithHeader("Existing-Workout-Id", fmt.Sprintf("%d", exists[0].Id))
	}

	// Set default properties
	workout.Name = a.getTypeName(workout.TypeId)

	// Create the workout in database
	selector := database.ColumnSelector{PointedKeyReference: true}
	if id, ee := a.R().Db.Struct.Insert(workout).Selector(selector).Run(); ee != nil {
		return nil, ee.GetResponse().Log("Failed to insert workout", ee, a)
	} else {
		rtc = &models.Workout{}
		q := a.R().Db.Struct.Query(rtc).Where().Column(models.Workout_Id, "=", id).Add()
		if e := q.Run(); e != nil {
			return nil, errors.InternalError().Log("Failed to query workout", e, a)
		}
		return
	}
}

// CreateWorkout creates a new workout and returns it if no
// error occured during processing the workout data
func (a *Api) CreateWorkout(data *WorkoutCreateUpdate) (*models.Workout, errors.Error) {

	// Get and validate tags
	workoutTags, e := a.validateTags(data.Tags)
	if e != nil {
		return nil, e
	}

	// Validate type
	if err := a.validateType(data.Type); err != nil {
		return nil, err
	}

	// Parse provided workout file
	gpxData, err := converter.ParseWorkoutFile(data.FileName, data.File)
	if err != nil {
		return nil, err.GetErrorStruct().Log("Failed to parse workout file", err, a)
	}

	// Get PAI score of last week
	startDate := gpxData.Points[0].Timestamp
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day()+1, 0, 0, 0, 0, startDate.Location())
	paiScoreWeek, errPai := a.Metric.GetSumOfPai(startDate, startDate.AddDate(0, 0, -7))
	if errPai != nil {
		return nil, errPai.GetResponse().Log("Failed to get PAI sum", errPai, a)
	}

	// Downsample
	workout, e := parser.Workout(gpxData, a.R().User.User, a.R().Db, paiScoreWeek)
	if e != nil {
		return nil, e.GetErrorStruct().Log("Failed to downsample workout / parse workout file", e, a)
	}

	// Overwrite file if provided
	if data.Name != "" {
		workout.Name = data.Name
	} else if workout.Name == "" {
		workout.Name = a.getTypeName(workout.TypeId)
	}
	// Overwrite City
	if data.City != "" {
		workout.City = data.City
	}

	// Add tags and type
	if data.Type > 0 {
		workout.TypeId = data.Type
	} else if workout.TypeId == models.TYPE_UNKNOWN {
		workout.TypeId = models.TYPE_HIKING
	}
	if data.Note != "" {
		workout.Note = null.StringFrom(data.Note)
	}
	workout.WorkoutTags = workoutTags

	// Check if workout already exists
	if exists, err := a.getDuplicates(workout); err != nil {
		return nil, err
	} else if len(exists) > 0 {
		return nil, ErrWorkoutExists
	}

	// Create the workout in database
	selector := database.ColumnSelector{PointedKeyReference: true}
	if id, ee := a.R().Db.Struct.Insert(workout).Selector(selector).Run(); ee != nil {
		return nil, ee.GetResponse().Log("Failed to insert workout", ee, a)
	} else {
		return &models.Workout{Id: int(id)}, nil
	}
}

// validateTags checks if all tags exist within the database and returns the transformed
// tags for the database
func (a *Api) validateTags(tagsI []int) (tags []models.WorkoutTags, err errors.Error) {
	// Nothing to do
	if len(tagsI) == 0 {
		return
	}

	tmp := models.Tag{}
	sel := a.R().Db.Struct.Query(&tmp).Where().Column(models.Tag_Id, "IN", tagsI).Add()

	c, e := sel.Count()
	if e != nil {
		return tags, e.GetResponse().Log("Failed to count tags", e.GetError(), a)
	}

	// We need to find all tags
	if len(tagsI) != c {
		return tags, ErrTagsNotFound
	}

	// Parse and validate tags into workout tags
	for _, tag := range tagsI {
		tags = append(tags, models.WorkoutTags{
			TagId: &models.Tag{Id: tag},
		})
	}

	return
}

// validateType validates the workout type
func (a *Api) validateType(typ int) (err errors.Error) {
	// Nothing to do
	if typ == 0 {
		return
	}

	tmp := models.WorkoutType{}
	sel := a.R().Db.Struct.Query(&tmp).Where().Column(models.WorkoutType_Id, "=", typ).Add()

	c, e := sel.Count()
	if e != nil {
		return e.GetResponse().Log("Failed to get existing type", e.GetError(), a)
	}

	// We need to find the type
	if c != 1 {
		return ErrTypeNotFound
	}

	return
}

// getExistingWorkout returns an existing workout without the workout details
func (a *Api) getExistingWorkout(id int) (workout models.Workout, err errors.Error) {
	sel := a.R().Db.Struct.Query(&workout)
	sel.Where().Column(models.Workout_Id, "=", id).Add()
	sel.Where().Column(models.User_Id, "=", a.R().User.Id)
	dbError := sel.Selector(database.ColumnSelector{
		// We don't need any workout details in edit dialog
		ExcludeColumns:      []string{models.Workout_WorkoutDetails},
		PointedKeyReference: true,
	}).Run()

	if dbError != nil {
		if dbError.Type() == database.NoRows {
			return workout, ErrWorkoutNotFound
		} else {
			return workout, dbError.GetResponse().Log("Failed to query existing workout", err, a)
		}
	}

	return
}

// getDuplicates checks weather this workout is already stored in
// the db with similar values and returns these similar workouts
func (a *Api) getDuplicates(workout *models.Workout) (existingworkouts []models.Workout, err errors.Error) {
	// Try to select workout with the same start / end time
	sel := a.R().Db.Struct.QuerySlice(&existingworkouts)
	sel.Where().Column(models.Workout_UserId, "=", workout.UserId).Add()
	sel.Where().Column(models.Workout_Start, ">=", workout.Start.Add(-2*time.Minute)).Add()
	sel.Where().Column(models.Workout_Start, "<=", workout.End.Add(2*time.Minute)).Add()

	if err := sel.Run(); err != nil {
		return []models.Workout{}, errors.InternalError().Log("Faield to count existing workout", err, a)
	} else {
		return existingworkouts, nil
	}
}

// getTypeName returns the translated name for the provided type
func (a *Api) getTypeName(typeId int) string {
	workoutType := models.WorkoutType{}
	if errD := a.R().Db.Struct.Query(&workoutType).Where().Column(models.WorkoutType_Id, "=", typeId).Add().Run(); errD != nil {
		logger.Warning("Failed to fetch details of workout type %d", typeId)
	}
	return a.Shared.GetWorkoutTypeName(workoutType)
}
