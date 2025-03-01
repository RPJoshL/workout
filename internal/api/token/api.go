package token

import (
	"encoding/json"
	"net/http"
	"strconv"

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
			"CreateToken",
			"POST",
			"/",
			api.CreateTokenApi,
			router.Options{IsApiEndpoint: true},
		),
		router.NewRoute(
			"GetToken",
			"GET",
			"/{id}",
			api.ShowTokenApi,
			router.Options{IsApiEndpoint: true},
		),
		router.NewRoute(
			"DeleteToken",
			"DELETE",
			"/{id}",
			api.DeleteTokenApi,
			router.Options{IsApiEndpoint: true},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

type RequestApikey struct {
	models.ApiKey

	ValidUntilOffset int `json:"validUntilOffset"`
}

func (a *Api) CreateTokenApi(w http.ResponseWriter, r *http.Request) {

	// Parse body
	tokenDetails := RequestApikey{}
	if err := json.NewDecoder(r.Body).Decode(&tokenDetails); err != nil {
		errors.BadRequest("").Log("Failed to decode body: %s", err, a).Write(w, r)
		return
	}

	// Create token
	if token, err := a.CreateToken(tokenDetails.ApiKey, tokenDetails.ValidUntilOffset); err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		response.WriteJson(token, 201, w)
	}
}

func (a *Api) ShowTokenApi(w http.ResponseWriter, r *http.Request) {
	// Get ID to query from path
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errors.BadRequest("#generic.numericError").Sprintf("id", r.PathValue("id")).Write(w, r)
		return
	}

	if key, err := a.showApikey(id); err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		response.WriteJsonWithoutFields(key, []string{models.ApiKey_Key}, 200, w)
	}
}

func (a *Api) DeleteTokenApi(w http.ResponseWriter, r *http.Request) {
	// Get ID to query from path
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errors.BadRequest("ID is not numeric").Write(w, r)
		return
	}

	if err := a.deleteApiKey(id); err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		response.WriteText("Deleted", 200, w)
	}
}
