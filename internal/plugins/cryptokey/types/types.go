package types

import (
	"aureole/internal/plugins"
	"github.com/lestrrat-go/jwx/jwk"
)

type KeyType string

const (
	Private KeyType = "private"
	Public  KeyType = "public"
)

type CryptoKey interface {
	plugins.MetaDataGetter
	GetPrivateSet() jwk.Set
	GetPublicSet() jwk.Set
}
