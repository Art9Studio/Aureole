package types

import (
	"github.com/lestrrat-go/jwx/jwk"
)

type CryptoKey interface {
	Init() error
	GetPrivateSet() jwk.Set
	GetPublicSet() jwk.Set
}
