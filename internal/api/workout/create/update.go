package create

import (
	"database/sql"

	"git.rpjosh.de/RPJosh/go-webserver/errors"
	"git.rpjosh.de/RPJosh/workout/internal/database"
	"git.rpjosh.de/RPJosh/workout/internal/models"
)

var (
	ErrFieldRequired = errors.NewError("Field %q is required", 400)
)

// Fields of a workout that can be updated from the form
var FormUpdateFields = []string{
	models.Workout_TypeId,
	models.Workout_WorkoutTags,
	models.Workout_Name,
	models.Workout_Note,
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
	_, err := a.getExistingWorkout(id)
	if err != nil {
		return err
	}

	// Patch workout
	workout := models.Workout{
		Id:          id,
		TypeId:      data.Type,
		WorkoutTags: workoutTags,
		Name:        data.Name,
		Note:        sql.NullString{String: data.Note, Valid: data.Note != ""},
	}
	sel := a.R().Db.Struct.Update(&workout).Selector(database.ColumnSelector{IncludeColumns: FormUpdateFields, PointedKeyReference: true})
	if err := sel.Run(); err != nil {
		return err.GetResponse().Log("Failed to update workout", err, a)
	}

	return nil
}
