package dbstruct

import (
	"fmt"
	"reflect"

	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

// isPointer returns weather "val" is a pointer to the given type
func isPointer(ref reflect.Value, typ reflect.Kind) error {
	if ref.Type().Kind() != reflect.Pointer {
		return errors.New("no pointer")
	} else if ref.IsNil() {
		return errors.New("nil pointer")
	} else if ref.Elem().Type().Kind() != typ {
		return fmt.Errorf("no %s", typ.String())
	}

	return nil
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
