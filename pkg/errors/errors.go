// Package errors provides an error type to handle and return errors for API reqeuests. It is a wrapper
// around the generic error interface extended with additional information and methods.
//
// It can be used from API endpoints to return custom
// messages with an HTTP ResponseWriter
package errors

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strings"

	"git.rpjosh.de/RPJosh/workout/pkg/response"
	"github.com/RPJoshL/go-logger"
)

// pointerStruct is a simple helper struct to check if the error instance
// equals another error instance
type pointerStruct struct {
	// We need at least a single value. Otherwise the pointer address
	// would always be the same...
	PointerUnique string
}

type Translator interface {
	Get(key string) string
}

// Config is a global static variable that you can use to customize some aspects
// in this package.
// See "CustomErrorConfig" for available methods and a detailed description
var Config CustomErrorConfig = DefaultConfig{}

// CustomErrorConfig contains methods used by this package.
// You can implement these to extend or modify the behaviour of this package
type CustomErrorConfig interface {
	// Used when the error should be written to the HTTP reqeust
	Write(err ErrorResponse, writer http.ResponseWriter, r *http.Request)

	// Is called when a panic occurred during processing a request
	HandlePanic(err any, trace string, w http.ResponseWriter, r *http.Request)

	// Returns a logger instance based on the provided dependency
	GetLoggerFromDependendency(dep any) *logger.Logger

	// Returns the translated message for errors beginning with a "#"
	GetEnTranslation(key string) string
}

// DefaultConfig is a struct that implements the default config
// for "CustomErrorconfig"
type DefaultConfig struct{}

func (c DefaultConfig) Write(err ErrorResponse, writer http.ResponseWriter, r *http.Request) {
	err.WriteHeaders(writer)
	response.WriteText(err.Message, err.Status, writer)
}

func (c DefaultConfig) HandlePanic(err any, trace string, w http.ResponseWriter, r *http.Request) {
	// Try to parse it to an error response (the error occurred probably in awareness of the developer :)
	if errResponse, ok := err.(ErrorResponse); ok {
		errResponse.Write(w, r)
		return
	}

	// Log error and write header
	logger.Error("Error: %s", fmt.Errorf("%s", err))
	w.WriteHeader(500)
	w.Header().Set("Connection", "close")

	// Write debug trace
	logger.Debug("%s", trace)
}

func (c DefaultConfig) GetLoggerFromDependendency(dep any) *logger.Logger {
	return logger.GetGlobalLogger()
}

func (c DefaultConfig) GetEnTranslation(key string) string {
	return key
}

// Error is an interface around [ErrorResponse] that
// you can use instead of [ErrorResponse] to support nil
// values
type Error interface {
	GetErrorStruct() ErrorResponse
	Error() string
}

var _ Error = ErrorResponse{}

// ErrorResponse represents an error which occurred during the run
// of the application.
//
//nolint:all
type ErrorResponse struct {
	Status  int
	Message string `json:"message"`

	// Internal and detailed error message of the problem
	InternalMessage string

	Data any `json:"-"`

	// Problem: when using sprintf, we don't have a reference to an
	// translator => store original value
	messageOrig string
	sprintfVals []any

	// Headers to be added to each call of [write]
	headers map[string]string

	// Reference to unique identify this error instance
	ref *pointerStruct
}

// New is a wrapper around [errors.New] that
// returns a distinct error value even if the message
// is identical
func New(message string) error {
	return errors.New(message)
}

// NewError creates a new ErrerResponse with the provided
// message and status code that is returned to the user.
// Each value created with [NewError] is distinct.
func NewError(message string, statusCode int) ErrorResponse {
	return ErrorResponse{
		Status:  statusCode,
		Message: message,
		ref:     &pointerStruct{},
		headers: map[string]string{},
	}
}

func NotFound() ErrorResponse {
	return ErrorResponse{
		Status:  404,
		Message: "The requested resource was not found",
		ref:     &pointerStruct{},
	}
}

func BadRequest(message string) ErrorResponse {
	if message == "" {
		message = "Your request is in a bad format"
	}

	return ErrorResponse{
		Status:  400,
		Message: message,
		ref:     &pointerStruct{},
	}
}

// NoContent identifies an empty response without a body / data
func NoContent() ErrorResponse {
	return ErrorResponse{
		Status:  204,
		Message: "",
		ref:     &pointerStruct{},
	}
}

func InternalError() ErrorResponse {
	return ErrorResponse{
		Status:  500,
		Message: "We encountered an error while processing your request",
		ref:     &pointerStruct{},
	}
}

