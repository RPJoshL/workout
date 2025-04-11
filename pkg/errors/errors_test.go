package errors

import (
	"errors"
	"fmt"
	"testing"
)

var (
	ErrTestA       = NewError("#workout.egon", 100)
	ErrTestB       = NewError("#workout.maria", 100)
	ErrTestWrapped = NewError("Wrapped error", 200)
)

func TestEqual(t *testing.T) {
	a := NewError("hello", 1)
	b := NewError("hello", 1)

	if errors.Is(a, b) {
		t.Errorf("Errors shouldn't match")
	}

	aa := a.Sprintf("dd")
	if IsNot(aa, a) {
		t.Errorf("Errors should be the same")
	}
	// Message should still be the same
	if a.Message != "hello" {
		t.Errorf("Message changed")
	}
}

func TestEqualConst(t *testing.T) {
	if Is(ErrTestA, ErrTestB) {
		t.Errorf("Errors shouldn't match")
	}

	errGot := ErrTestA.Sprintf("ola")
	if IsNot(errGot, ErrTestA) {
		t.Errorf("Errors should be same")
	}

	if Is(errGot, ErrTestB) {
		t.Errorf("Errors shouldn't match")
	}
}

func TestEqualWrapped(t *testing.T) {
	wrapped := fmt.Errorf("Wrapping it: %w", ErrTestWrapped)

	if !Is(wrapped, ErrTestWrapped) {
		t.Error("Errors should match")
	}

	if Is(wrapped, ErrTestA) {
		t.Error("Errors should not match")
	}
}
