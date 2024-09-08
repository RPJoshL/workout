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

func (api *Api) GetDashbaord(w http.ResponseWriter, r *http.Request) {

	// Fetch data
	data, err := api.GetDashboardData()
	if err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	api.R().Tmpl.Render(api.main(&data), "generic.appName", "generic.appName")
}
