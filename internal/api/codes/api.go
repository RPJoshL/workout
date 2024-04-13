package codes

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
			"NotFound",
			"GET",
			"/*",
			api.NotFound,
			router.Options{ForNotFound: true},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

func (api *Api) NotFound(w http.ResponseWriter, r *http.Request) {
	api.R().Tmpl.Render(api.notFound(), "notFound.title", "notFound.description")
}
