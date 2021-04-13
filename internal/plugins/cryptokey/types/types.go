package types

import (
	"aureole/internal"
	"github.com/lestrrat-go/jwx/jwk"
)

type CryptoKey interface {
	internal.Initializer
	Get(string) (jwk.Set, error)
}
