// Package externalapi handles the communication with external API
// services that can be used to upload workouts
package externalapi

import (
	"net/http"

	"git.rpjosh.de/RPJosh/workout/internal/api/components/form"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"github.com/a-h/templ"
)

type AuthenticationUI struct {
	R         *router.Request
	BuildForm func(fields []form.Field, buttonTranslation string, stepIdx int) templ.Component
}

type AuthenticationStepResult struct {
	// When set, indicates that the authentication process was successful
	// and the user is authenticated
	Config *models.ExternalApi
	// UI component for the next authentication step
	NextComponent templ.Component
}

type Configuration struct {
	// The default server URL to use when the user didn't specify one
	DefaultURL string
	// Internal (public) URL of the logo
	IconURL string
	// Display name of the service
	Name string
	Type models.ExternalAPIType
	// Authentication steps sorted by the flow of the authentication.
	// The order is important
	AuthenticationSteps []AuthenticationStep
	// Workout types that are supported by this API. Return an empty slice
	// if there is no restriction
	WorkoutTypes []int
}

type UploadMarker func(w *models.Workout, fieldToUpdate string)

// API defines common methods used for external API
type API interface {
	Config() *Configuration
	UploadWorkout(w *models.Workout, apiConfig *models.ExternalApi, marker UploadMarker) error
	// IsAlreadyUploaded returns true if the workout is already uploaded to the remote service.
	// It should return a link to open the workout on the remote service frontend
	IsAlreadyUploaded(w *models.Workout) (bool, string)
}

type AuthenticationStep interface {
	// Fields to display in the UI for this step
	Fields() []form.Field
	// Execute this authentication step
	Execute(r *http.Request, ui AuthenticationUI) (*AuthenticationStepResult, error)
}
