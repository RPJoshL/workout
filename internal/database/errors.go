package database

import "git.rpjosh.de/RPJosh/workout/pkg/errors"

var _ errors.Error = databaseErr{}

// Error defines the type of the error in a database context
type Error int

const (
	UnexpectedError = 0
	NoRows          = iota
	TooManyRows
)

func (t Error) String() string {
	switch t {
	case 0:
		return "Unexpected error"
	case 1:
		return "Received no rows"
	case 2:
		return "Received more than a single row"
	default:
		return "Unknown error"
	}
}

// DatabaseError extends the default error interface
// to provide additional information why the query failed
type DatabaseError interface {
	error

	// Type returns the type of the error in a database context
	Type() Error

	// GetResponse returns an error response for the client
	GetResponse() errors.ErrorResponse

	// GetError returns the internal error of this DatabasError
	GetError() error
}

// Make sure that DatabaseErrorStruct implements database error
var _ DatabaseError = databaseErr{}

type databaseErr struct {
	Typ      Error
	Err      error
	Response errors.ErrorResponse
}

func (e databaseErr) Error() string {
	return e.Err.Error()
}
func (e databaseErr) Type() Error {
	return e.Typ
}
func (e databaseErr) GetResponse() errors.ErrorResponse {
	return e.Response
}
func (e databaseErr) GetErrorStruct() errors.ErrorResponse {
	return e.Response
}
func (e databaseErr) GetError() error {
	return e.Err
}
