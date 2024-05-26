package workout

import (
	"fmt"
	"net/http"
	"strconv"

	"git.rpjosh.de/RPJosh/go-webserver/errors"
	"git.rpjosh.de/RPJosh/go-webserver/response"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/api/workout/create"
	"git.rpjosh.de/RPJosh/workout/internal/database"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"github.com/a-h/templ"
)

type Api struct {
	router.ApiRequest

	Create *create.Api
	Types  *[]models.WorkoutType
}

func GetRoutes(db *database.DatabaseUtils) *router.Router {
	api := &Api{
		Types: &[]models.WorkoutType{},
	}

	api.Create = &create.Api{
		Root: api,
	}

	// Get workout types from the database once at startup
	if err := db.Struct.QuerySlice(api.Types).Run(); err != nil {
		panic(fmt.Sprintf("Failed to query workout types from db: %s", err))
	}

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
			api.Create.CreateWorkoutPage,
			router.Options{},
		),
		router.NewRoute(
			"UpdateWorkoutPage",
			"GET",
			"/{id}/update",
			api.Create.UpdateWorkoutPage,
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
			api.Create.CreateNewWorkout,
			router.Options{},
		),
		router.NewRoute(
			"DeleteWorkout",
			"DELETE",
			"/{id}",
			api.DeleteWorkout,
			router.Options{},
		),
	}

	rout := &router.Router{
		Dependency: api,
		Routes:     routes,
	}

	// Add (sub) routers
	rout.AddRouter(api.Create.GetRouter())

	return rout
}

func (api *Api) GetWorkoutTablePage(w http.ResponseWriter, r *http.Request) {

	// Get data to display
	data, e := api.GetTableData()
	if e != nil {
		e.GetErrorStruct().Write(w, r)
		return
	}

	api.R().Tmpl.Render(api.MainWithData(data), "generic.appName", "generic.appName")
}

func (api *Api) Main() templ.Component {
	return api.MainWithData(&TableData{})
}
func (api *Api) Details(id int) templ.Component {
	// Get data to render
	data, e := api.GetWorkoutDetailsData(id)
	if e != nil {
		panic(e)
	}

	return api.WorkoutView(data)
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

	api.R().Tmpl.RenderModal(
		api.WorkoutView(data), "workout.details",
		api.Main(), "/workout/", "generic.appName", "generic.appName",
	)
}

func (api *Api) DeleteWorkout(w http.ResponseWriter, r *http.Request) {

	// Get ID of workout to display
	workoutId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errors.BadRequest("#generic.numericError").Sprintf("id", r.PathValue("id")).Write(w, r)
		return
	}

	// Delete workout
	if err := api.Delete(workoutId); err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		response.WriteText("Workout deleted", 200, w)
	}
}
