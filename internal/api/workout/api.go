package workout

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/go-webserver/errors"
	"git.rpjosh.de/RPJosh/go-webserver/response"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/api/utils"
)

var (
	ErrFileToLarge = errors.NewError("#workout.fileToLarge", 413)
	ErrFileRead    = errors.BadRequest("#workout.fileError")
	ErrFileFormat  = errors.BadRequest("#workout.gpxError")
)

type Api struct {
	router.ApiRequest
}

func GetRoutes() *router.Router {
	api := &Api{}

	routes := router.Routes{

		// Pages
		router.NewRoute(
			"WorkoutTablePage",
			"GET",
			"/",
			api.GetWorkoutTablePage,
			router.Options{},
		),
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
			"GetWorkout",
			"GET",
			"/{id}",
			api.GetWorkoutDetails,
			router.Options{},
		),
		router.NewRoute(
			"CreateWorkout",
			"POST",
			"/",
			api.CreateNewWorkout,
			router.Options{},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

func (api *Api) GetWorkoutTablePage(w http.ResponseWriter, r *http.Request) {
	api.R().Tmpl.Render(api.main(), "generic.appName", "generic.appName")
}

func (api *Api) GetWorkoutDetails(w http.ResponseWriter, r *http.Request) {
	//api.R().Tmpl.RenderModal(
	//	api.workout(), "workout.create",
	//	api.main(), "/workout/", "generic.appName", "generic.appName",
	//)
}

type WorkoutCreate struct {
	Name int    `request:"name"`
	File []byte `request:"file"`
}

// our struct which contains the complete
// array of all Users in the file
type Users struct {
	XMLName xml.Name `xml:"users"`
	Users   []User   `xml:"user"`
}

// the user struct, this contains our
// Type attribute, our user's name and
// a social struct which will contain all
// our social links
type User struct {
	XMLName xml.Name `xml:"user"`
	Type    string   `xml:"type,attr"`
	Name    string   `xml:"name"`
	Social  Social   `xml:"social"`
}

// a simple struct which contains all our
// social links
type Social struct {
	XMLName  xml.Name `xml:"social"`
	Facebook string   `xml:"facebook"`
	Twitter  string   `xml:"twitter"`
	Youtube  string   `xml:"youtube"`
}

func (api *Api) CreateWorkoutPage(w http.ResponseWriter, r *http.Request) {
	// Get workout data
	data, err := api.GetWorkoutNewEditData(-1)
	if err != nil {
		panic(err)
	}

	// Render page
	api.R().Tmpl.RenderModal(
		api.workoutNewEdit(data), "workout.create",
		api.main(), "/workout/", "generic.appName", "generic.appName",
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

	// Render page
	api.R().Tmpl.RenderModal(
		api.workoutNewEdit(data), "workout.update",
		api.main(), "/workout/", "generic.appName", "generic.appName",
	)
}

func (api *Api) CreateNewWorkout(w http.ResponseWriter, r *http.Request) {

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
	file, _, err := r.FormFile("file")
	if err != nil {
		logger.Warning("Failed to read workout file from request: %s", err)
		ErrFileRead.Write(w, r)
		return
	}
	defer file.Close()

	// Read file contents and parse XML
	bb, err := io.ReadAll(file)
	if err != nil {
		logger.Warning("Failed to read workout file from request: %s", err)
		ErrFileRead.Write(w, r)
		return
	}
	var workoutFile Users
	if err := xml.Unmarshal(bb, &workoutFile); err != nil {
		logger.Warning("Failed to parse workout file: %s", err)
		ErrFileFormat.Write(w, r)
		return
	}

	// We have to inspect tags manually
	tags := make([]int, 0)
	for i, t := range r.Form["tags"] {
		tagId, err := strconv.Atoi(t)
		if err != nil {
			errors.BadRequest(fmt.Sprintf("Invalid numeric value provided for tags[%d]: %q", i, t)).Write(w, r)
			return
		}

		tags = append(tags, tagId)
		logger.Debug("Received tag %d", tagId)
	}

	logger.Debug("Received workout %s with first user %q", r.FormValue("name"), workoutFile.Users[0].Name)

	response.WriteText("Workout created", 200, w)
}
