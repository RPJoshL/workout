package cities

import (
	"net/http"

	"git.rpjosh.de/RPJosh/workout/internal/api/router"
)

type Api struct {
	router.ApiRequest
}

func (a *Api) GetRouter() *router.Router {
	routes := router.Routes{
		router.NewRoute(
			"WorkoutCitiesSelect",
			"GET",
			"/city",
			a.GetWorkoutCitysForSelect,
			router.Options{},
		),
	}

	return &router.Router{
		Dependency: a,
		Routes:     routes,
	}
}

func (a *Api) GetWorkoutCitysForSelect(w http.ResponseWriter, r *http.Request) {
	// Get city name to filter for
	input := r.URL.Query().Get("input")

	comp, err := a.GetCityOptions(input)
	if err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		a.R().Tmpl.RenderDirect(comp)
	}
}
