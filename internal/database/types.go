package database

import (
	"database/sql"
	"reflect"
)

// isZero reports weather val is the zero value
// for it's type
func isZero(val any) bool {
	if reflect.ValueOf(val).IsZero() {
		return true
	}

	return false
}

// isNull reports weather val is a nil pointer
// or !valid for [sql.]
func isNull(val any) bool {
	if val == nil {
		return true
	}

	switch v := val.(type) {
	case sql.NullString:
		return !v.Valid
	case sql.NullBool:
		return !v.Valid
	case sql.NullByte:
		return !v.Valid
	case sql.NullInt16:
		return !v.Valid
	case sql.NullInt32:
		return !v.Valid
	case sql.NullInt64:
		return !v.Valid
	case sql.NullFloat64:
		return !v.Valid
	case sql.NullTime:
		return !v.Valid
	}

	return false
}
