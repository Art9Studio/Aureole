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
	SetDefault(&p.PingPath, "/ping")

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

	a.Identity.setDefaults()
}

func (i *Identity) setDefaults() {
	SetDefault(&i.Collection, "identity")

	keys := []string{"enabled", "unique", "required", "credential"}

	i.Id = setDefaultTrait(i.Id, keys, []bool{true, true, true, false})
	i.Username = setDefaultTrait(i.Username, keys, []bool{true, false, false, true})
	i.Email = setDefaultTrait(i.Email, keys, []bool{false, true, true, true})
	i.Phone = setDefaultTrait(i.Phone, keys, []bool{false, true, false, true})
}

func setDefaultTrait(trait map[string]bool, keys []string, vals []bool) map[string]bool {
	if trait == nil {
		trait = map[string]bool{}
		for i, key := range keys {
			trait[key] = vals[i]
		}
	} else {
		for i, key := range keys {
			if _, ok := trait[key]; !ok {
				trait[key] = vals[i]
			}
		}
	}
	return trait
}

func (authn *Authn) setDefaults() {
	SetDefault(&authn.PathPrefix, "/")
}

func (authz *Authz) setDefaults() {
	SetDefault(&authz.PathPrefix, "/")
}
