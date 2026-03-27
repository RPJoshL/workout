package webserver

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"runtime/debug"
	"strings"

	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"github.com/RPJoshL/go-logger"
	"github.com/justinas/nosurf"
)

type ContextKeys int

const (

	// Context key for a unique request ID set by "RequestId"
	KeyIdentifier ContextKeys = iota

	// Context key to get a username as a string / [fmt.Sringer].
	// This string value will be added as a prefix for the logger
	KeyUsername ContextKeys = iota
)

var trueClientIP = http.CanonicalHeaderKey("True-Client-IP")
var xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
var xRealIP = http.CanonicalHeaderKey("X-Real-IP")

// SecureHeaders adds securely relevant headers to the response
func (server WebServer[T]) SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src * 'unsafe-inline' 'unsafe-eval'; script-src * 'unsafe-inline' 'unsafe-eval'; connect-src * 'unsafe-inline'; img-src * data: blob: 'unsafe-inline'; frame-src *; style-src * 'unsafe-inline';")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		next.ServeHTTP(w, r)
	})
}

// LogRequest logs all requests from the webserver to the debug logger
func (server WebServer[T]) LogRequest(next http.Handler) http.Handler {
	return LogRequest(next)
}

// LogRequest logs all requests from the webserver to the debug logger
func LogRequest(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get thre request path
		path := r.URL.RequestURI()

		// Build the log message
		message := fmt.Sprintf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, path)

		// Kubernetes health and readiness calls are logged with a trace level.
		// They are not relevant!
		if strings.HasSuffix(path, "/readyz") || strings.HasSuffix(path, "/healthz") {
			logger.Trace("%s", message)
		} else {
			l := getLogger(r)
			l.Debug("%s", message)
		}

		next.ServeHTTP(w, r)
	})
}

// getLogger returns a configured logging instance in context of the current request
func getLogger(r *http.Request) *logger.Logger {
	// Try to get request id
	id := r.Context().Value(KeyIdentifier)
	if id == nil {
		return logger.GetGlobalLogger()
	} else {
		// Get logger with prefix
		l := logger.CloneLogger(logger.GetGlobalLogger())
		prefix := fmt.Sprintf(" [%s]", id)

		// Maybe we have even a user reference
		if usr := r.Context().Value(KeyUsername); usr != nil {
			switch username := usr.(type) {
			case string:
				prefix += " [" + username + "]"
			case fmt.Stringer:
				prefix += " [" + username.String() + "]"
			}
		}

		l.Prefix = prefix
		return l
	}
}

// RecoverPanic is a middleware that catches panics inside the handling of a response.
// It closes the connection and logs the error cause
func (server WebServer[T]) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Get the debug stack trace with file and line informations
				trace := fmt.Sprintf("%s\n%s", fmt.Errorf("%s", err).Error(), debug.Stack())

				// Call the errors function
				errors.Config.HandlePanic(err, trace, w, r)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// NoSurf createa a noSurf middleware which uses a customized CSRF cookie with
// the Secure, Path and HttpOnly attributes set
func (server WebServer[T]) NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}

// RequestId adds an unique identified to every request
func (server WebServer[T]) RequestId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate random ID
		id, _ := generateRandomString(8)

		// Set it as a context value
		req := r.WithContext(context.WithValue(r.Context(), KeyIdentifier, id))

		next.ServeHTTP(w, req)
	})
}

// GenerateRandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly
func generateRandomString(n int) (string, error) {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, n)
	for i := range n {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}
	return string(ret), nil
}

// RealIP is a middleware that sets a http.Request's RemoteAddr to the results
// of parsing the True-Client-IP, X-Real-IP or the X-Forwarded-For headers
// (in that order).
//
// You should only use this middleware if you can trust the headers passed to
// you (in particular, the headers this middleware uses).
func (server WebServer[T]) RealIP(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if rip := realIP(r); rip != "" {
			r.RemoteAddr = rip
		}
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func realIP(r *http.Request) string {
	var ip string

	if tcip := r.Header.Get(trueClientIP); tcip != "" {
		ip = tcip
	} else if xrip := r.Header.Get(xRealIP); xrip != "" {
		ip = xrip
	} else if xff := r.Header.Get(xForwardedFor); xff != "" {
		i := strings.Index(xff, ",")
		if i == -1 {
			i = len(xff)
		}
		ip = xff[:i]
	}
	if ip == "" || net.ParseIP(ip) == nil {
		return ""
	}
	return ip
}
