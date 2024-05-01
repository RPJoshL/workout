package router

import (
	"net/http"
	"reflect"
	"strconv"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/go-webserver/errors"
)

var (
	ErrNoPointerStruct     = errors.InternalError()
	ErrIntValue            = errors.BadRequest("Received invalid numeric %q for field %q")
	ErrUnsupportedDataType = errors.InternalError()
)

// Tag name for query parameters
const TagQuery = "query"

// RequestParser parses request data to a generic struct
// with tags defined in this request
type RequestParser struct {
	Request *http.Request
}

// Parse parses the request data into the provided
// [*struct].
//
// The returned error is expected to be passed to the client and is
// already logged internally
func (p *RequestParser) Parse(dst any) errors.Error {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Pointer || dstVal.Elem().Kind() != reflect.Struct {
		logger.Error("No pointer to a struct given")
		return ErrNoPointerStruct
	}

	// Parse all fields of the destination
	for i := 0; i < dstVal.Elem().NumField(); i++ {
		field := dstVal.Elem().Field(i)
		fieldType := dstVal.Elem().Type().Field(i)

		// Value to set for this field
		var value any
		var error errors.Error

		// Query parameter
		if queryTag := fieldType.Tag.Get(TagQuery); queryTag != "" {
			value, error = p.validateAndConvert(fieldType, p.Request.URL.Query().Get(queryTag))
		}

		// Return erro if failed
		if error != nil {
			return error
		}

		// Set field if not zero
		if value != nil {
			field.Set(reflect.ValueOf(value))
		}
	}

	return nil
}

// validateAndConvert validates the user input based on the rules
// defined in this struct tag and converts it to the filed value
func (p *RequestParser) validateAndConvert(field reflect.StructField, value string) (any, errors.Error) {
	return ConvertStringToType(value, field.Type)
}

func ConvertStringToType(val string, typ reflect.Type) (any, errors.Error) {
	switch typ.Kind() {
	case reflect.String:
		return val, nil
	case reflect.Int:
		// Empty value
		if val == "" {
			return 0, nil
		}

		intVal, err := strconv.Atoi(val)
		if err != nil {
			return nil, ErrIntValue.Sprintf(val, typ.Name())
		}
		return intVal, nil
	}

	logger.Error("Received unsupported data type to convert: %s", typ.Kind())
	return nil, ErrUnsupportedDataType
}
