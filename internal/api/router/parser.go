package router

import (
	"net/http"
	"reflect"
	"strconv"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

var (
	ErrNoPointerStruct     = errors.InternalError()
	ErrIntValue            = errors.BadRequest("Invalid numeric value %q for field %q")
	ErrTimeValue           = errors.BadRequest("Invalid time format provided: %q. Expected ISO8601")
	ErrUnsupportedDataType = errors.InternalError()
)

// Tag identifier for query parameters
const TagQuery = "query"

// RequestParser parses request data to a generic struct
// with tags defined in this request
type RequestParser struct {
	Request *http.Request
}

// Parse parses the request data into the provided *struct{}.
//
// The returned error is expected to be passed to the client and already
// logged in this package
func (p *RequestParser) Parse(dst any) errors.Error {
	return p.parse(reflect.ValueOf(dst))
}

// parse is an internal wrapper that accepts a reflection value instead of a
// raw struct
func (p *RequestParser) parse(dstVal reflect.Value) errors.Error {
	if dstVal.Kind() != reflect.Pointer || dstVal.Elem().Kind() != reflect.Struct {
		logger.Error("No pointer to a struct given")
		return ErrNoPointerStruct
	}

	// Parse all fields of the destination
	for i := 0; i < dstVal.Elem().NumField(); i++ {
		field := dstVal.Elem().Field(i)
		fieldType := dstVal.Elem().Type().Field(i)

		// Check for embedded struct
		if fieldType.Anonymous && fieldType.Type.Kind() == reflect.Struct {
			if err := p.parse(field.Addr()); err != nil {
				return err
			}
		}

		// Value to set for this field
		var value any
		var error errors.Error

		// Query parameter
		if queryTag := fieldType.Tag.Get(TagQuery); queryTag != "" {
			val, found := p.Request.URL.Query()[queryTag]
			if !found {
				// Fallback to array (axios)
				val = p.Request.URL.Query()[queryTag+"[]"]
			}
			value, error = p.validateAndConvert(fieldType, val)
		}

		// Return error if failed
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

// validateAndConvert validates the user input based on various rules
// and converts it to the structs data type.
//
// Only a selection of datatypes are supported by this function
func (p *RequestParser) validateAndConvert(field reflect.StructField, value []string) (any, errors.Error) {
	return p.ConvertStringToType(value, field.Type)
}
func (p *RequestParser) ConvertStringToType(valArr []string, typ reflect.Type) (any, errors.Error) {
	val := ""
	if len(valArr) > 0 {
		val = valArr[0]
	}

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
	case reflect.TypeOf(time.Time{}).Kind():
		// Expect it in ISO format
		if tim, err := time.Parse(time.RFC3339, val); err != nil {
			return nil, ErrTimeValue.Sprintf(val)
		} else {
			return tim, nil
		}
	case reflect.Slice:
		rtc := reflect.MakeSlice(typ, 0, 0)
		for _, v := range valArr {
			if convValue, convErr := p.validateAndConvert(reflect.StructField{Type: typ.Elem()}, []string{v}); convErr != nil {
				return nil, convErr
			} else {
				rtc = reflect.Append(rtc, reflect.ValueOf(convValue))
			}
		}

		return rtc.Interface(), nil
	}

	logger.Error("Received unsupported data type to convert: %s", typ.Kind())
	return nil, ErrUnsupportedDataType
}
