package dbstruct

import (
	"fmt"
	"reflect"

	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

// isPointer returns weather "val" is a pointer to the given type.
// When a nil pointer is given, it will allocate a new value of the given type
// and set the pointer to it, if allocateNil is true
func isPointer(value any, typ reflect.Kind, allocateNil bool) (rtc reflect.Value, err error) {
	ref := reflect.ValueOf(value)

	if ref.Type().Kind() != reflect.Pointer {
		return rtc, errors.New("no pointer")
	} else if ref.IsNil() {
		return rtc, errors.New("nil pointer")
	}

	// Allocate new value if it's a nil pointer and requested
	if allocateNil && ref.Elem().Kind() == reflect.Pointer && ref.Elem().IsNil() {
		allocateNewPointer(ref)
		// We expect a pointer overall
		ref = ref.Elem()
	}

	if ref.Type().Elem().Kind() != typ {
		return rtc, fmt.Errorf("no %s. Got %s", typ.String(), ref.Type().Elem().Kind().String())
	}

	return ref, nil
}

func allocateNewPointer(ptrToPointer reflect.Value) {
	val := ptrToPointer.Elem()
	newVal := reflect.New(val.Type().Elem())
	ptrToPointer.Elem().Set(newVal)
}

// isPointerType returns weather "val" is a pointer to the given type
func isPointerType(ref reflect.Type, typ reflect.Kind) error {
	if ref.Kind() != reflect.Pointer {
		return errors.New("no pointer")
	} else if ref.Elem().Kind() != typ {
		return fmt.Errorf("no %s", typ.String())
	}

	return nil
}

// isZero reports weather val is the zero value
// for it's type
func isZero(val any) bool {
	refValue := reflect.ValueOf(val)
	return !refValue.IsValid() || refValue.IsZero()
}
