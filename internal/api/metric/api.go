package metric

import (
	"encoding/json"
	"net/http"

	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
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
		router.NewRoute(
			"StoreSteps",
			"POST",
			"/steps",
			api.StoreStepsApi,
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

func (a *Api) StoreStepsApi(w http.ResponseWriter, r *http.Request) {
	// Parse body
	steps := []models.Steps{}
	if err := json.NewDecoder(r.Body).Decode(&steps); err != nil {
		errors.BadRequest("").Log("Failed to decode body of steps", err, a).Write(w, r)
	}

	if rtc, err := a.StoreSteps(steps); err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		response.WriteJson(rtc, 200, w)
	}
}
