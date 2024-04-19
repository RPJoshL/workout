package statistics

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
			"StatisticView",
			"GET",
			"/",
			api.GetDashbaord,
			router.Options{},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

func (api *Api) GetDashbaord(w http.ResponseWriter, r *http.Request) {
	api.R().Tmpl.Render(api.main(), "generic.appName", "generic.appName")
}
