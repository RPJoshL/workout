package database

import (
	"database/sql"
	"reflect"
	"strings"

	"github.com/RPJoshL/go-logger"
)

// mappDbColumns initializes an array of pointers that are pointing to the matching
// fields of the given dst that are looked up by the 'db' tag.
// This result can be used for "rows.Scan()"
func mappDbColumns(dst reflect.Value, columns []string) []any {
	// Create internal reflection values
	val := dst.Elem()
	structType := val.Type()
	mappedColumns := make([]any, len(columns))

	// Loop through every column and find a matching tag of the struct.
	for colNr, name := range columns {
		mappedValue := findInStruct(dst, name)
		if mappedValue != nil {
			mappedColumns[colNr] = mappedValue
			continue
		}

		// We should print a debug message if a database column couldn't be mapped to a
		// struct field
		logger.Info("Didn't found a matching tag for %q inside struct %q", name, structType.Name())

		// Because otherwise the query would fail, we add a pointer to a value that doesn't mapp to a field
		// of dst. We use a string for that
		noMappedValue := reflect.New(reflect.PointerTo(reflect.TypeFor[sql.NullString]()))
		mappedColumns[colNr] = noMappedValue.Interface()
	}

	return mappedColumns
}

func findInStruct(ref reflect.Value, name string) any {
	refType := ref.Type()
	if refType.Kind() == reflect.Pointer {
		refType = refType.Elem()
		ref = ref.Elem()
	}

	// Find column with tag
	for i := range refType.NumField() {
		fieldType := refType.Field(i)

		tag, exists := refType.Field(i).Tag.Lookup("db")
		if exists && strings.EqualFold(tag, name) {
			return ref.Field(i).Addr().Interface()
		}

		// When there is no tag available, we try to match the column by the fields name
		fieldName := refType.Field(i).Name
		if strings.EqualFold(fieldName, name) || strings.EqualFold(fieldName, strings.ReplaceAll(name, "_", "")) {
			return ref.Field(i).Addr().Interface()
		}

		if fieldType.Anonymous && fieldType.Type.Kind() == reflect.Struct {
			res := findInStruct(ref.Field(i).Addr(), name)
			if res != nil {
				return res
			}
		}
	}

	return nil
}
