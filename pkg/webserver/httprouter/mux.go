package httprouter

import (
	"context"
	"net/http"
	"strings"
)

type contextKeyType int

const (
	KeyRealPath contextKeyType = iota
)

// Methods contains generic HTTP methods this router "supports".
// Feel free to extend this methods if you want to use another one!
var Methods = []string{
	"GET", "PUT", "POST", "PATCH", "DELETE", "PORPFIND",
}

// Mux is a wrapper around [*http.ServeMux] with
// additional methods to make the use of middlewares
// and mounting of routes easier
type Mux struct {
	*http.ServeMux

	// Path to append to all routes
	prefixPath string

	// Middlewares of the upper root mux
	rootMiddlewares []*[]func(next http.Handler) http.Handler

	// middleware chain to use for this Mux instance
	middlewares *[]func(next http.Handler) http.Handler
}

// NewMux creates a new instance of a mux
func NewMux() *Mux {
	return &Mux{
		ServeMux:    http.NewServeMux(),
		middlewares: &[]func(next http.Handler) http.Handler{},
		prefixPath:  "",
	}
}

// Get adds an handler func for the provided path and a GET request.
// It's the same as calling [HandleFunc] with a "GET" prefixed path
func (r *Mux) Get(path string, handler http.HandlerFunc) {
	r.HandleFunc("GET "+path, handler)
}

func (r *Mux) Handle(pattern string, handler http.Handler) {
	r.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	})
}

func (r *Mux) HandleFunc(pattern string, handler http.HandlerFunc) {
	// We overwrite the handler function
	r.ServeMux.HandleFunc(pattern, r.wrappWithMiddlewares(handler))
}

// Mount attaches another [http.Handler] or router as a subrouter along a routing path (e.g. "/api"). This path
// will be prefixed to the subrouter paths.
// It's useful to split up a large API as many independent routers and compose them as a single service using Mount.
//
// The middleware chain of this mux will also be used for the created sub router.
//
// To get correct paths in your endpoints, you have to mount [r.]
func (r *Mux) Mount(prefixPath string, handler http.Handler) *Mux {
	// Don't add double slash for root paths
	prefixPathAll := prefixPath
	if before, ok := strings.CutSuffix(prefixPath, "/"); ok {
		prefixPathAll = before
	}

	// Modify / stripe the path away that the handler does match again
	wrappedHandler := func(w http.ResponseWriter, r *http.Request) {
		// To set the correct and full URL, we have to do it like this:
		//   r.URL.Path = prefixPathAll + "/" + r.PathValue("xxNewPathDontUseThis")
		// Problem: with the modfied path, the already mounted sub paths doesn't match anymore!
		// So or only option is to store the path inside context and apply it on the last path
		r.URL.Path = "/" + r.PathValue("xxNewPathDontUseThis")

		// Store "real" path inside context
		if r.Context().Value(KeyRealPath) == nil {
			r = r.WithContext(context.WithValue(r.Context(), KeyRealPath, prefixPathAll+"/"+r.PathValue("xxNewPathDontUseThis")))
		}

		handler.ServeHTTP(w, r)
	}

	// Go only applies the new routing features if the route starts with a method definition :/.
	// We can't use the old one because of a "/*" route that would be in conflict with
	// the old one.
	// So the only soulution is setting up a listener for all "supported" methods
	for _, method := range Methods {
		r.HandleFunc(method+" "+prefixPathAll+"/{xxNewPathDontUseThis...}", wrappedHandler)

		// Also add a path for root handlers
		if prefixPathAll != "" {
			r.HandleFunc(method+" "+prefixPath, wrappedHandler)
		}
	}

	return r
}

// Group creates another [http.Handler] or router as a subrouter along a routing path. This path
// will be prefixed to the subrouter paths (it can be empty for the same route).
// It's useful to split up a large API or using an additional set of middlewares.
//
// The middlewares of this mux will also be used for the created sub router
func (r *Mux) Group(prefixPath string, inline func(mx *Mux)) *Mux {
	//nolint:all
	newRootMiddlewares := append(r.rootMiddlewares, r.middlewares)

	// Create new mux with sub paths and additional middlewares
	rtc := &Mux{
		rootMiddlewares: newRootMiddlewares,
		prefixPath:      r.prefixPath + prefixPath,
		middlewares:     &[]func(next http.Handler) http.Handler{},
		ServeMux:        http.NewServeMux(),
	}

	// Call inline function with the created mux
	inline(rtc)

	// Register mux
	if prefixPath == "" {
		prefixPath = "/"
	}

	r.Handle(prefixPath, rtc)

	return rtc
}

// ApplyRealPath sets and applied the correct path value for a request.
// You have to mount this middleware if you are using the [Mount] function
func ApplyRealPath(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		realPath := r.Context().Value(KeyRealPath)
		if realPath != "" && realPath != nil {
			r.URL.Path = realPath.(string)
		}

		next.ServeHTTP(w, r)
	})
}
