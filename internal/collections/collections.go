package collections

import (
	"aureole/internal/configs"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

type (
	CollectionType struct {
		Name           string
		IsAppendix     bool
		ParentCollType string
	}

	// Collection is a shorthand for CollectionImplementation
	Collection struct {
		Name       string
		Type       string // CollectionType name
		ParentName string // CollectionType name which applies to
		Parent     *Collection
		Spec       Spec
	}

	Spec struct {
		Name      string
		Pk        string
		FieldsMap map[string]FieldSpec
	}

	FieldSpec struct {
		Name    string
		Default interface{}
	}
)

func NewFieldsMap(rawFieldsMap map[string]interface{}) (map[string]FieldSpec, error) {
	fieldsMap := make(map[string]FieldSpec, len(rawFieldsMap))

	for fieldName, rawSpec := range rawFieldsMap {
		fieldSpec := FieldSpec{}

		switch rawSpec := rawSpec.(type) {
		case string:
			fieldSpec.Name = rawSpec
		case map[interface{}]interface{}:
			err := mapstructure.Decode(rawSpec, &fieldSpec)
			if err != nil {
				return nil, err
			}
		}
		fieldsMap[fieldName] = fieldSpec
	}

	return fieldsMap, nil
}

func Create(conf *configs.Collection) (*Collection, error) {
	fieldsMap, err := NewFieldsMap(conf.Spec.FieldsMap)
	if err != nil {
		return nil, err
	}

	coll := &Collection{
		Name:       conf.Name,
		Type:       conf.Type,
		ParentName: conf.Parent,
		Spec: Spec{
			Name:      conf.Spec.Name,
			Pk:        conf.Spec.Pk,
			FieldsMap: fieldsMap,
		},
	}

	return coll, nil
}

func (c *Collection) Init(collections map[string]*Collection) error {
	collType, err := Repository.Get(c.Type)
	if err != nil {
		return err
	}

	if collType.IsAppendix && c.ParentName != "" {
		//todo: print warning "You passed ParentName prop to main col. It skipped"
	}

	if c.ParentName != "" {
		parentColl := collections[c.ParentName]
		if collType.ParentCollType != parentColl.Type {
			return fmt.Errorf("declared CollectionType is not same with used one in config for collection '%s'", c.Name)
		}
		c.Parent = parentColl
	} else if collType.ParentCollType != "" {
		return fmt.Errorf("valid parent is required for collection '%s'", c.Name)
	}

	return nil
}
