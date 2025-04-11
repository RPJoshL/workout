package dashboard

import (
	"net/http"

	"git.rpjosh.de/RPJosh/workout/internal/api/metric"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
)

type Api struct {
	router.ApiRequest

	Metric metric.Api
}

func GetRoutes() *router.Router {
	api := &Api{}

	routes := router.Routes{
		router.NewRoute(
			"DashbaordView",
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

func (a *Api) GetDashbaord(w http.ResponseWriter, r *http.Request) {
	// Fetch data
	data, err := a.GetDashboardData()
	if err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	a.R().Tmpl.Render(a.main(&data), "generic.appName", "generic.appName")
}
