package response

import (
	"database/sql"
	"encoding/json"
	"maps"
	"reflect"
	"strings"

	"github.com/RPJoshL/go-ddl-parser/structt"
	"github.com/RPJoshL/go-logger"
)

// StructToJSON converts the provided struct or a slice of struct to a JSON string.
//
// The fields can have an extra tag "exJson":
//   - "hide": the field will be ignored in the output
//   - if it does have any other value than "hide", it will be put to
//     the response with the value as the attribute name
//
// Also, the struct field can still be tagged by the default "json" tag.
// If it's present, the field will be put to the response by the tag value.
// Options like "-" or "omitempty" will be considered like in [json], but can be overwritten by "exJson".
//
// For dynamic and programmatically ignoring, you can also specify an array containing the names
// of the fields. A hierarchie cann be expressed with:
//   - root.child.child2
//   - *.child3
//
// The key values are based on the json tags / variable name
func StructToJSON(str any, fieldsToExclude, fieldsToShow []string) any {
	refl := reflect.ValueOf(str)

	if refl.Kind() == reflect.Pointer {
		refl = refl.Elem()
	}
	if refl.Kind() == reflect.Slice && refl.Type().Elem().Kind() == reflect.Struct {
		// Parse an array of struct
		arr := []any{}
		for i := range refl.Len() {
			arr = append(arr, parseStruct(refl.Index(i), refl.Type().Elem(), fieldsToExclude, fieldsToShow, ""))
		}
		return arr
	} else if refl.Kind() != reflect.Struct {
		logger.Error("No struct to convert given: %s", refl.Kind())
		return nil
	}

	return parseStruct(refl, refl.Type(), fieldsToExclude, fieldsToShow, "")
}

// Parses the given struct recursively and returns the map with the values
func parseStruct(str reflect.Value, typ reflect.Type, fieldsToExclude, fieldsToShow []string, root string) any {
	rtc := make(map[string]any)

	// Resolve pointers
	if typ.Kind() == reflect.Pointer {
		if str.IsNil() {
			return nil
		}

		str = str.Elem()
		typ = typ.Elem()
	}

	// Customize the JSON behaviour of special types
	realValue := str.Interface()
	if newValue, transformed := transformValue(realValue); transformed {
		return newValue
	}

	for i := range typ.NumField() {
		structField := typ.Field(i)
		concreteField := str.Field(i)

		tag := structField.Tag.Get("exJson")
		jName, omitEmpty := parseJsonTag(structField.Tag.Get("json"))

		fieldName := getFieldName(root, structField.Name, jName)

		// the exJson tag can overwrite the json name
		if tag != "hide" && tag != "" {
			jName = tag
		}

		// Always include embedded fields. We have to add every field separately
		if structField.Anonymous && isStruct(&structField) {
			rtcMap := parseStruct(concreteField, concreteField.Type(), fieldsToExclude, fieldsToShow, root)
			if rtcMapConc, ok := rtcMap.(map[string]any); ok {
				maps.Copy(rtc, rtcMapConc)
			}
			continue
		}

		// Skip hidden fields
		if hideField(fieldName, tag, fieldsToExclude, fieldsToShow, jName, typ, i) {
			continue
		}

		// Skip zero fields if omitEmpty was specified
		if concreteField.IsZero() && omitEmpty {
			continue
		}

		// Array handling
		if structField.Type.Kind() == reflect.Slice {
			// Whether the element type of the array is a struct or a "value"
			isStruct := structField.Type.Elem().Kind() == reflect.Struct
			arr := []any{}

			for a := range concreteField.Len() {
				if isStruct {
					newVal := parseStruct(concreteField.Index(a), structField.Type.Elem(), fieldsToExclude, fieldsToShow, fieldName)
					arr = append(arr, newVal)
				} else {
					realValue := concreteField.Index(a).Interface()
					if newValue, transformed := transformValue(realValue); transformed {
						arr = append(arr, newValue)
					} else {
						arr = append(arr, realValue)
					}
				}
			}

			rtc[jName] = arr
			continue
		}

		// Raw value we can't handle directly
		if !isStruct(&structField) {
			if concreteField.Type().Kind() == reflect.Pointer {
				rtc[jName] = concreteField.Elem().Interface()
			} else {
				rtc[jName] = concreteField.Interface()
			}
			continue
		}

		if structField.Type.Kind() == reflect.Pointer {
			// If it's a nil pointer, create an empty struct
			if concreteField.IsZero() {
				rtc[jName] = nil
			} else {
				rtc[jName] = parseStruct(concreteField.Elem(), concreteField.Elem().Type(), fieldsToExclude, fieldsToShow, fieldName)
			}
		} else {
			newVal := parseStruct(concreteField, concreteField.Type(), fieldsToExclude, fieldsToShow, fieldName)
			if newMap, ok := newVal.(map[string]any); ok && jName == "~" {
				maps.Copy(rtc, newMap)
			} else {
				rtc[jName] = newVal
			}
		}
	}

	return rtc
}

