package util

import (
	"reflect"
	"unsafe"
)

// MustGetUnexportedFieldByName extract value from given struct even if it is unexported.
// x must be pointer on struct not value. Panic on errors.
func MustGetUnexportedFieldByName(name string, x interface{}) interface{} {
	xReflection := reflect.ValueOf(x).Elem()
	fieldReflection := xReflection.FieldByName(name)
	return reflect.NewAt(fieldReflection.Type(), unsafe.Pointer(fieldReflection.UnsafeAddr())).Elem().Interface() // #nosec
}

// MustSetUnexportedFieldByName set value in given struct on defined value even if field is unexported.
// x must be pointer on struct not value. Panic on errors.
func MustSetUnexportedFieldByName(name string, obj interface{}, value interface{}) {
	objValue := reflect.ValueOf(obj).Elem()
	targetVal := objValue.FieldByName(name)
	targetVal = reflect.NewAt(targetVal.Type(), unsafe.Pointer(targetVal.UnsafeAddr())).Elem() // #nosec
	targetVal.Set(reflect.ValueOf(value))
}
