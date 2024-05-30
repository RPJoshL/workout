// errors is a package to handle and write errors for API reqeuests. It is a wrapper
// around the generic error interface extended with additional information and methods.
//
// It can be used from API endpoints to return custom
// messages with an HTTP ResponseWriter
package errors

import (
	"errors"
	"fmt"
	"net/http"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/pkg/response"
)

// Config is a global static variable that you can use to customize some aspects
// in this package.
// See "CustomErrorConfig" for available methods and a detailed description
var Config CustomErrorConfig = DefaultConfig{}

// CustomErrorConfig contains methods used by this package.
// You can implement these to extend or modify the behaviour of this package
type CustomErrorConfig interface {
	// Used when the error should be written to the HTTP reqeust
	Write(err ErrorResponse, writer http.ResponseWriter, r *http.Request)

	// Is called when a panic occured during processing a request
	HandlePanic(err any, trace string, w http.ResponseWriter, r *http.Request)

	// Returns a logger instance based on the provided dependency
	GetLoggerFromDependendency(dep any) *logger.Logger
}

// DefaultConfig is a struct that implements the default config
// for "CustomErrorconfig"
type DefaultConfig struct{}

func (c DefaultConfig) Write(err ErrorResponse, writer http.ResponseWriter, r *http.Request) {
	response.WriteText(err.Message, err.Status, writer)
}

func (c DefaultConfig) HandlePanic(err any, trace string, w http.ResponseWriter, r *http.Request) {
	// Try to parse it to an error response (the error occured probably in awareness of the developer :)
	if errResponse, ok := err.(ErrorResponse); ok {
		errResponse.Write(w, r)
		return
	}

	// Log error and write header
	logger.Error("Error: %s", fmt.Errorf("%s", err))
	w.WriteHeader(500)
	w.Header().Set("Connection", "close")

	// Write debug trace
	logger.Debug(trace)
}

func (c DefaultConfig) GetLoggerFromDependendency(dep any) *logger.Logger {
	return logger.GetGlobalLogger()
}

// Error is an interface around [ErrorResponse] that
// you can use instead of [ErrorResponse] to support nil
// values
type Error interface {
	GetErrorStruct() ErrorResponse
	Error() string
}

var _ Error = ErrorResponse{}

// ErrorResponse represents an error which occured during the run
// of the application
type ErrorResponse struct {
	Status  int
	Message string `json:"message"`

	// Internal and detailed error message of the problem
	InternalMessage string

	Data any `json:"-"`
}

// New is a wrapper around [errors.New] that
// returns a distinct error value even if the message
// is identical
func New(message string) error {
	return errors.New(message)
}

// NewError creates a new ErrerResponse with the provided
// message and status code that is returned to the user
func NewError(message string, statusCode int) ErrorResponse {
	return ErrorResponse{
		Status:  statusCode,
		Message: message,
	}
}

// The requested ressource was not found
func NotFound() ErrorResponse {
	return ErrorResponse{
		Status:  404,
		Message: "The requested resource was not found",
	}
}

// Request is in a bad format
func BadRequest(message string) ErrorResponse {
	if message == "" {
		message = "Your request is in a bad format"
	}

	return ErrorResponse{
		Status:  400,
		Message: message,
	}
}

// No response data
func NoContent() ErrorResponse {
	return ErrorResponse{
		Status:  204,
		Message: "",
	}
}

// Internal server error
func InternalError() ErrorResponse {
	return ErrorResponse{
		Status:  500,
		Message: "We encountered an error while processing your request",
	}
}

// A ressource with the same id, name, ... does already exists -> conflict
func AlreadyExists(message string) ErrorResponse {
	if message == "" {
		message = "A ressource with the same data already exists"
	}

	return ErrorResponse{
		Status:  409,
		Message: message,
	}
}

// Returns a default error message for the status code
func (err ErrorResponse) GetDefaultErrorMessage(statusCode int) string {
	return http.StatusText(statusCode)
}

// Write writes the error to the request.
// Note that after calling this method no additional write to the
// response is allowed
func (err ErrorResponse) Write(writer http.ResponseWriter, r *http.Request) {
	if err.Status == 0 {
		err = ErrorResponse{
			Status:  500,
			Message: "We encountered an error while processing your request",
		}
	}

	Config.Write(err, writer, r)
}

// Attach attaches a custom dependency to the error struct
// that can later be used to customize the behaviour with the
// "Config" variable.
//
// Because "data" is copied around some times, it should be a pointer
func (err ErrorResponse) Attach(data any) ErrorResponse {
	err.Data = data

	return err
}

// Write tries to convert the error to an ErrorResponse
// or writes an 500 Request if an generic error was provided
func Write(writer http.ResponseWriter, r *http.Request, err error) {
	e, ok := err.(ErrorResponse)
	if !ok {
		e = ErrorResponse{
			Status:  500,
			Message: "We encountered an error while processing your request",
		}
	}

	e.Write(writer, r)
}

func (err ErrorResponse) Error() string {
	return err.Message
}

func (err ErrorResponse) GetErrorStruct() ErrorResponse {
	return err
}

// Log logs the given error with a logger that is obtained from [Config.(dep)]
// and returns this object.
// Msg is used as a prefix before the error message
func (err ErrorResponse) Log(msg string, e error, dep any, args ...any) ErrorResponse {
	if msg != "" {
		msg += ": "
	}
	msg += "%s"

	// Get logger to log with
	log := Config.GetLoggerFromDependendency(dep)
	log = logger.CloneLogger(log)
	log.FuncCallIncrement = log.FuncCallIncrement + 1

	// Write the message aut
	args = append(args, e)
	log.Error(msg, args...)

	return err
}

// Sprintf replaces the internal message of this error
// with [fmt.Sprintf] and returns it.
// The original error won't be modified!
func (err ErrorResponse) Sprintf(vals ...any) ErrorResponse {
	err.Message = fmt.Sprintf(err.Message, vals...)
	return err
}
