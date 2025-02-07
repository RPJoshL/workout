package dbstruct

import (
	"fmt"
	"reflect"
)

// isPointer returns weather "val" is a pointer to the given type
func isPointer(ref reflect.Value, typ reflect.Kind) error {
	if ref.Type().Kind() != reflect.Pointer {
		return fmt.Errorf("no pointer")
	} else if ref.IsNil() {
		return fmt.Errorf("nil pointer")
	} else if ref.Elem().Type().Kind() != typ {
		return fmt.Errorf("no %s", typ.String())
	}

	return nil
}

// isPointerType returns weather "val" is a pointer to the given type
func isPointerType(ref reflect.Type, typ reflect.Kind) error {
	if ref.Kind() != reflect.Pointer {
		return fmt.Errorf("no pointer")
	} else if ref.Elem().Kind() != typ {
		return fmt.Errorf("no %s", typ.String())
	}

	return nil
}

// isZero reports weather val is the zero value
// for it's type
func isZero(val any) bool {
	return reflect.ValueOf(val).IsZero()
}
