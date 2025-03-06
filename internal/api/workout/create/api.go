package create

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/api/metric"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/api/utils"
	"git.rpjosh.de/RPJosh/workout/internal/api/workout/shared"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"git.rpjosh.de/RPJosh/workout/pkg/response"
	"github.com/a-h/templ"
)

type RootComponents interface {
	Main() (templ.Component, string)
}
type DetailsComponents interface {
	Details(id int) (templ.Component, string)
}

type Api struct {
	router.ApiRequest

	// Helper interface that renders main workout component
	// shared across different pages
	Root    RootComponents
	Details DetailsComponents

	Metric metric.Api

	Shared shared.Shared
}

var (
	ErrFileToLarge = errors.NewError("#workout.fileToLarge", 413)
	ErrFileRead    = errors.BadRequest("#workout.fileError")
)

func (api *Api) GetRouter() *router.Router {
	routes := router.Routes{
		router.NewRoute(
			"CreateWorkoutPage",
			"GET",
			"/new",
			api.CreateWorkoutPage,
			router.Options{},
		),
		router.NewRoute(
			"UpdateWorkoutPage",
			"GET",
			"/{id}/update",
			api.UpdateWorkoutPage,
			router.Options{},
		),
		router.NewRoute(
			"CreateWorkout",
			"POST",
			"/",
			api.CreateNewWorkout,
			router.Options{},
		),
		router.NewRoute(
			"CreateWorkoutApi",
			"POST",
			"/",
			api.CreateNewWorkoutApi,
			router.Options{IsApiEndpoint: true},
		),
		router.NewRoute(
			"MergeWorkout",
			"PUT",
			"/{id1}/merge/{id2}",
			api.MergeWorkoutsEndpoint,
			router.Options{},
		),
		router.NewRoute(
			"MergeWorkout",
			"PUT",
			"/{id1}/merge/{id2}",
			api.MergeWorkoutsEndpoint,
			router.Options{IsApiEndpoint: true},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

type WorkoutCreateUpdate struct {
	Name     string
	Type     int
	File     []byte
	FileName string
	Tags     []int
	Note     string
	City     string
}

func (api *Api) CreateWorkoutPage(w http.ResponseWriter, r *http.Request) {
	// Get workout data
	data, err := api.GetWorkoutNewEditData(-1)
	if err != nil {
		panic(err)
	}

	// Render page
	main, dep := api.Root.Main()
	api.R().Tmpl.RenderModal(
		api.workoutNewEdit(data), "workout.create",
		main, "/workout/", "generic.appName", "generic.appName", dep,
	)
}

func (api *Api) UpdateWorkoutPage(w http.ResponseWriter, r *http.Request) {
	// Get existing workout to edit
	editWorkout, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		panic(errors.BadRequest("Invalid workout id provided"))
	}

	// Get workout data
	data, err := api.GetWorkoutNewEditData(editWorkout)
	if err != nil {
		panic(err)
	}

	// The user can edit the workout directly from details view
	details, dep := api.Details.Details(editWorkout)
	if strings.HasSuffix(r.Header.Get("HX-Current-URL"), fmt.Sprintf("/workout/%d", editWorkout)) {
		api.R().Tmpl.RenderModal(
			api.workoutNewEdit(data), "workout.details",
			details, fmt.Sprintf("/workout/%d", editWorkout), "generic.appName", "generic.appName", dep,
		)
	} else {
		// Render page
		main, dep := api.Root.Main()
		api.R().Tmpl.RenderModal(
			api.workoutNewEdit(data), "workout.update",
			main, "/workout/", "generic.appName", "generic.appName", dep,
		)
	}

}

func (api *Api) CreateNewWorkout(w http.ResponseWriter, r *http.Request) {
	data := WorkoutCreateUpdate{}
	var err error

	// Parse body and get workout file
	if exit, workoutName, workoutFile := api.fetchWorkoutFile(w, r); exit {
		return
	} else {
		data.File = workoutFile
		data.FileName = workoutName
	}

	// Generic data
	data.Name = r.Form.Get("name")
	data.Note = r.Form.Get("note")
	data.City = r.Form.Get("city")
	activity := r.Form.Get("type")
	if activity != "" {
		if data.Type, err = strconv.Atoi(activity); err != nil {
			errors.BadRequest(api.R().Tr.Getf("generic.numericError", "type", activity)).Write(w, r)
			return
		}
	}

	// Tags has to be inspected manually (array)
	for i, t := range r.Form["tags"] {
		tagId, err := strconv.Atoi(t)
		if err != nil {
			errors.BadRequest(api.R().Tr.Getf("generic.numericError", fmt.Sprintf("tags[%d]", i), t)).Write(w, r)
			return
		}

		data.Tags = append(data.Tags, tagId)
	}

	// Get existing workout to update. Updating a workout is not solved in a Rest way!
	id := r.Form.Get("id")
	if id != "" && id != "0" {
		idInt, err := strconv.Atoi(id)
		if err != nil {
			errors.BadRequest(api.R().Tr.Getf("generic.numericError", "id", id)).Write(w, r)
			return
		}

		if err := api.UpdateWorkout(idInt, &data); err != nil {
			err.GetErrorStruct().Log("Failed to update workout %d", err, api, idInt).Write(w, r)
			return
		}

		response.WriteText("Workout updated", 200, w)
		return
	}

	// Create workout
	newWorkout, e := api.CreateWorkout(&data)
	if e != nil {
		e.GetErrorStruct().Log("Failed to create workout", e, api).Write(w, r)
		return
	}

	response.WriteText(fmt.Sprintf("%d", newWorkout.Id), 200, w)
}

// parseWorkoutFile parses the "multipart/form-data" body and tries to obtain
// an uploaded file.
// If [exit=true] is returned, you should not write any data to [r] anymore. Errors
// are already handled inside this function.
//
// If no file was found, the returned byte array is empty
func (api *Api) fetchWorkoutFile(w http.ResponseWriter, r *http.Request) (exit bool, filename string, fileContent []byte) {
	exit = true

	// Limit max file size to 5 Mbyte
	if r.ContentLength > utils.MToBytes(10) {
		// We need to parse the multipart data in order to
		logger.Info("Workout file is to big: %d Mbyte", r.ContentLength/1024/1024)
		ErrFileToLarge.Write(w, r)
		return
	}
	// Limit body size if content lenght is spoofed
	r.Body = http.MaxBytesReader(w, r.Body, utils.MToBytes(10))

	// Parse multipart form value
	if err := r.ParseMultipartForm(utils.MToBytes(2)); err != nil {
		response.WriteText(err.Error(), 400, w)
		return
	}

	// Read the provided file
	file, fileHeader, err := r.FormFile("file")
	if err == http.ErrMissingFile {
		return false, "", []byte{}
	} else if err != nil {
		logger.Warning("Failed to read workout file from request: %s", err)
		ErrFileRead.Write(w, r)
		return
	}
	defer file.Close()

	// Read file contents and parse XML
	fileContent, err = io.ReadAll(file)
	if err != nil {
		logger.Warning("Failed to read workout file from request: %s", err)
		ErrFileRead.Write(w, r)
		return
	}

	return false, fileHeader.Filename, fileContent
}

// MergeWorkout merges two separate workouts into a single one
func (api *Api) MergeWorkoutsEndpoint(w http.ResponseWriter, r *http.Request) {
	// Get workout IDs to merge
	id1, err := strconv.Atoi(r.PathValue("id1"))
	id2, err2 := strconv.Atoi(r.PathValue("id2"))
	if err != nil || err2 != nil {
		response.WriteError(errors.BadRequest("Invalid workout id provided"), w, r)
		return
	}

	if err := api.MergeWorkouts(id1, id2); err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	response.WriteText(api.R().Tr.Get("workout.mergedSuccess"), 200, w)
}

func (api *Api) CreateNewWorkoutApi(w http.ResponseWriter, r *http.Request) {

	// Parse body
	gpxFile := models.GpxFile{}
	if err := json.NewDecoder(r.Body).Decode(&gpxFile); err != nil {
		errors.BadRequest("").Log("Failed to decode body: %s", err, api).Write(w, r)
		return
	}

	if workout, err := api.CreateWorkoutByApi(gpxFile); err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		response.WriteJson(workout, 200, w)
	}

}
