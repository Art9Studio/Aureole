package types

import "github.com/lestrrat-go/jwx/jwk"

type CryptoKey interface {
	Get(string) (jwk.Set, error)
}
