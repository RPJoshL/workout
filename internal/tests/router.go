package tests

import (
	"net/http"
	"net/http/httptest"
	"reflect"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/go-webserver/response"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/database"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
)

const defaultUsername = "TEST"
const defaultUserID = 1

// RouterConfig implements [router.Config] with test functions
type RouterConfig struct {

	// User to pass as a model
	User models.User

	// Database to use
	Db database.SqlConnection
}

// InjectRequestData sets all fields for the struct type
// [router.ApiRequestler] with a mocked one
func InjectRequestData(dst router.ApiRequestler) {

	// Config with data
	conf := &RouterConfig{
		User: models.User{
			Id:   defaultUserID,
			Name: defaultUsername,
			Mail: defaultUsername,
		},
		Db: GetDbConnection(),
	}

	InjectRequestDataWithConfig(dst, conf)
}

// InjectRequestData sets all fields for the struct type
// [router.ApiRequestler] with a mocked one and the provided config
func InjectRequestDataWithConfig(dst router.ApiRequestler, conf *RouterConfig) {

	// We don't need any data inside router struct for parsing
	r := &router.Router{}

	// Recorder for response
	rec := httptest.NewRecorder()

	rtc := r.ParseAndCloneStruct(
		reflect.ValueOf(dst), &http.Request{}, rec, router.NewRoute(
			"TestName",
			"GET",
			"/navigator",
			SampleHandler,
			router.Options{
				UseNoAuth: true,
			},
		),
		conf.createApiRequest,
	)

	reflect.ValueOf(dst).Elem().Set(rtc.Elem())
}

func (r *RouterConfig) createApiRequest(request *http.Request, response http.ResponseWriter, route router.Route) router.ApiRequest {
	trans := translator.NewTranslator()
	trans.Language = translator.English

	req := router.NewApiRequestWithValues(
		route,
		database.NewDatabaseUtilsByDb(r.Db),
		logger.GetGlobalLogger(),
		"test",
		r.User,
		*trans,
	)

	return req
}

func SampleHandler(w http.ResponseWriter, r *http.Request) {
	response.WriteText("Not implemented in test mode :)", 200, w)
}
