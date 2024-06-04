package router

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/api/components"
	"git.rpjosh.de/RPJosh/workout/internal/api/templates"
	"git.rpjosh.de/RPJosh/workout/internal/database"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
	"git.rpjosh.de/RPJosh/workout/pkg/webserver"
)

// ApiRequest is a base struct you can embed inside a controller or repository to obtain
// request specific data.
//
// A struct is only allowed to embed "routes.ApiRequest" if it was passed to a router or it is a child from a struct that does implement "routes.ApiRequest"
//
// For each struct that embed "routes.ApiRequest", the following rules apply:
//   - request specific objects like the database or user information are available to access via the field `requestData`
//   - non pointer fields inside the struct are resseted for every request to the default value that was used during the creation of the route
//   - the reference of pointer fields are kept present, if it is not a struct that does embed `routes.ApiRequest`
//   - all fields that should be parsed as an "ApiRequest" have to be public (begin with an uppercase character)
type ApiRequest struct {

	// Contains the request specific information. This can be nil if the rules for the struct "ApiRequest" are invalidated
	requestData *Request
}

// Request contains the request specific information for the current request
type Request struct {

	// The route that was called from the client
	Route Route

	// Tr contains all the translations of the app
	Tr translator.Translator

	// Tmpl contains generic components to render your HTML component
	Tmpl templates.Templates

	// Comp is a collection of generic components for an HTML page
	Comp *components.Components

	// Db is a wrapper around "sql.Db" with functions to query data
	Db *database.DatabaseUtils

	// User which initiated the request (on authorized path)
	User *models.WebUser

	// ID is a "unique" ID of this request
	id string

	// Logger instance that logs message in a request context (username + request ID)
	Logger *logger.Logger

	// Generic parser for request data
	Parser *RequestParser

	// Request
	request *http.Request
	// Response
	response http.ResponseWriter
}

// ApiRequestler is a interface that is used to identifiy nested structs
// inside a root struct that does embed "ApiRequest".
//
// In general, you should not use this interface directly as it is only used during
// the injection of the request
type ApiRequestler interface {
	IsApiRequestInjectable() bool
	R() *Request
}

// Make sure that "ApiRequest" always embed "ApiRequestler"
var _ ApiRequestler = (*ApiRequest)(nil)

// Global variables that are initialized BEFORE server is starting and doesn't change
var GlobalTranslator *translator.Translator
var GlobalConfig *models.AppConfig
var GlobalDb *sql.DB

// R returns a request data object that contains information for the current request within
// the invoction context of the function.
//
// You should only use this method inside a struct that embeds "ApiRequest"
func (api ApiRequest) R() *Request {
	return api.requestData
}

// Logger is a shortcut for [ApiRequest.Request().Logger]
func (api *ApiRequest) Logger() *logger.Logger {
	return api.requestData.Logger
}

func NewApiRequest(request *http.Request, response http.ResponseWriter, route Route) ApiRequest {
	api := ApiRequest{requestData: &Request{
		Route:    route,
		request:  request,
		response: response,
	}}

	// Request ID. Try to get an existing ID set by a middleware
	if id := request.Context().Value(webserver.KeyIdentifier); id != nil {
		if idVal, ok := id.(string); ok {
			api.requestData.id = idVal
		}
	} else {
		api.requestData.id, _ = utils.GenerateRandomString(8)
	}
	loggerPrefix := fmt.Sprintf(" [%s]", api.requestData.id)

	// Maybe we have even a user context
	if usr := request.Context().Value(webserver.KeyUsername); usr != nil {
		if username, ok := usr.(string); ok {
			loggerPrefix += " [" + username + "]"
		} else if userSt, ok := usr.(fmt.Stringer); ok {
			loggerPrefix += " [" + userSt.String() + "]"
		}
	}

	// Logger
	api.requestData.Logger = logger.CloneLogger(logger.GetGlobalLogger())
	api.requestData.Logger.Prefix = loggerPrefix

	// Add translator based on path (copy its value)
	trans := *GlobalTranslator
	// Get language based on browser language
	if acceptLang := request.Header.Get("Accept-Language"); acceptLang != "" {
		if strings.HasPrefix(acceptLang, "de") {
			trans.Language = translator.German
		} else {
			trans.Language = translator.English
		}
	}
	api.R().Tr = trans

	// Set components
	api.R().Comp = components.NewComponents(&trans)

	// Add databse
	api.R().Db = database.NewDatabaseUtils(GlobalDb)

	// Get own user reference
	if user := request.Context().Value(models.KeyUser); user != nil {
		updateUser := false
		api.R().User, updateUser = models.NewWebUser(user.(*models.User), request.Header.Get("Time-Zone"))
		if updateUser {
			sel := api.R().Db.Struct.Update(api.R().User.User).Selector(database.ColumnSelector{IncludeColumns: models.WebUserProperties})
			if err := sel.Run(); err != nil {
				api.Logger().Warning("Failed to update user from Web-Data: %s", err)
			}
		}
	}

	// Add generic template functions
	api.R().Tmpl = *templates.NewTemplates(&trans, GlobalConfig, response, request, api.R().Comp, api.R().User)

	// Add request parser
	api.requestData.Parser = &RequestParser{Request: request}

	return api
}

// NewApiRequestWithValues returns a new [ApiRequest] with the provided data
func NewApiRequestWithValues(route Route, db *database.DatabaseUtils, logger *logger.Logger, id string, user models.WebUser, Tr translator.Translator) ApiRequest {
	rtc := ApiRequest{requestData: &Request{
		Route:  route,
		Db:     db,
		Logger: logger,
		id:     id,
		User:   &user,
		Tr:     Tr,
	}}

	return rtc
}

func (api ApiRequest) IsApiRequestInjectable() bool {
	return true
}

// GetHttpRequest returns the underlaying http request and
// response of the current HTTP call
func (req Request) GetHttpRequest() (*http.Request, http.ResponseWriter) {
	return req.request, req.response
}
