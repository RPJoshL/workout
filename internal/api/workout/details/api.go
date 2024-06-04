package details

import (
	"net/http"
	"strconv"

	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/api/utils"
	"git.rpjosh.de/RPJosh/workout/internal/api/workout/shared"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"github.com/a-h/templ"
)

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
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

func (api *Api) Details(id int) (templ.Component, string) {
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
