package webserver

import (
	"bufio"
	"context"
	"errors"
	"net"
	"net/http"

	"git.rpjosh.de/RPJosh/workout/pkg/webserver/httprouter"
)

// InternalProcessedHeader is used to identify endpoints of your application so they
// won't be modified by [BodyOverrider]
const InternalProcessedHeader = "X-Processed"

// BodyOverrider overrides the response body for a 404 error code
// that was not handled within the application (returned by go [net/http]).
//
// You have to use [SetOverrideHeader]
type BodyOverrider struct {
	http.ResponseWriter

	// HTTP code that was written to the response
	code    int
	request *http.Request

	// Weather to override the response body within [Write]
	override bool

	override404 func(request *http.Request) []byte
	headers404  func(request *http.Request, writer http.ResponseWriter)
}

// NewBodyOverride initializes a new struct that overrides the response body for 404 status
// codes by the provided function.
// The pre override function should set any headers for the content in the main function.
// You won't be able to modify the headers inside "override404"!
func NewBodyOverride(override404 func(request *http.Request) []byte, headers404 func(request *http.Request, writer http.ResponseWriter)) *BodyOverrider {
	return &BodyOverrider{
		override404: override404,
		headers404:  headers404,
	}
}

// SetOverrideHeader adds a header to a response so it won't be modified
// by [BodyOverrider].
// You should add it as a middleware for all your API requests
func SetOverrideHeader(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(InternalProcessedHeader, "true")
		next.ServeHTTP(w, r)
	})
}

func (b *BodyOverrider) WriteHeader(code int) {
	// Set override flag
	b.override = b.Header().Get(InternalProcessedHeader) != "true"
	b.code = code

	// Remove the internal header
	b.Header().Del(InternalProcessedHeader)

	// Set any headers required for content later
	if b.override && code == 404 {
		b.headers404(b.request, b.ResponseWriter)
	}

	b.ResponseWriter.WriteHeader(code)
}

func (b *BodyOverrider) Unwrap() http.ResponseWriter {
	return b.ResponseWriter
}

// Wrap wraps the handler with this custom body overrider
func (b *BodyOverrider) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Function to call for 404
		f404 := b.override404
		fh404 := b.headers404
		if f404 == nil {
			f404 = default404
		}
		if fh404 == nil {
			fh404 = default404Headers
		}

		// Fallback for real path: always set it here (before it's processed by [httprouter.Mount])
		r = r.WithContext(context.WithValue(r.Context(), httprouter.KeyRealPath, r.URL.Path))

		// Create a copy of overrider
		lw := &BodyOverrider{
			override404:    f404,
			headers404:     fh404,
			request:        r,
			ResponseWriter: w,
		}

		next.ServeHTTP(lw, r)
	})
}

func (b *BodyOverrider) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := b.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return h.Hijack()
}

func (b *BodyOverrider) Write(body []byte) (int, error) {
	// Override body
	if b.override && b.code == http.StatusNotFound {
		body = b.override404(b.request)
	}

	return b.ResponseWriter.Write(body)
}

func default404(request *http.Request) []byte {
	return []byte(`404 page not found`)
}
func default404Headers(request *http.Request, writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "text/plain")
}
