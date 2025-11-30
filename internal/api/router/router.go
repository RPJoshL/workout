package router

import (
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/api/middleware"
	"git.rpjosh.de/RPJosh/workout/internal/dbutils"
	"git.rpjosh.de/RPJosh/workout/pkg/webserver"
	"git.rpjosh.de/RPJosh/workout/pkg/webserver/httprouter"
)

type Router struct {

	// Concrete struct dependency that is used for all functions given in "Routes.[].HandlerFunc".
	// This field can be nil and is only needed if the dependency includes the embedded field
	// "ApiRequest" for injecting request data
	Dependency ApiRequestler

	// Routes to mount on the root path
	Routes Routes

	// Any additional routers to mount while building the [http.Handler]
	ExtraRouter []*Router

	// Wheather to only mount API endpoints
	OnlyMountApi bool
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

	// Mount this route with a prefix of "/api/v1".
	// The response should always be in the JSON format!
	IsApiEndpoint bool
}

type Routes []Route

func NewRoute(name, method, pattern string, handlerFunc func(w http.ResponseWriter, r *http.Request), options Options) Route {
	return Route{
		Name:        name,
		Method:      method,
		Pattern:     pattern,
		HandlerFunc: handlerFunc,
		Options:     options,
	}
}

// OnlyApi mounts only API endpoints into this router
func (router *Router) OnlyApi() *Router {
	router.OnlyMountApi = true
	return router
}

// GetHandlerWithRouter returns a "http.Handler" that can be mounted with chi
// for all routes defined in this router
func (router *Router) GetHandlerWithRouter(r *httprouter.Mux) http.Handler {
	for _, route := range router.Routes {
		// Only mount API endpoints or UI endpoints
		if route.Options.IsApiEndpoint != router.OnlyMountApi {
			continue
		}

		var handlerFunc = http.HandlerFunc(route.HandlerFunc)

		// Register injection middleware for struct "ApiRequest"
		if router.Dependency != nil {
			// Overwrite the handler function
			handlerFunc = router.InjectionMiddleware(route.HandlerFunc, route)
		}

		// Log all requests
		handlerFunc = webserver.LogRequest(handlerFunc)

		// Add authentication middleware
		if !route.Options.UseNoAuth {
			next := handlerFunc
			handlerFunc = func(w http.ResponseWriter, r *http.Request) {
				middleware.AuthenticationMiddleware(
					next, GlobalConfig.JWTKey, dbutils.New(GlobalDb),
				).ServeHTTP(w, r)
			}
		}

		// Add a key that this request was processed
		handlerFunc = webserver.SetOverrideHeader(handlerFunc)

		// Apply correct path
		handlerFunc = httprouter.ApplyRealPath(handlerFunc)

		r.Handle(route.Method+" "+route.Pattern, handlerFunc)
	}

	// Add additional router
	for _, rr := range router.ExtraRouter {
		// This property should be passed through
		rr.OnlyMountApi = router.OnlyMountApi

		rr.GetHandlerWithRouter(r)
	}

	return r
}

// GetHandler returns a "http.Handler" that can be mounted with chi
// for all routes defined in this router
func (router *Router) GetHandler() http.Handler {
	return router.GetHandlerWithRouter(httprouter.NewMux())
}

// InjectionMiddleware is a middleware that calls the "next" function in the chane stack with a copy of
// the internal dependency injected with the "ApiRequest"
func (router *Router) InjectionMiddleware(next func(w http.ResponseWriter, r *http.Request), route Route) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cp := router.ParseAndCloneStruct(reflect.ValueOf(router.Dependency), r, w, route, NewApiRequest, "")

		// Find function name to call with reflection
		nameOfFunc := getFunctionName(next)
		cp.MethodByName(nameOfFunc).Call([]reflect.Value{reflect.ValueOf(w), reflect.ValueOf(r)})
	})
}

// ParseAndCloneStruct creates a clone (pointer derefernce) of the given reflection value injected with the data for
// the embedded struct "ApiRequest"
func (router *Router) ParseAndCloneStruct(
	ref reflect.Value, r *http.Request, w http.ResponseWriter, route Route,
	newApiRequest func(request *http.Request, response http.ResponseWriter, route Route) ApiRequest,
	ignoreFields string,
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
	for i := range typ.NumField() {
		newValField := newValue.Field(i)

		// Indirect "newValFiel" if it's a pointer to obmit further pointer checks
		newValFieldDe := newValField
		if newValField.Kind() == reflect.Pointer {
			newValFieldDe = newValField.Elem()
		}

		if newValFieldDe.IsValid() && newValFieldDe.CanInterface() {
			// Check if type is api endpointler (struct → no Pointer)
			if newValFieldDe.Kind() == reflect.Struct {
				if _, ok := newValFieldDe.Interface().(ApiRequest); ok {
					// Create a new instance
					var newReq = newApiRequest(r, w, route)

					newValFieldDe.Set(reflect.ValueOf(newReq))
				} else if newValField.Type().Implements(reflect.TypeFor[ApiRequestler]()) {
					newValField.Set(router.ParseAndCloneStruct(newValField, r, w, route, newApiRequest, ignoreFields))
				}
			} else if newValFieldDe.Kind() == reflect.Interface && newValFieldDe.Elem().IsValid() {
				// Instead of a directly specified struct, an interface was used.
				// We try to get the underlaying value of the interface from the original
				// value. Because this will probably result into an import cycle, we have
				// to ignore this field in any other parsing method
				newValFieldInterfaced := newValFieldDe.Elem()
				ignore := "," + newValFieldInterfaced.Type().String() + "#" + newValFieldDe.Type().String() + ","
				if strings.Contains(ignoreFields, ignore) {
					logger.Trace("Ignoring type %q to avoid a cycle", ignore)
					continue
				}

				if _, ok := newValFieldInterfaced.Interface().(ApiRequest); ok {
					// Create a new instance
					var newReq = newApiRequest(r, w, route)

					newValFieldDe.Set(reflect.ValueOf(newReq))
				} else if newValFieldInterfaced.Type().Implements(reflect.TypeFor[ApiRequestler]()) {
					newValField.Set(router.ParseAndCloneStruct(newValFieldInterfaced, r, w, route, newApiRequest, ignoreFields+ignore))
				}
			} else if newValFieldDe.Kind() == reflect.Interface {
				logger.Trace("No concrete type for interface %q given. Cannot inject it!", newValFieldDe.Type().Name())
			}
		}
	}

	// Return the new (copied) newValue
	if isPointer {
		return newValue.Addr()
	}
	return newValue
}

// AddRouter adds an external router that is mounted to this
// [http.Handler] when retrieving the handler
func (router *Router) AddRouter(rr *Router) *Router {
	router.ExtraRouter = append(router.ExtraRouter, rr)
	return router
}

// getFunctionName returns the raw name of the given function (without struct name or other details)
func getFunctionName(temp any) string {
	strs := strings.Split((runtime.FuncForPC(reflect.ValueOf(temp).Pointer()).Name()), ".")
	name := strs[len(strs)-1]

	if lastDash := strings.LastIndex(name, "-"); lastDash != -1 {
		return name[0:lastDash]
	}

	return name
}
