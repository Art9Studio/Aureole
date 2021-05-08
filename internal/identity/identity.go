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
	}

	Trait struct {
		Enabled  bool
		Unique   bool
		Required bool
		Internal bool
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

	return &Identity{
		Collection: coll,
		Id:         Trait(conf.Id),
		Username:   Trait(conf.Username),
		Phone:      Trait(conf.Phone),
		Email:      Trait(conf.Email),
	}, nil
}
