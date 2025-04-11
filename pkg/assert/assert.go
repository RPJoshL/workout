// Package assert provides various comparison functions
// for unit tests
package assert

import "testing"

// NoErrorf expects a non nil error. It writes out an message
// with the error and your own "msg" formatted with [fmt.Sprintf]
// if the error is not nil. This function stops your test
func NoErrorf(t *testing.T, err error, msg string, args ...any) {
	t.Helper()

	if err != nil {
		args = append(args, err)
		t.Fatalf(msg+": %s", args...)
	}
}

// NoError expects a nil error. It writes out the error message
// if the error is not nil. This function stops your test
func NoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("Error is not nil: %s", err)
	}
}
