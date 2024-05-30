// response contains some generic helper functions to write data
// to the HTTP ResponseWriter
package response

import (
	"encoding/json"
	"net/http"

	"git.rpjosh.de/RPJosh/go-logger"
)

func WriteJson(data interface{}, statusCode int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func WriteJsonRaw(json []byte, statusCode int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(json)
}

func Write(statusCode int, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
}

func WriteText(text string, statusCode int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	w.Write([]byte(text))
}

// WriteError handles an unexpected server error.
// A generic error message will be returned to the client.
func WriteError(err error, w http.ResponseWriter, r *http.Request) {
	logger.Debug("Unexpected server error for path '%s': %s", r.URL, err)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(500)
	w.Write([]byte("Internal Server error"))
}

// RedirectTo redirects the user to a specific path by setting
// the "Location" Header and returning a temporary redirect code
func RedirectTo(path string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", path)
	w.WriteHeader(302)
}
