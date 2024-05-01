package database

import (
	"database/sql"
	"encoding/json"
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

// NewNullString returns a new (non-null) NullString
// with the provided value
func NewNullString(val string) sql.NullString {
	return sql.NullString{
		Valid:  true,
		String: val,
	}
}

func NewNullInt(val int) sql.NullInt64 {
	return sql.NullInt64{
		Valid: true,
		Int64: int64(val),
	}
}

// NullString is a wrapper around sql.NullString
type NullString sql.NullString

// NullString is a wrapper around sql.NullInt64
type NullInt sql.NullInt64

// MarshalJSON method is called by json.Marshal,
// whenever it is of type NullString
func (x *NullString) MarshalJSON() ([]byte, error) {
	if !x.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(x.String)
}

// MarshalJSON method is called by json.Marshal,
// whenever it is of type NullString
func (x *NullInt) MarshalJSON() ([]byte, error) {
	if !x.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(x.Int64)
}
