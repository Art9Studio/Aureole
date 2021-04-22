package collections

import (
	"aureole/internal/configs"
	"fmt"
)

type (
	CollectionType struct {
		Name           string
		IsAppendix     bool
		ParentCollType string
	}

	// Collection is a shorthand for CollectionImplementation
	Collection struct {
		Name        string
		Type        string // CollectionType name
		Parent      string // CollectionType name which applies to
		UseExistent bool
		Spec        Spec
	}

	Spec struct {
		Name      string
		Pk        string
		FieldsMap map[string]string
	}
)

func Create(conf *configs.Collection) (*Collection, error) {
	coll := &Collection{
		Name:        conf.Name,
		Type:        conf.Type,
		Parent:      conf.Parent,
		UseExistent: conf.UseExistent,
		Spec: Spec{
			Name:      conf.Spec.Name,
			Pk:        conf.Spec.Pk,
			FieldsMap: conf.Spec.FieldsMap,
		},
	}

	return coll, nil
}

func (c Collection) Init(collections map[string]*Collection) error {
	collType, err := Repository.Get(c.Type)
	if err != nil {
		return err
	}

	if collType.IsAppendix && c.Parent != "" {
		//todo: print warning "You passed Parent prop to main col. It skipped"
	}

	if c.Parent != "" {
		parentColl := collections[c.Parent]
		if collType.ParentCollType != parentColl.Type {
			return fmt.Errorf("declared CollectionType is not same with used one in config for collection '%s'", c.Name)
		}
	} else {
		if collType.ParentCollType != "" {
			return fmt.Errorf("valid parent is required for collection '%s'", c.Name)
		}
	}

	return nil
}
