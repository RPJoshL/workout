package workout

import (
	"net/http"

	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/api/workout/create"
)

type Api struct {
	router.ApiRequest

	Create *create.Api
}

func GetRoutes() *router.Router {
	api := &Api{}

	api.Create = &create.Api{
		Root: api,
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
	api.R().Tmpl.Render(api.Main(), "generic.appName", "generic.appName")
}

func (api *Api) GetWorkoutDetails(w http.ResponseWriter, r *http.Request) {
	//api.R().Tmpl.RenderModal(
	//	api.workout(), "workout.create",
	//	api.main(), "/workout/", "generic.appName", "generic.appName",
	//)
}
