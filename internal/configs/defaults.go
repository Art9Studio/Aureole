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
	for i := range p.Apps {
		a := p.Apps[i]
		a.setDefaults()
		p.Apps[i] = a
	}
}

func (a *App) setDefaults() {
	for i := range a.Authn {
		a.Authn[i].setDefaults()
	}

	for i := range a.Authz {
		a.Authz[i].setDefaults()
	}
}

func (i *Identity) setDefaults() {
	SetDefault(&i.Collection, "identity")

	SetDefault(&i.Id, trait{
		Enabled:  true,
		Unique:   true,
		Required: true,
		Internal: true,
	})

	SetDefault(&i.Username, trait{
		Enabled:  true,
		Unique:   false,
		Required: false,
		Internal: false,
	})

	SetDefault(&i.Phone, trait{
		Enabled:  false,
		Unique:   true,
		Required: false,
		Internal: false,
	})

	SetDefault(&i.Email, trait{
		Enabled:  true,
		Unique:   true,
		Required: true,
		Internal: false,
	})
}

func (authn *Authn) setDefaults() {
	SetDefault(&authn.PathPrefix, "/")
}

func (authz *Authz) setDefaults() {
	SetDefault(&authz.PathPrefix, "/")
}
