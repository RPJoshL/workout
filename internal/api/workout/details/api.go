package details

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/api/utils"
	"git.rpjosh.de/RPJosh/workout/internal/api/workout/shared"
	"git.rpjosh.de/RPJosh/workout/internal/converter"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"git.rpjosh.de/RPJosh/workout/pkg/response"
	"github.com/a-h/templ"
)

var validExportFormats = []string{
	"gpx", "tcx",
}

type RootComponents interface {
	Main() (templ.Component, string)
}

type Api struct {
	router.ApiRequest

	// Helper interface that renders main workout component
	// shared across different pages
	Root RootComponents

	Shared shared.Shared
}

func (api *Api) GetRouter() *router.Router {
	routes := router.Routes{
		router.NewRoute(
			"GetWorkout",
			"GET",
			"/{id}",
			api.GetWorkoutDetails,
			router.Options{},
		),
		router.NewRoute(
			"PathWorkoutDetails",
			"PATCH",
			"/{id}",
			api.PatchWorkoutDetails,
			router.Options{},
		),
		router.NewRoute(
			"ExportWorkout",
			"GET",
			"/{id}/export/{format}",
			api.ExportWorkout,
			router.Options{},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

func (api *Api) Details(id int) (comp templ.Component, path string) {
	// Get data to render
	data, e := api.GetWorkoutDetailsData(id)
	if e != nil {
		panic(e)
	}

	return api.WorkoutView(data), utils.GetCallerFile()
}

func (api *Api) GetWorkoutDetails(w http.ResponseWriter, r *http.Request) {
	// Get ID of workout to display
	workoutId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errors.BadRequest("#generic.numericError").Sprintf("id", r.PathValue("id")).Write(w, r)
		return
	}

	// Get data to display
	data, e := api.GetWorkoutDetailsData(workoutId)
	if e != nil {
		e.GetErrorStruct().Write(w, r)
		return
	}

	main, dep := api.Root.Main()
	api.R().Tmpl.RenderModal(
		api.WorkoutView(data), "workout.details",
		main, "/workout/", "generic.appName", "generic.appName", dep,
	)
}

func (api *Api) PatchWorkoutDetails(w http.ResponseWriter, r *http.Request) {
	// Get ID of workout to update
	workoutId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errors.BadRequest("#generic.numericError").Sprintf("id", r.PathValue("id")).Write(w, r)
		return
	}

	body := &WorkoutDetailsPatch{}
	if err := api.R().Parser.Parse(body, router.RequestParserOptions{
		InterpreteJson: true,
		Mode:           router.ParseModeForm,
	}); err != nil {
		errors.BadRequest("").Log("Failed to decody workout path body", err, api).Write(w, r)
		return
	}

	if body.Latitude == 0 || body.Longitude == 0 {
		errors.BadRequest("Invalid latitude or longitude provided").Write(w, r)
		return
	}

	// Update workout details
	if err := api.PatchWorkoutLocation(workoutId, body.Latitude, body.Longitude); err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	// Return updated details page
	details, _ := api.Details(workoutId)
	main, dep := api.Root.Main()
	api.R().Tmpl.RenderModal(
		details, "workout.details",
		main, "/workout/", "generic.appName", "generic.appName", dep,
	)
}

func (api *Api) ExportWorkout(w http.ResponseWriter, r *http.Request) {
	workoutId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errors.BadRequest("#generic.numericError").Sprintf("id", r.PathValue("id")).Write(w, r)
		return
	}

	// Get format to export
	format := r.PathValue("format")
	if !slices.Contains(validExportFormats, format) {
		errors.BadRequest("Invalid export format provided. Supported are: "+strings.Join(validExportFormats, ",")).Write(w, r)
		return
	}

	// Get workout details
	workout, errApi := api.getWorkoutData(workoutId)
	if errApi != nil {
		errApi.GetErrorStruct().Write(w, r)
		return
	}

	var fileContent []byte
	var contentType string
	switch format {
	case "gpx":
		fileContent, err = converter.ToGPX(workout)
		contentType = "application/gpx+xml"
	case "tcx":
		fileContent, err = converter.ToTCX(workout)
		contentType = "application/vnd.garmin.tcx+xml"
	default:
		api.Logger().Warning("Unmapped file format provided: %s", format)
		errors.InternalError().Write(w, r)
		return
	}

	if err != nil {
		errors.InternalError().Log("Failed to convert workout to file for export", err, api).Write(w, r)
		return
	}

	fileName := fmt.Sprintf("workout_%d_%s.%s", workout.Id, workout.Start.Format(time.DateOnly), format)
	response.DonwloadableFile(w, fileName, contentType, fileContent)
}
