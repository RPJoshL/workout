package externalapi

import (
	extapi "git.rpjosh.de/RPJosh/workout/internal/externalapi"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/database"
	"git.rpjosh.de/RPJosh/workout/pkg/database/dbstruct"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

func (a *Api) upload(api extapi.API, workoutID int) errors.Error {
	config, err := a.getUserExternalAPI(api.Config().Type)
	if err != nil {
		return err
	}

	if config == nil {
		return errors.BadRequest("#settings.externalAPI.notConfigured")
	}

	var workout models.Workout
	sel := a.R().Db.Struct.Query(&workout)
	sel.Where().Column(models.Workout_UserId, "=", a.R().User.Id).Add()
	sel.Where().Column(models.Workout_Id, "=", workoutID).Add()
	sel.OrderBy("workout_details", models.WorkoutDetails_Duration, "ASC")
	if err := sel.Selector(dbstruct.ColumnSelector{PointedKeyReference: true, ForeignKeyReference: true}).Run(); err != nil {
		if err.Type() == database.NoRows {
			return errors.NotFound()
		}

		return err.GetResponse().Log("Failed to query workout", err.GetError(), a)
	}
	workout.FromDB()

	if err := api.UploadWorkout(&workout, config, a.markWorkoutUploaded); err != nil {
		return errors.BadRequest(err.Error())
	}

	return nil
}

func (a *Api) markWorkoutUploaded(w *models.Workout, field string) {
	sel := a.R().Db.Struct.Update(w)
	sel.Selector(dbstruct.ColumnSelector{
		IncludeColumns: []string{field},
	})

	if err := sel.Run(); err != nil {
		a.Logger().Warning("Failed to mark workout %d as uploaded", w.Id)
	}
}

func (a *Api) getUserExternalAPI(apiType models.ExternalAPIType) (*models.ExternalApi, errors.Error) {
	configs, err := a.GetExternalAPIs()
	if err != nil {
		return nil, err
	}

	return configs[apiType], nil
}
