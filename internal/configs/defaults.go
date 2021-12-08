package configs

import (
	"reflect"
)

type Defaultable interface {
	setDefaults()
}

func SetDefault(target, def interface{}) {
	val := reflect.ValueOf(target)
	if isZero(val.Elem()) {
		val.Elem().Set(reflect.ValueOf(def))
	}
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && isZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		z := true
		for i := 0; i < v.NumField(); i++ {
			z = z && isZero(v.Field(i))
		}
		return z
	}
	// Compare other types directly:
	z := reflect.Zero(v.Type())
	return v.Interface() == z.Interface()
}

// todo: run all setDefaults recursively with reflect
func (p *Project) setDefaults() {
	p.PingPath = "/ping"
}
