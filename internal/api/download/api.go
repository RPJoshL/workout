package download

import (
	"net/http"

	"git.rpjosh.de/RPJosh/workout/internal/api/router"
)

type Api struct {
	router.ApiRequest
	version string
}

func GetRoutes(version string) *router.Router {
	api := &Api{
		version: version,
	}

	routes := router.Routes{
		router.NewRoute(
			"DownloadPage",
			"GET",
			"/",
			api.GetDownloadPage,
			router.Options{},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

func (a *Api) GetDownloadPage(w http.ResponseWriter, r *http.Request) {
	a.R().Tmpl.Render(a.downloadPage(), "generic.appName", "generic.appName")
}
