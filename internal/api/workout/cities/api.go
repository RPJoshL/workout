package cities

import (
	"net/http"

	"git.rpjosh.de/RPJosh/workout/internal/api/router"
)

type Api struct {
	router.ApiRequest
}

func (api *Api) GetRouter() *router.Router {

	routes := router.Routes{
		router.NewRoute(
			"WorkoutCitiesSelect",
			"GET",
			"/city",
			api.GetWorkoutCitysForSelect,
			router.Options{},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

func (api *Api) GetWorkoutCitysForSelect(w http.ResponseWriter, r *http.Request) {

	// Get city name to filter for
	input := r.URL.Query().Get("input")

	comp, err := api.GetCityOptions(input)
	if err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		api.R().Tmpl.RenderDirect(comp)
	}
}
