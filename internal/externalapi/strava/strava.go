// Package strava handles the communication with the Strava API
package strava

import (
	"fmt"
	"net/http"

	"git.rpjosh.de/RPJosh/workout/internal/api/components/form"
	"git.rpjosh.de/RPJosh/workout/internal/externalapi"
	"git.rpjosh.de/RPJosh/workout/internal/models"
)

const defaultURL = "https://www.strava.com"

var _ externalapi.API = (*Api)(nil)

type Api struct{}

func (a *Api) Config() *externalapi.Configuration {
	return &externalapi.Configuration{
		DefaultURL: defaultURL,
		Name:       "Strava",
		IconURL:    "/static/img/logos/strava_icon.png",
		Type:       models.ExternalAPIStrave,
		AuthenticationSteps: []externalapi.AuthenticationStep{
			&notImplemented{},
		},
		WorkoutTypes: []int{},
	}
}

func (a *Api) UploadWorkout(workout *models.Workout, apiConfig *models.ExternalApi, marker externalapi.UploadMarker) error {
	return fmt.Errorf("strava upload is not implemented yet")
}

func (a *Api) IsAlreadyUploaded(workout *models.Workout) (uploaded bool, href string) {
	return false, ""
}

type notImplemented struct{}

func (a *notImplemented) Fields() []form.Field {
	return []form.Field{}
}

func (a *notImplemented) Execute(r *http.Request, ui externalapi.AuthenticationUI) (*externalapi.AuthenticationStepResult, error) {
	return nil, fmt.Errorf("strava authentication is not implemented yet")
}
