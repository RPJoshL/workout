package kubernetes

import (
	"net/http"

	"git.rpjosh.de/RPJosh/go-webserver/response"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
)

type Api struct {
	router.ApiRequest
}

func GetRoutes() *router.Router {
	api := &Api{}

	routes := router.Routes{
		router.NewRoute(
			"Ready",
			"GET",
			"/readyz",
			api.Ready,
			router.Options{UseNoAuth: true},
		),
		router.NewRoute(
			"Health",
			"GET",
			"/healthz",
			api.Health,
			router.Options{UseNoAuth: true},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

func (api *Api) Ready(w http.ResponseWriter, r *http.Request) {
	response.WriteText("OK", 200, w)
}
func (api *Api) Health(w http.ResponseWriter, r *http.Request) {
	response.WriteText("OK", 200, w)
}
