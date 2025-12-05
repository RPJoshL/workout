package generic

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
			"GenericSettingsPage",
			"GET",
			"/",
			api.GenericSettingsPage,
			router.Options{},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

func (api *Api) GenericSettingsPage(w http.ResponseWriter, r *http.Request) {
	api.R().Tmpl.Render(api.genericPage(api.R().User.User), "generic.appName", "generic.appName")
}
