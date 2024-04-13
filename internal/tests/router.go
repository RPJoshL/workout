package tests

import (
	"net/http"
	"net/http/httptest"
	"reflect"

	"git.rpjosh.de/RPJosh/go-webserver/response"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
)

const defaultUsername = "TEST"

// RouterConfig implements [router.Config] with test functions
type RouterConfig struct {
	// Username to pass as user model
	Username string
}

// InjectRequestData sets all fields for the struct type
// [router.ApiRequestler] with a mocked one
func InjectRequestData(dst router.ApiRequestler, conf *RouterConfig) {

	// We don't need any data inside router struct for parsing
	r := &router.Router{}

	// Recorder for response
	rec := httptest.NewRecorder()

	r.ParseAndCloneStruct(
		reflect.ValueOf(dst), &http.Request{}, rec, router.NewRoute(
			"TestName",
			"GET",
			"/navigator",
			SampleHandler,
			router.Options{
				UseNoAuth: true,
			},
		),
		router.NewApiRequest,
	)
}

func SampleHandler(w http.ResponseWriter, r *http.Request) {
	response.WriteText("Not implemented in test mode :)", 200, w)
}
