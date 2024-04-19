package router

import (
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"git.rpjosh.de/RPJosh/workout/internal/api/middleware"
	"git.rpjosh.de/RPJosh/workout/internal/database"
	"github.com/go-chi/chi/v5"
)

type Router struct {

	// Concrete struct dependency that is used for all functions given in "Routes.[].HandlerFunc".
	// This field can be nil and is only needed if the dependency includes the embedded field
	// "ApiRequest" for injecting request data
	Dependency ApiRequestler

	// Routes to mount on the root path
	Routes Routes
}

// Route represents a single API route consisting out of a path and a handler function
// that should be called on a match
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc func(w http.ResponseWriter, r *http.Request)
	Options     Options
}

// Options contains additional options that CAN be used to controle the behaviour
// of a route.
// This is totally optional
type Options struct {

	// Don't require any authentication
	UseNoAuth bool

	// Mount this endpoint for default 404 errors
	ForNotFound bool
}

type Routes []Route

func NewRoute(name string, method string, pattern string, handlerFunc func(w http.ResponseWriter, r *http.Request), options Options) Route {
	return Route{
		Name:        name,
		Method:      method,
		Pattern:     pattern,
		HandlerFunc: handlerFunc,
		Options:     options,
	}
}

// GetHandler returns a "http.Handler" that can be mounted with chi
// for all routes defined in this router
func (router *Router) GetHandlerWithRouter(r *chi.Mux) http.Handler {
	for _, route := range router.Routes {
		var handlerFunc http.HandlerFunc = http.HandlerFunc(route.HandlerFunc)

		// Register injection middleware for struct "ApiRequest"
		if router.Dependency != nil {
			// Overwrite the handler function
			handlerFunc = router.InjectionMiddleware(route.HandlerFunc, route)
		}

		// Add authentication middleware
		if !route.Options.UseNoAuth {
			next := handlerFunc
			handlerFunc = func(w http.ResponseWriter, r *http.Request) {
				middleware.AuthenticationMiddleware(
					next, GlobalConfig.JWTKey, database.NewDatabaseUtils(GlobalDb),
				).ServeHTTP(w, r)
			}
		}

		if route.Options.ForNotFound {
			r.NotFound(handlerFunc)
		} else {
			r.Method(route.Method, route.Pattern, handlerFunc)
		}
	}
	return r
}

// GetHandler returns a "http.Handler" that can be mounted with chi
// for all routes defined in this router
func (router *Router) GetHandler() http.Handler {
	return router.GetHandlerWithRouter(chi.NewRouter())
}

// InjectionMiddleware is a middleware that calls the "next" function in the chane stack with a copy of
// the internal dependency injected with the "ApiRequest"
func (router *Router) InjectionMiddleware(next func(w http.ResponseWriter, r *http.Request), route Route) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		copy := router.ParseAndCloneStruct(reflect.ValueOf(router.Dependency), r, w, route, NewApiRequest)

		// Find function name to call with reflection
		nameOfFunc := getFunctionName(next)
		copy.MethodByName(nameOfFunc).Call([]reflect.Value{reflect.ValueOf(w), reflect.ValueOf(r)})
	})
}

// ParseAndCloneStruct creates a clone (pointer derefernce) of the given reflection value injected with the data for
// the embedded struct "ApiRequest"
func (router *Router) ParseAndCloneStruct(
	ref reflect.Value, r *http.Request, w http.ResponseWriter, route Route,
	newApiRequest func(request *http.Request, response http.ResponseWriter, route Route) ApiRequest,
) reflect.Value {
	isPointer := ref.Kind() == reflect.Pointer

	// Indirect "ref" if it's a pointer to obmit further pointer checks
	refIndirect := ref
	for refIndirect.Kind() == reflect.Pointer {
		refIndirect = refIndirect.Elem()
	}

	// Copy the value given in "ref"
	var newValue reflect.Value
	if isPointer {
		newValue = reflect.New(ref.Elem().Type())
		newValue.Elem().Set(ref.Elem())
		for newValue.Kind() == reflect.Pointer {
			newValue = newValue.Elem()
		}
	} else {
		newValue = reflect.Indirect(ref)
	}

	// Loop through all struct fields and find fields with the struct or interface type
	// of "ApiRequest"
	typ := newValue.Type()
	for i := 0; i < typ.NumField(); i++ {
		newValField := newValue.Field(i)

		// Indirect "newValFiel" if it's a pointer to obmit further pointer checks
		newValFieldDe := newValField
		if newValField.Kind() == reflect.Pointer {
			newValFieldDe = newValField.Elem()
		}

		if newValFieldDe.IsValid() && newValFieldDe.CanInterface() {
			// Check if type is api endpointler (struct -> no Pointer)
			if newValFieldDe.Kind() == reflect.Struct {
				if _, ok := newValFieldDe.Interface().(ApiRequest); ok {
					// Create a new instance
					var newReq ApiRequest = newApiRequest(r, w, route)

					newValFieldDe.Set(reflect.ValueOf(newReq))
				} else if newValField.Type().Implements(reflect.TypeOf((*ApiRequestler)(nil)).Elem()) {
					newValField.Set(router.ParseAndCloneStruct(newValField, r, w, route, newApiRequest))
				}
			}
		}
	}

	// Return the new (copied) newValue
	if isPointer {
		return newValue.Addr()
	}
	return newValue
}

// getFunctionName returns the raw name of the given function (without struct name or other details)
func getFunctionName(temp interface{}) string {
	strs := strings.Split((runtime.FuncForPC(reflect.ValueOf(temp).Pointer()).Name()), ".")
	name := strs[len(strs)-1]

	if lastDash := strings.LastIndex(name, "-"); lastDash != -1 {
		return name[0:lastDash]
	}

	return name
}
