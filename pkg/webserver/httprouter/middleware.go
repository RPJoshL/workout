package httprouter

import (
	"net/http"
)

// Use appends the provided middlewares to the middleware stack.
// The order is kept present! The middlewares are called in squence
func (r *Mux) Use(handler ...func(next http.Handler) http.Handler) {
	*r.middlewares = append(*r.middlewares, handler...)
}

// wrappWithMiddlewares wrapps the provided handler with the previously
// configured middleware stack
func (r *Mux) wrappWithMiddlewares(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		//nolint:all Build a chain for all middlewares
		allMiddlewares := append(r.rootMiddlewares, r.middlewares)

		// Call all middlewares in sequence
		var prevHandler http.Handler = handler
		for i := len(allMiddlewares) - 1; i >= 0; i-- {
			for a := len(*allMiddlewares[i]) - 1; a >= 0; a-- {
				prevHandler = (*allMiddlewares[i])[a](prevHandler)
			}
		}
		prevHandler.ServeHTTP(w, req)
	}
}