// AlreadyExists returns error response for a
// ressource with the same id, name, ... -> conflict
func AlreadyExists(message string) ErrorResponse {
	if message == "" {
		message = "A ressource with the same data already exists"
	}

	return ErrorResponse{
		Status:  409,
		Message: message,
		ref:     &pointerStruct{},
	}
}

// GetDefaultErrorMessage returns a standard error message for the status code
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
			ref:     &pointerStruct{},
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
	e, ok := GetAs[ErrorResponse](err)
	if !ok {
		e = ErrorResponse{
			Status:  500,
			Message: "We encountered an error while processing your request",
			ref:     &pointerStruct{},
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
	// Get logger to log with
	log := Config.GetLoggerFromDependendency(dep)
	log = logger.CloneLogger(log)
	log.FuncCallIncrement++

	if e != nil {
		if msg != "" {
			msg += ": "
		}
		msg += "%s"

		errMessage := e.Error()
		// Translate any error
		if strings.HasPrefix(errMessage, "#") && len(errMessage) > 1 {
			errMessage = Config.GetEnTranslation(errMessage[1:])
		}

		args = append(args, errMessage)
	}

	log.Error(msg, args...)

	return err
}

// Sprintf replaces the internal message of this error
// with [fmt.Sprintf] and returns it.
// The original error won't be modified!
func (err ErrorResponse) Sprintf(vals ...any) ErrorResponse {
	err.messageOrig = err.Message
	err.sprintfVals = vals

	err.Message = fmt.Sprintf(err.Message, vals...)
	return err
}

// ApplySprintf translates the provided value with [trans]
// (if starting with a "#"), and applies any previously provided
// placeholders to the translated value
func (err ErrorResponse) ApplySprintf(trans Translator) ErrorResponse {
	// Only translate message if starting with "#"
	if !strings.HasPrefix(err.Message, "#") {
		return err
	}

	// Get original message (if modified by [Sprintf])
	origMessage := err.Message
	if len(err.sprintfVals) > 0 {
		origMessage = err.messageOrig
	}
	err.Message = trans.Get(origMessage[1:])

	// Apply sprintf
	if len(err.sprintfVals) > 0 {
		err.Message = fmt.Sprintf(err.Message, err.sprintfVals...)
	}

	return err
}

// WithHeader returns a clone of this error response
// with the provided headers attached
func (err ErrorResponse) WithHeader(name, value string) ErrorResponse {
	rtc := err.clone()
	rtc.headers[name] = value
	return rtc
}

// WriteHeaders writes all the previously set header to the provided
// response
func (err ErrorResponse) WriteHeaders(resp http.ResponseWriter) {
	for key, val := range err.headers {
		resp.Header().Set(key, val)
	}
}

func (err ErrorResponse) clone() ErrorResponse {
	rtc := ErrorResponse{
		messageOrig:     err.messageOrig,
		ref:             err.ref,
		sprintfVals:     err.sprintfVals,
		Status:          err.Status,
		Message:         err.Message,
		InternalMessage: err.InternalMessage,
		Data:            err.Data,
		headers:         map[string]string{},
	}

	// Clone headers
	maps.Copy(rtc.headers, err.headers)

	return rtc
}

// Is checks if "a" is the same instance as the provided value of "b".
//
// Even if this error value was "modified" with [Sprintf], this methode
// will still return "true".
//
// Wrapped errors are also supported by this function.
//
// If a or b are nil, "false" will be return
func Is(a error, b Error) bool {
	if a == nil || b == nil {
		return false
	}

	// Extract also a possible wrapped error
	if e, ok := GetAs[Error](a); ok {
		return e.GetErrorStruct().ref == b.GetErrorStruct().ref
	}

	return false
}

// IsGeneric is a wrapper around [errors.Is] if one of the errors is not of the type [Error]
func IsGeneric(err, target error) bool {
	var errConc Error
	var targetConc Error

	if errors.As(err, &errConc) && errors.As(target, &targetConc) {
		return Is(errConc, targetConc)
	}

	return errors.Is(err, target)
}

// IsNot checks if "a" is not the same instance as the provided value of "b".
//
// Even if this error value was "modified" with [Sprintf], this methode
// will still return "false".
//
// If a or b are nil, "true" will be return
func IsNot(a, b Error) bool {
	return !Is(a, b)
}

// As is a wrapper around [errors.As]
func As(err error, target any) bool {
	return errors.As(err, target)
}

// GetAs is a wrapper around [errors.As] that writes it result
// into "rtc" directly
func GetAs[T error](err error) (rtc T, found bool) {
	found = errors.As(err, &rtc)
	return
}

func Wrap(err error, message string, args ...any) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", fmt.Sprintf(message, args...), err)
}
