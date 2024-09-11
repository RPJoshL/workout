package codes

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/database"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
	"git.rpjosh.de/RPJosh/workout/pkg/webserver/httprouter"
)

type Api struct {
	Tr *translator.Translator
	Db *database.DatabaseUtils
}

// NotFound renders a simple 404 page
func (api *Api) NotFound(request *http.Request) []byte {

	// Simple output for API
	if isApi(request) {
		return []byte("Resource / endpoint not found")
	}

	// Capture output of response writer
	recorder := httptest.NewRecorder()

	// Mock up a simple ApiRequest where the bytes a written to
	apiRequest := router.NewApiRequestWithValues(
		router.Route{}, api.Db, logger.GetGlobalLogger(), "404",
		models.WebUser{User: &models.User{DarkTheme: 1}}, *api.Tr, request, recorder,
	)

	// Generate output
	apiRequest.R().Tmpl.Render(api.notFound(apiRequest.R()), "notFound.title", "notFound.description")

	// Return mocked response body
	return recorder.Body.Bytes()
}

func (api *Api) NotFoundHeaders(request *http.Request, writer http.ResponseWriter) {
	if isApi(request) {
		writer.Header().Set("Content-Type", "text/plain")
	} else {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
}

func isApi(request *http.Request) bool {
	path := request.URL.Path
	realPath := request.Context().Value(httprouter.KeyRealPath)
	if realPath != "" && realPath != nil {
		path = realPath.(string)
	}

	return strings.HasPrefix(path, "/api/")
}