// transformValue transforms the provided value to a specialized and more
// compatible form for a JSON response.
// Only specific types and the generic marshal interface are transformed
func transformValue(origValue any) (newValue any, transformed bool) {
	switch v := origValue.(type) {
	case sql.NullInt64:
		if v.Valid {
			return v.Int64, true
		} else {
			return nil, true
		}
	case sql.NullString:
		if v.Valid {
			return v.String, true
		} else {
			return nil, true
		}
	case sql.NullTime:
		if v.Valid {
			return v.Time, true
		} else {
			return nil, true
		}
	case json.Marshaler:
		vv, err := v.MarshalJSON()
		if err != nil {
			logger.Warning("Failed to marshal struct: %s", err)
		}
		var vRaw json.RawMessage = vv
		return vRaw, true
	}

	return nil, false
}

// isStruct returns wheather the given field is a struct.
// This can either be a struct type or a pointer to a struct
func isStruct(structField *reflect.StructField) bool {
	return structField.Type.Kind() == reflect.Struct ||
		(structField.Type.Kind() == reflect.Pointer && structField.Type.Elem().Kind() == reflect.Struct)
}

// getFieldName determines the field name with the given hiearchie
func getFieldName(root, name, tag string) string {
	if root == "" || tag == "~" {
		return name
	} else {
		return root + "." + name
	}
}

// parseJsonTag parses a json tag and returns its value
func parseJsonTag(tag string) (fieldName string, hideEmpty bool) {
	tags := strings.Split(tag, ",")

	if len(tags) == 0 {
		return "", false
	}
	return tags[0], len(tags) > 1 && tags[1] == "omitempty"
}

// hideField checks whether the given field should be excluded from the output
func hideField(fieldName, exJson string, fieldsToExclude, fieldsToShow []string, jName string, rootType reflect.Type, fieldIndex int) bool {
	// Exclude based on tag value
	if jName == "" || jName == "-" || exJson == "hide" {
		return !IsInArray(fieldName, fieldsToShow)
	}

	// Custom behaviour for metadata
	field := rootType.Field(fieldIndex)
	fieldTag := structt.FromColumnTag(field.Tag.Get(structt.ColumnTagId))

	if len(fieldsToShow) > 0 {
		// If we are at root level, we always include the field if it has a db metadata tag.
		// It will be filtered later
		if field.Type.Kind() == reflect.Struct {
			if _, exists := field.Type.FieldByName(structt.MetadataFieldName); exists {
				return false
			}
		} else if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct {
			if _, exists := field.Type.Elem().FieldByName(structt.MetadataFieldName); exists {
				return false
			}
		}

		if dbTag, exists := rootType.FieldByName(structt.MetadataFieldName); exists {
			metadata := structt.FromMetadataTag(dbTag.Tag.Get(structt.MetadataTagId))

			// Field has to be contained in [fieldsToShow]
			for _, ex := range fieldsToShow {
				if doesMatchDbTag(ex, metadata, fieldTag, field) {
					return false
				}
			}

			return true
		}
	}

	// We have a db tag to apply custom logic
	if dbTag, exists := rootType.FieldByName(structt.MetadataFieldName); exists {
		metadata := structt.FromMetadataTag(dbTag.Tag.Get(structt.MetadataTagId))

		// Check if field should be included
		for _, ex := range fieldsToExclude {
			if doesMatchDbTag(ex, metadata, fieldTag, field) {
				return true
			}
		}
	} else {
		return IsInArray(fieldName, fieldsToExclude)
	}

	return false
}

// doesMatchDbTag checks whether the provided [fieldName] that was generated
// with go-ddl matches the field metadata of the struct field
func doesMatchDbTag(fieldName string, metadata *structt.MetadataTag, fieldTag *structt.ColumnTag, field reflect.StructField) bool {
	// Get full reference
	if strings.Count(fieldName, "|") == 1 {
		pipe := strings.Index(fieldName, "|")
		colName := fieldName[pipe+1:]

		if colName == metadata.Schema+"."+metadata.Table+"."+fieldTag.Name {
			return true
		}

		// Embedded or referenced structs: include also if a single field should be shown
		if fieldTag.Name == "" && field.Type.Kind() == reflect.Struct && strings.HasPrefix(colName, metadata.Schema+"."+metadata.Table) {
			return true
		}

		// Include all fields of struct if [any] was provided
		if fieldName[:pipe] == "*" && strings.HasPrefix(colName, metadata.Schema+"."+metadata.Table) {
			return true
		}
	} else if fieldName == field.Name {
		return true
	}

	return false
}

// IsInArray Checks if the given value is inside the array.
// Note that *servus (in the array) does match anything ending with 'servus'
func IsInArray(val string, arr []string) bool {
	for _, b := range arr {
		regex := b != "" && b[0:1] == "*"
		if b == val || (regex && strings.HasSuffix(val, b[1:])) {
			return true
		}
	}
	return false
}
