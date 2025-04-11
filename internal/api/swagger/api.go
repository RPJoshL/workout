package swagger

import (
	"net/http"

	docs "git.rpjosh.de/RPJosh/workout/doc"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

type Api struct {
	router.ApiRequest
}

func GetRoutes() *router.Router {
	api := &Api{}

	routes := router.Routes{
		router.NewRoute(
			"SwaggerUI",
			"GET",
			"/",
			api.GetSwaggerUI,
			router.Options{
				UseNoAuth: true,
			},
		),
		router.NewRoute(
			"SwaggerFile",
			"GET",
			"/swagger.yaml",
			api.GetSwaggerFile,
			router.Options{
				UseNoAuth: true,
			},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

func (a *Api) GetSwaggerFile(w http.ResponseWriter, request *http.Request) {
	content, err := docs.ApiV1.ReadFile("api/api_v1.yaml")
	if err != nil {
		a.Logger().Warning("Failed to read api documentation file (api_v1.yaml): %s", err)
		errors.InternalError().Write(w, request)
		return
	}

	// Return content of file
	w.Header().Set("Content-Type", "application/x-yaml")
	w.WriteHeader(200)
	_, _ = w.Write(content)
}

func (a *Api) GetSwaggerUI(w http.ResponseWriter, request *http.Request) {
	content, err := docs.SwaggerUI.ReadFile("api/swagger.html")
	if err != nil {
		a.Logger().Warning("Failed to read swagger UI file (swagger.html): %s", err)
		errors.InternalError().Write(w, request)
		return
	}

	// Return content of file
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	_, _ = w.Write(content)
}
