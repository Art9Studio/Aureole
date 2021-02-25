package collections

import (
	"aureole/configs"
)

// todo: reorganize this structures
type (
	Collection struct {
		Type        string
		UseExistent bool
		Spec        Specification
	}

	Specification struct {
		Name      string
		Pk        string
		FieldsMap map[string]string
	}
)

func NewCollection(collType string, conf *configs.Collection) *Collection {
	return &Collection{
		Type:        collType,
		UseExistent: conf.UseExistent,
		Spec: Specification{
			Name:      conf.Spec.Name,
			Pk:        conf.Spec.Pk,
			FieldsMap: conf.Spec.FieldsMap,
		},
	}
}
