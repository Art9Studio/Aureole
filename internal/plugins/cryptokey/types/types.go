package types

import (
	"github.com/lestrrat-go/jwx/jwk"
)

type CryptoKey interface {
	Init() error
	Get(string) (jwk.Set, error)
}
