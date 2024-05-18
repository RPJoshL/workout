package create

import (
	"git.rpjosh.de/RPJosh/go-webserver/errors"
	"git.rpjosh.de/RPJosh/workout/internal/converter"
	"git.rpjosh.de/RPJosh/workout/internal/database"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/parser"
)

var (
	ErrWorkoutNotFound = errors.NewError("#workout.notFound", 404)
	ErrFileFormat      = errors.BadRequest("#workout.gpxError")
	ErrTagsNotFound    = errors.NewError("#workout.tagsNotFound", 404)
	ErrTypeNotFound    = errors.NewError("#workout.typeNotFound", 404)
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
		return nil, ErrFileFormat.Log("Failed to parse workout file", err, a)
	}

	// Downsample
	workout, e := parser.Workout(gpxData, a.R().User.User, a.R().Db)
	if e != nil {
		return nil, e.GetErrorStruct().Log("Failed to downsample workout / parse workout file", e, a)
	}

	// Overwrite file if provided
	if data.Name != "" {
		workout.Name = data.Name
	} else if workout.Name == "" {
		// No name specified. Use the translated name of the workout type
		workoutType := models.WorkoutType{}
		a.R().Db.Struct.Query(&workoutType).Where().Column(models.WorkoutType_Id, "=", workout.TypeId).Add().Run()
		workout.Name = a.Shared.GetWorkoutTypeName(workoutType)
	}

	// Add tags and type
	if data.Type > 0 {
		workout.TypeId = data.Type
	} else if workout.TypeId == models.TYPE_UNKNOWN {
		workout.TypeId = models.TYPE_HIKING
	}
	if data.Note != "" {
		workout.Note = database.NewNullString(data.Note)
	}
	workout.WorkoutTags = workoutTags

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
