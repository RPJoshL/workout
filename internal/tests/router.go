package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/dbutils"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
	"git.rpjosh.de/RPJosh/workout/pkg/database"
	"git.rpjosh.de/RPJosh/workout/pkg/response"
)

const DefaultUsername = "TEST"
const DefaultUserID = 10

// RouterConfig implements [router.Config] with test functions
type RouterConfig struct {

	// User to pass as a model
	User *models.WebUser

	// Database to use
	Db database.SqlConnection
}

// InjectRequestData sets all fields for the struct type
// [router.ApiRequestler] with a mocked one
func InjectRequestData(dst router.ApiRequestler, t *testing.T) {
	// Config with data
	conf := &RouterConfig{
		User: &models.WebUser{
			User: &models.User{
				Id:   DefaultUserID,
				Name: DefaultUsername,
				Mail: DefaultUsername,
			},
		},
		Db: GetDbConnection(t),
	}

	InjectRequestDataWithConfig(dst, conf)
}

// InjectRequestDataWithConfig sets all fields for the struct type
// [router.ApiRequestler] with a mocked one and the provided config
func InjectRequestDataWithConfig(dst router.ApiRequestler, conf *RouterConfig) {
	// Use UTC timezone globally
	if err := os.Setenv("TZ", "UTC"); err != nil {
		logger.Warning("Failed to normalize the time zone to UTC: %s", err)
	}

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
		"",
	)

	reflect.ValueOf(dst).Elem().Set(rtc.Elem())

	// Create a user on the db itself.
	// Almost everywhere the requests are user scoped...
	if conf.User.User != nil && conf.User.Id != 0 {
		createUser(*conf.User.User, dst.R().Db)
	}
}

func (r *RouterConfig) createApiRequest(request *http.Request, response http.ResponseWriter, route router.Route) router.ApiRequest {
	trans := translator.NewTranslator()
	trans.Language = translator.English

	req := router.NewApiRequestWithValues(
		route,
		dbutils.NewByDb(r.Db),
		logger.GetGlobalLogger(),
		"test",
		r.User,
		*trans,
		nil, nil,
	)

	return req
}

func SampleHandler(w http.ResponseWriter, r *http.Request) {
	response.WriteText("Not implemented in test mode :)", 200, w)
}

// createUser creates a user with the provided data on the database
func createUser(user models.User, db *dbutils.Db) {
	if _, err := db.Struct.Insert(&user).Run(); err != nil {
		logger.Fatal("Failed to create user on db: %s", err)
	}
}

// CreateDefaultUser creates the default dummy user within the database
func CreateDefaultUser(db *dbutils.Db) {
	createUser(models.User{
		Id:   DefaultUserID,
		Name: DefaultUsername,
		Mail: DefaultUsername,
	}, db)
}
