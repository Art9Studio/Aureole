package plugins

import (
	"aureole/internal/configs"
	"fmt"

	"github.com/lestrrat-go/jwx/jwk"
)

var CryptoKeyRepo = createRepository()

const (
	Private KeyType = "private"
	Public  KeyType = "public"
)

type (
	// CryptoKeyAdapter defines methods for authentication plugins
	CryptoKeyAdapter interface {
		// Create returns desired crypto key depends on the given config
		Create(*configs.CryptoKey) CryptoKey
	}

	CryptoKey interface {
		MetaDataGetter
		GetPrivateSet() jwk.Set
		GetPublicSet() jwk.Set
	}

	KeyType string
)

// NewCryptoKey returns desired CryptoKey depends on the given config
func NewCryptoKey(conf *configs.CryptoKey) (CryptoKey, error) {
	a, err := CryptoKeyRepo.Get(conf.Type)
	if err != nil {
		return nil, err
	}

	adapter, ok := a.(CryptoKeyAdapter)
	if !ok {
		return nil, fmt.Errorf("trying to cast adapter was failed: %v", err)
	}

	return adapter.Create(conf), nil
}
