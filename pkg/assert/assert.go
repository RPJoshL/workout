// Package assert provides various comparison functions
// for unit tests
package assert

import (
	"testing"

	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"github.com/google/go-cmp/cmp"
)

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

// EqualStruct compares two structs with each other and prints the diff.
// It returns whether the two structs are the same
func EqualStruct(t *testing.T, subject string, expected, got any, opts ...cmp.Option) bool {
	if diff := cmp.Diff(expected, got, opts...); diff != "" {
		t.Errorf("Mismatch of %s(-expected +got):\n%s", subject, diff)
		return false
	}

	return true
}

// Equal compares two simple comparable types with each other
func Equal[T comparable](t *testing.T, expected, got T, messages ...string) {
	if expected == got {
		return
	}

	message := ""
	if len(messages) > 0 {
		message = messages[0] + ". "
	}

	t.Errorf("%sExpected %v, got %v", message, expected, got)
}

// Require compares two simple comparable types with each other and exits
// the current test if they are not equal
func Require[T comparable](t *testing.T, expected, got T, messages ...string) {
	if expected == got {
		return
	}

	message := ""
	if len(messages) > 0 {
		message = messages[0] + ". "
	}

	t.Fatalf("%sExpected %v, got %v", message, expected, got)
}

func ErrorIs(t *testing.T, err, target error) {
	t.Helper()

	if !errors.IsGeneric(err, target) {
		t.Errorf("Expected error %v, got %v", target, err)
	}
}
