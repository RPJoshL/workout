package router

import (
	"net/http"
	"reflect"
	"strconv"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/api/utils"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

var (
	ErrNoPointerStruct     = errors.InternalError()
	ErrIntValue            = errors.BadRequest("Invalid numeric value %q for field %q")
	ErrTimeValue           = errors.BadRequest("Invalid time format provided: %q. Expected ISO8601")
	ErrUnsupportedDataType = errors.InternalError()
)

// RequestParseMode specifies which tag values should be used
type RequestParseMode int8

const (
	ParseModeQuery RequestParseMode = iota
	ParseModeForm

	// Tries to get a non-empty value base on the query and form tag
	// in the above order
	ParseModeAll
)

const (
	// Tag identifier for query parameters
	TagQuery = "query"

	// Tag identifier for form values
	TagForm = "form"

	// Tag identifier for the default json tag
	TagJson = "json"
)

// RequestParser parses request data to a generic struct
// with tags defined in this request
type RequestParser struct {
	Request *http.Request
}

// RequestParserOptions contains various options to modify the
// behaviour of the request parsing
type RequestParserOptions struct {
	// From which tag and request field the values should be parsed
	Mode RequestParseMode

	// Use the key from the json tag to parse the request values based on the specified
	// parsing mode
	InterpreteJson bool
}

// Parse parses the request data into the provided *struct{}.
//
// The returned error is expected to be passed to the client and already
// logged in this package.
//
// The form of the HTTP request is automatically parsed with a limit of 1 Mbyte
func (p *RequestParser) Parse(dst any, opt RequestParserOptions) errors.Error {

	// Parse form data
	if opt.Mode == ParseModeAll || opt.Mode == ParseModeForm {
		if err := p.Request.ParseMultipartForm(utils.MToBytes(1)); err != nil {
			logger.Warning("Failed to parse form content: %s", err)
			return errors.BadRequest("Failed to parse form content")
		}
	}

	return p.parse(reflect.ValueOf(dst), opt)
}

// parse is an internal wrapper that accepts a reflection value instead of a
// raw struct
func (p *RequestParser) parse(dstVal reflect.Value, opt RequestParserOptions) errors.Error {
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
			if err := p.parse(field.Addr(), opt); err != nil {
				return err
			}
		}

		var value any
		var error errors.Error

		// Ignore special fields
		if fieldType.Name == "DbMetadata_" {
			continue
		}

		// Parse value if any tag is present
		if val, found := p.getValueForMode(fieldType.Tag, opt); found {
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

// getValueForMode returns the raw values to parse based on the provided mode
// and weather JSON tags should be used
func (p *RequestParser) getValueForMode(tag reflect.StructTag, opt RequestParserOptions) (val []string, tagFound bool) {
	tagJson := ""
	if opt.InterpreteJson {
		tagJson = tag.Get(TagJson)
	}
	tagQuery := tag.Get(TagQuery)
	tagForm := tag.Get(TagForm)

	// Query
	if opt.Mode == ParseModeQuery || opt.Mode == ParseModeAll {
		tag := getFirstNonEmptyValue(tagJson, tagQuery)

		if tag != "" {
			found := false
			tagFound = true

			if val, found = p.Request.URL.Query()[tag]; found {
				return val, tagFound
			} else {
				// Fallback to array (axios)
				val = p.Request.URL.Query()[tag+"[]"]
			}
		}

		if len(val) != 0 {
			return val, tagFound
		}
	}

	// Form
	if opt.Mode == ParseModeForm || opt.Mode == ParseModeAll {
		tag := getFirstNonEmptyValue(tagJson, tagForm)

		if tag != "" {
			tagFound = true

			if val, found := p.Request.Form[tag]; found {
				return val, tagFound
			}
		}
	}

	return []string{}, tagFound
}

func getFirstNonEmptyValue(values ...string) string {
	for _, val := range values {
		if val != "" {
			return val
		}
	}

	return ""
}

// validateAndConvert validates the user input based on various rules
// and converts it to the struct's data type.
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
