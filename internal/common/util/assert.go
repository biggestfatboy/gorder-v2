package util

import (
	"fmt"
	"reflect"
)

func AssertNotEmpty(fields ...any) error {

	for _, field := range fields {
		if isEmpty(field) {
			return fmt.Errorf("assert: not empty failed")
		}
	}
	return nil
}

func isEmpty(object interface{}) bool {
	// get nil case out of the way
	if object == nil {
		return true
	}

	return isEmptyValue(reflect.ValueOf(object))
}

func isEmptyValue(objValue reflect.Value) bool {
	if objValue.IsZero() {
		return true
	}
	// Special cases of non-zero values that we consider empty
	switch objValue.Kind() {
	// collection types are empty when they have no element
	// Note: array types are empty when they match their zero-initialized state.
	case reflect.Chan, reflect.Map, reflect.Slice:
		return objValue.Len() == 0
	// non-nil pointers are empty if the value they point to is empty
	case reflect.Ptr:
		return isEmptyValue(objValue.Elem())
	}
	return false
}
