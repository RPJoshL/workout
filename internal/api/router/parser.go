package router

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/api/utils"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"github.com/guregu/null/v5"
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

	// Weather to search for tags also in child structs.
	// The form key is expected to be separated by a "." (eg. "TagNameRoot.TagNameChild").
	// Also pointers to structs are resolved and set if at least a single
	// field is present.
	// Because the tag value is used as a prefix for child structs, it will not work
	// for ParseModeAll
	Recursive bool
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

	err, _ := p.parse(reflect.ValueOf(dst), opt, "")
	return err
}

// parse is an internal wrapper that accepts a reflection value instead of a
// raw struct
func (p *RequestParser) parse(dstVal reflect.Value, opt RequestParserOptions, subkey string) (err errors.Error, valueFound bool) {
	if dstVal.Kind() != reflect.Pointer || dstVal.Elem().Kind() != reflect.Struct {
		logger.Error("No pointer to a struct given")
		return ErrNoPointerStruct, valueFound
	}

	// Parse all fields of the destination
	for i := range dstVal.Elem().NumField() {
		field := dstVal.Elem().Field(i)
		fieldType := dstVal.Elem().Type().Field(i)

		// Check for embedded struct
		if fieldType.Anonymous && fieldType.Type.Kind() == reflect.Struct {
			if err, found := p.parse(field.Addr(), opt, subkey); err != nil {
				return err, found
			}
		}

		var value any
		var err errors.Error

		// Ignore special fields
		if fieldType.Name == "DbMetadata_" {
			continue
		}

		// Parse value if any tag is present
		if val, found, tagName := p.getValueForMode(fieldType.Tag, opt, subkey); found {
			value, err = p.validateAndConvert(fieldType, val, opt)
			isStruct := false
			if !fieldType.Anonymous && opt.Recursive {
				if fieldType.Type.Kind() == reflect.Struct {
					isStruct = true
					if err, _ := p.parse(field.Addr(), opt, subkey+tagName+"."); err != nil {
						return err, valueFound
					}
				} else if fieldType.Type.Kind() == reflect.Pointer && fieldType.Type.Elem().Kind() == reflect.Struct {
					isStruct = true

					// Create a new value of that type. We only set it if it least one field was present
					childVal := reflect.New(fieldType.Type.Elem())
					if err, found := p.parse(childVal, opt, subkey+tagName+"."); err != nil {
						return err, found
					} else if found {
						// Only set a non nil value if at least one value was found
						field.Set(childVal)
					}
				} else if len(val) > 0 {
					valueFound = true
				}
			} else if len(val) > 0 {
				valueFound = true
			}

			// Reset value getter error
			if errors.Is(err, ErrUnsupportedDataType) && isStruct {
				err = nil
			}
		}

		// Return error if failed
		if err != nil {
			return err, valueFound
		}

		// Set field if not zero
		if value != nil {
			field.Set(reflect.ValueOf(value))
		}
	}

	return nil, valueFound
}

// getValueForMode returns the raw values to parse based on the provided mode
// and weather JSON tags should be used
func (p *RequestParser) getValueForMode(tagRef reflect.StructTag, opt RequestParserOptions, prefix string) (val []string, tagFound bool, tag string) {
	tagJson := ""
	if opt.InterpreteJson {
		tagJson = tagRef.Get(TagJson)
	}
	tagQuery := tagRef.Get(TagQuery)
	tagForm := tagRef.Get(TagForm)

	// Query
	if opt.Mode == ParseModeQuery || opt.Mode == ParseModeAll {
		tag = getFirstNonEmptyValue(tagJson, tagQuery)

		if tag != "" {
			found := false
			tagFound = true

			if val, found = p.Request.URL.Query()[prefix+tag]; found {
				return val, tagFound, tag
			} else {
				// Fallback to array (axios)
				val = p.Request.URL.Query()[tag+"[]"]
			}
		}

		if len(val) != 0 {
			return val, tagFound, tag
		}
	}

	// Form
	if opt.Mode == ParseModeForm || opt.Mode == ParseModeAll {
		tag = getFirstNonEmptyValue(tagJson, tagForm)

		if tag != "" {
			tagFound = true

			if val, found := p.Request.Form[prefix+tag]; found {
				return val, tagFound, tag
			}
		}
	}

	return []string{}, tagFound, tag
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
func (p *RequestParser) validateAndConvert(field reflect.StructField, value []string, opt RequestParserOptions) (any, errors.Error) {
	return p.ConvertStringToType(value, field.Type, opt)
}
func (p *RequestParser) ConvertStringToType(valArr []string, typ reflect.Type, opt RequestParserOptions) (any, errors.Error) {
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
	case reflect.Float64:
		// Empty value
		if val == "" {
			return 0.0, nil
		}

		if floatVal, err := strconv.ParseFloat(val, 64); err != nil {
			return nil, ErrIntValue.Sprintf(val, typ.Name())
		} else {
			return floatVal, nil
		}
	case reflect.Struct:
		switch typ {
		case reflect.TypeOf(time.Time{}):
			// Expect it in ISO format
			if tim, err := time.Parse(time.RFC3339, val); err != nil {
				return nil, ErrTimeValue.Sprintf(val)
			} else {
				return tim, nil
			}

		case reflect.TypeOf(null.Int64{}):
			if val == "" {
				return null.Int64{}, nil
			} else {
				intVal, err := strconv.Atoi(val)
				if err != nil {
					return nil, ErrIntValue.Sprintf(val, typ.Name())
				}

				return null.NewInt(int64(intVal), true), nil
			}
		}
	case reflect.Slice:
		rtc := reflect.MakeSlice(typ, 0, 0)
		for _, v := range valArr {
			if convValue, convErr := p.validateAndConvert(reflect.StructField{Type: typ.Elem()}, []string{v}, opt); convErr != nil {
				return nil, convErr
			} else {
				rtc = reflect.Append(rtc, reflect.ValueOf(convValue))
			}
		}

		return rtc.Interface(), nil
	case reflect.Bool:
		val = strings.ToLower(val)
		return val == "1" || val == "true" || val == "ja", nil
	}

	if !opt.Recursive || (typ.Kind() != reflect.Struct && typ.Kind() != reflect.Pointer) {
		logger.Warning("Received unsupported data type to convert: %s", typ.Kind())
	}

	return nil, ErrUnsupportedDataType
}
