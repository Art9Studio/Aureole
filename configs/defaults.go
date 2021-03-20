package configs

import (
	"reflect"
)

type Defaultable interface {
	setDefaults()
}

func SetDefault(target interface{}, def interface{}) {
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
	for i := range p.Apps {
		a := p.Apps[i]
		a.setDefaults()
		p.Apps[i] = a
	}
}

func (a *app) setDefaults() {
	for i := range a.Authn {
		a.Authn[i].setDefaults()
	}
}

func (authn *Authn) setDefaults() {
	SetDefault(&authn.PathPrefix, "/")
}
