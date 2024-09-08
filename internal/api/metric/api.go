package metric

import (
	"net/http"

	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/pkg/response"
)

type Api struct {
	router.ApiRequest
}

func GetRoutes() *router.Router {
	api := &Api{}

	routes := router.Routes{
		router.NewRoute(
			"GetPaiScore",
			"GET",
			"/pai",
			api.GetPaiScore,
			router.Options{IsApiEndpoint: true},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

func (a *Api) GetPaiScore(w http.ResponseWriter, r *http.Request) {
	rtc, err := a.GetPaiProgression()
	if err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		response.WriteJson(rtc, 200, w)
	}
}
