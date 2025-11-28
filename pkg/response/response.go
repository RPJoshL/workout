// Package response contains some generic helper functions to write data
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
	logError("json", json.NewEncoder(w).Encode(data))
}

func WriteJsonRaw(json []byte, statusCode int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err := w.Write(json)
	logError("jsonRaw", err)
}

// WriteJsonWithFields writes the provided data as a JSON response body.
// Only struct fields that are present in [fieldsToInclude] will be includede in the
// JSON response. Fields must be genereted by "go-ddl"
func WriteJsonWithFields(data interface{}, fieldsToInclude []string, statusCode int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	logError("jsonWithFields", json.NewEncoder(w).Encode(StructToJSON(data, nil, fieldsToInclude)))
}

// WriteJsonWithoutFields writes the provided data as a JSON response body.
// Struct fields that are present in [fieldsToExclude] will not be includede in the
// JSON response. Fields must be genereted by "go-ddl"
func WriteJsonWithoutFields(data interface{}, fieldsToExclude []string, statusCode int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	logError("jsonWithoutfields", json.NewEncoder(w).Encode(StructToJSON(data, fieldsToExclude, nil)))
}

func Write(statusCode int, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
}

func WriteText(text string, statusCode int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	_, err := w.Write([]byte(text))
	logError("text", err)
}

// WriteError handles an unexpected server error.
// A generic error message will be returned to the client.
func WriteError(err error, w http.ResponseWriter, r *http.Request) {
	logger.Debug("Unexpected server error for path '%s': %s", r.URL, err)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(500)
	_, e := w.Write([]byte("Internal Server error"))
	logError("error", e)
}

// RedirectTo redirects the user to a specific path by setting
// the "Location" Header and returning a temporary redirect code (302)
func RedirectTo(path string, code int, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", path)
	w.WriteHeader(code)
}

func logError(context string, err error) {
	if err != nil {
		logger.Debug("Failed to write %s response: %s", context, err)
	}
}

// DonwloadableFile sets required fields so the browser will open
// a download dialog with the provided filename and the binary content
func DonwloadableFile(w http.ResponseWriter, fileName, contentType string, content []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", `attachment; filename="`+fileName+`"`)
	w.WriteHeader(http.StatusOK)

	_, e := w.Write(content)
	logError("error", e)
}
