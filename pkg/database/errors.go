package database

import "git.rpjosh.de/RPJosh/workout/pkg/errors"

var _ errors.Error = DatabaseError{}

// ErrorType defines the type of the error in a database context
type ErrorType int

const (
	UnexpectedError = iota
	NoRows
	TooManyRows
)

func (t ErrorType) String() string {
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

// Error extends the default error interface
// to provide additional information why the query failed
type Error interface {
	error

	// Type returns the type of the error in a database context
	Type() ErrorType

	// GetResponse returns an error response for the client
	GetResponse() errors.ErrorResponse

	// GetError returns the internal error of this DatabasError
	GetError() error
}

// Make sure that DatabaseErrorStruct implements database error
var _ Error = DatabaseError{}

type DatabaseError struct {
	Response errors.ErrorResponse
	Err      error
	Typ      ErrorType
}

func (e DatabaseError) Error() string {
	return e.Err.Error()
}
func (e DatabaseError) Type() ErrorType {
	return e.Typ
}
func (e DatabaseError) GetResponse() errors.ErrorResponse {
	return e.Response
}
func (e DatabaseError) GetErrorStruct() errors.ErrorResponse {
	return e.Response
}
func (e DatabaseError) GetError() error {
	return e.Err
}
