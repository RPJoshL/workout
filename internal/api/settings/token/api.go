package token

import (
	"net/http"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/api/token"
	"git.rpjosh.de/RPJosh/workout/internal/api/utils"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"github.com/guregu/null/v5"
)

type Api struct {
	router.ApiRequest

	Token token.Api
}

func (api *Api) GetRouter() *router.Router {
	routes := router.Routes{
		router.NewRoute(
			"TokenSettingsPage",
			"GET",
			"/token/",
			api.TokenSettingsPage,
			router.Options{},
		),
		router.NewRoute(
			"CreateTokenPage",
			"GET",
			"/token/new",
			api.CreateTokenPage,
			router.Options{},
		),
		router.NewRoute(
			"CreateToken",
			"POST",
			"/token",
			api.CreateToken,
			router.Options{},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

func (api *Api) TokenSettingsPage(w http.ResponseWriter, r *http.Request) {

	// Get all tokens
	allTokens, err := api.Token.GetAllTokens()
	if err != nil {
		err.GetErrorStruct().Write(w, r)
	}

	api.R().Tmpl.Render(api.tokenPage(allTokens), "generic.appName", "generic.appName")
}

func (api *Api) CreateTokenPage(w http.ResponseWriter, r *http.Request) {

	// Get all tokens
	allTokens, err := api.Token.GetAllTokens()
	if err != nil {
		err.GetErrorStruct().Write(w, r)
	}

	// Render as a token
	api.R().Tmpl.RenderModal(
		api.createToken(), "settings.token.create",
		api.tokenPage(allTokens), "/settings/token/",
		"generic.appName", "generic.appName", "",
	)
}

func (api *Api) CreateToken(w http.ResponseWriter, r *http.Request) {

	// Parse data from form
	if err := r.ParseMultipartForm(utils.MToBytes(1)); err != nil {
		errors.BadRequest("Failed to parse form").Log("Failed to parse form", err, api).Write(w, r)
	}

	token := models.ApiKey{}
	token.Alias = null.StringFrom(r.Form.Get("alias"))
	logger.Debug("Got %s", token.Alias.String)
	if r.Form.Get("validUntil") != "" {
		var parseError error
		token.ValidUntil, parseError = time.Parse("02.01.2006", r.Form.Get("validUntil"))
		if parseError != nil {
			errors.BadRequest("Invalid date format").Write(w, r)
			return
		}
	}

	// Create token
	createdToken, err := api.Token.CreateToken(token, 0)
	if err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	// Set token value as a response header to read it out from the UI
	w.Header().Set("X-Api-Key", createdToken.Key)

	// Refresh token list
	allTokens, err := api.Token.GetAllTokens()
	if err != nil {
		err.GetErrorStruct().Write(w, r)
	}

	api.R().Tmpl.RenderDirect(api.tokenTable(allTokens))
}
