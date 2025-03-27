// Package assert provides various comparison functions
// for unit tests
package assert

import "testing"

// Errorf expects a non nil error. It writes out an message
// with the error and your own "msg" formatted with [fmt.Sprintf]
// if the error is not nil. This function stops your test
func Errorf(t *testing.T, err error, msg string, args ...any) {
	t.Helper()

	if err != nil {
		allArgs := append(args, err)
		t.Fatalf(msg+": %s", allArgs...)
	}
}

// Error expects a non nil error. It writes out the error message
// if the error is not nil. This function stops your test
func Error(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("Error is not nil: %s", err)
	}
}
