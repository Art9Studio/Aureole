package plugins

import (
	"aureole/internal/configs"
	"fmt"
)

var PWHasherRepo = createRepository()

// PWHasherAdapter defines methods for pwhasher plugins
type (
	PWHasherAdapter interface {
		// Create returns desired pwHasher depends on the given config
		Create(*configs.PwHasher) PWHasher
	}

	PWHasher interface {
		MetaDataGetter
		HashPw(pw string) (hashPw string, err error)
		ComparePw(pw string, hashPw string) (match bool, err error)
	}
)

// NewPWHasher returns desired pwHasher depends on the given config
func NewPWHasher(conf *configs.PwHasher) (PWHasher, error) {
	a, err := PWHasherRepo.Get(conf.Type)
	if err != nil {
		return nil, err
	}

	adapter, ok := a.(PWHasherAdapter)
	if !ok {
		return nil, fmt.Errorf("trying to cast adapter was failed: %v", err)
	}

	return adapter.Create(conf), nil
}
