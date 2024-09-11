package httprouter

import (
	"net/http"
)

// Use appends the provided middlewares to the middleware stack.
// The order is kept present! The middlewares are called in squence
func (m *Mux) Use(handler ...func(next http.Handler) http.Handler) {
	*m.middlewares = append(*m.middlewares, handler...)
}

// wrappWithMiddlewares wrapps the provided handler with the previously
// configured middleware stack
func (m *Mux) wrappWithMiddlewares(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Build a chain for all middlewares
		allMiddlewares := append(m.rootMiddlewares, m.middlewares)

		// Call all middlewares in sequence
		var prevHandler http.Handler = http.HandlerFunc(handler)
		for i := len(allMiddlewares) - 1; i >= 0; i-- {
			for a := len(*allMiddlewares[i]) - 1; a >= 0; a-- {
				prevHandler = http.Handler((*allMiddlewares[i])[a](prevHandler))
			}
		}
		prevHandler.ServeHTTP(w, r)
	}
}
