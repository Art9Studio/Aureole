package types

import (
	"github.com/lestrrat-go/jwx/jwk"
)

type KeyType string

const (
	Private KeyType = "private"
	Public  KeyType = "public"
)

type CryptoKey interface {
	Init() error
	GetPrivateSet() jwk.Set
	GetPublicSet() jwk.Set
}
