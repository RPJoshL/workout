package dbstruct

import (
	"reflect"
	"testing"

	"git.rpjosh.de/RPJosh/workout/pkg/assert"
)

type testStruct struct {
	Field1 int
}

func TestIsPointer(t *testing.T) {
	target := testStruct{}

	val, err := isPointer(&target, reflect.Struct, false)
	assert.NoError(t, err)

	val.Elem().FieldByIndex([]int{0}).SetInt(42)

	assert.Equal(t, 42, target.Field1)
}

func TestIsPointerAllocateNil(t *testing.T) {
	var target *testStruct

	val, err := isPointer(&target, reflect.Struct, true)
	assert.NoError(t, err)

	val.Elem().FieldByIndex([]int{0}).SetInt(42)

	assert.Equal(t, 42, target.Field1)
}
