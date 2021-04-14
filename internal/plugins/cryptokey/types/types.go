package types

import (
	"github.com/lestrrat-go/jwx/jwk"
)

type CryptoKey interface {
	Initialize() error
	Get(string) (jwk.Set, error)
}
