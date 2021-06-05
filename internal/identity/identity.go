package identity

import (
	"aureole/internal/collections"
	"aureole/internal/configs"
	"fmt"
)

type (
	Identity struct {
		Collection *collections.Collection
		Id         Trait
		Username   Trait
		Phone      Trait
		Email      Trait
		Additional map[string]ExtraTrait
	}

	Trait struct {
		IsEnabled    bool
		IsUnique     bool
		IsRequired   bool
		IsCredential bool
	}

	ExtraTrait struct {
		IsUnique   bool
		IsRequired bool
		IsInternal bool
	}
)

func init() {
	collections.Repository.Register(identColType)
}

func Create(conf *configs.Identity, collections map[string]*collections.Collection) (*Identity, error) {
	coll, ok := collections[conf.Collection]
	if !ok {
		return nil, fmt.Errorf("can't find collection named '%s'", conf.Collection)
	}

	additional := make(map[string]ExtraTrait, len(conf.Additional))
	for k, rawTrait := range conf.Additional {
		additional[k] = ExtraTrait(rawTrait)
	}

	return &Identity{
		Collection: coll,
		Id:         NewTrait(conf.Id),
		Username:   NewTrait(conf.Username),
		Phone:      NewTrait(conf.Phone),
		Email:      NewTrait(conf.Email),
		Additional: additional,
	}, nil
}

func NewTrait(rawTrait map[string]bool) Trait {
	return Trait{
		IsEnabled:    rawTrait["enabled"],
		IsUnique:     rawTrait["unique"],
		IsRequired:   rawTrait["required"],
		IsCredential: rawTrait["credential"],
	}
}
