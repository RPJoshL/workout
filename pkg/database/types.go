package database

import (
	"database/sql"
	"encoding/json"
)

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

// NullInt is a wrapper around sql.NullInt64
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
