package database

import (
	"database/sql"
	"reflect"
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
)

// mappDbColumns initializes an array of pointers that are pointing to the matching
// fields of the given dst that are looked up by the 'db' tag.
// This result can be used for "rows.Scan()"
func mappDbColumns(dst reflect.Value, columns []string) []interface{} {

	// Create internal reflection values
	val := dst.Elem()
	structType := val.Type()
	mappedColumns := make([]interface{}, len(columns))

	// Loop through every column and find a matching tag of the struct.
outer:
	for colNr, name := range columns {

		// Find column with tag
		for i := 0; i < structType.NumField(); i++ {
			tag, exists := structType.Field(i).Tag.Lookup("db")
			if exists && strings.EqualFold(tag, name) {
				mappedColumns[colNr] = val.Field(i).Addr().Interface()
				continue outer
			}

			// When there is no tag available, we try to match the column by the fields name
			fieldName := structType.Field(i).Name
			if strings.EqualFold(fieldName, name) || strings.EqualFold(fieldName, strings.ReplaceAll(name, "_", "")) {
				mappedColumns[colNr] = val.Field(i).Addr().Interface()
				continue outer
			}
		}

		// We should print a debug message if a database column couldn't be mapped to a
		// struct field
		logger.Info("Didn't found a matching tag for %q inside struct %q", name, structType.Name())

		// Because otherwise the query would fail, we add a pointer to a value that doesn't mapp to a field
		// of dst. We use a string for that
		noMappedValue := reflect.New(reflect.PointerTo(reflect.TypeOf(sql.NullString{})))
		mappedColumns[colNr] = noMappedValue.Interface()
	}

	return mappedColumns
}
