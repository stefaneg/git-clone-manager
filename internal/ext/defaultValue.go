package ext

import "reflect"

func DefaultValue[T any](value T, fallback T) T {
	if isZero(value) {
		return fallback
	}
	return value
}

func isZero[T any](value T) bool {
	return reflect.DeepEqual(value, reflect.Zero(reflect.TypeOf(value)).Interface())
}
