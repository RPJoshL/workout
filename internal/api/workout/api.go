package workout

import (
	"net/http"

	"git.rpjosh.de/RPJosh/workout/internal/api/router"
)

type Api struct {
	router.ApiRequest
}

func GetRoutes() *router.Router {
	api := &Api{}

	routes := router.Routes{
		router.NewRoute(
			"WorkoutTable",
			"GET",
			"/",
			api.GetWorkoutTablePage,
			router.Options{},
		),
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

func (api *Api) GetWorkoutTablePage(w http.ResponseWriter, r *http.Request) {
	api.R().Tmpl.Render(api.main(), "generic.appName", "generic.appName")
}

func (api *Api) GetWorkoutDetails(w http.ResponseWriter, r *http.Request) {
	api.R().Tmpl.RenderModal(
		api.workout(), "workout.new",
		api.main(), "/workout/", "generic.appName", "generic.appName",
	)
}
