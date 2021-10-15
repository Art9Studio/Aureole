package pem

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/cryptokey"
	"aureole/internal/plugins/cryptokey/types"
	_interface "aureole/internal/router/interface"
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/crypto/ed25519"
	"strings"
)

type Pem struct {
	rawConf    *configs.CryptoKey
	conf       *config
	privateSet jwk.Set
	publicSet  jwk.Set
}

func (p *Pem) Init() (err error) {
	p.rawConf.PathPrefix = "/" + strings.Replace(p.rawConf.Name, "_", "-", -1)
	if p.conf, err = initConfig(&p.rawConf.Config); err != nil {
		return err
	}
	err = initKeySets(p)
	if err != nil {
		return err
	}
	createRoutes(p)

	return nil
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}

	return adapterConf, nil
}

func initKeySets(p *Pem) error {
	keySet, err := jwk.ReadFile(p.conf.Path, jwk.WithPEM(true))
	if err != nil {
		return err
	}

	setType, err := getKeySetType(keySet)
	if err != nil {
		return err
	}

	if err := setAttr(keySet, p.conf.Alg); err != nil {
		return err
	}

	if setType == types.Private {
		p.privateSet = keySet
		if p.publicSet, err = jwk.PublicSetOf(p.privateSet); err != nil {
			return err
		}
	} else {
		p.publicSet = keySet
	}

	return nil
}

func getKeySetType(keySet jwk.Set) (types.KeyType, error) {
	isPrivate, err := isPrivateSet(keySet)
	if err != nil {
		return "", err
	}
	if isPrivate {
		return types.Private, nil
	}

	isPublic, err := isPublicSet(keySet)
	if err != nil {
		return "", err
	}
	if isPublic {
		return types.Public, nil
	}

	return "", errors.New("public and private keys in the same key set")
}

func isPrivateSet(keySet jwk.Set) (bool, error) {
	for it := keySet.Iterate(context.Background()); it.Next(context.Background()); {
		pair := it.Pair()
		key := pair.Value.(jwk.Key)

		var rawKey interface{}
		if err := key.Raw(&rawKey); err != nil {
			return false, err
		}

		if _, ok := rawKey.(*rsa.PublicKey); ok {
			return false, nil
		}
		if _, ok := rawKey.(*ed25519.PublicKey); ok {
			return false, nil
		}
		if _, ok := rawKey.(*ecdsa.PublicKey); ok {
			return false, nil
		}
	}
	return true, nil
}

func isPublicSet(keySet jwk.Set) (bool, error) {
	for it := keySet.Iterate(context.Background()); it.Next(context.Background()); {
		pair := it.Pair()
		key := pair.Value.(jwk.Key)

		var rawKey interface{}
		if err := key.Raw(&rawKey); err != nil {
			return false, err
		}

		if _, ok := rawKey.(*rsa.PrivateKey); ok {
			return false, nil
		}
		if _, ok := rawKey.(*ed25519.PrivateKey); ok {
			return false, nil
		}
		if _, ok := rawKey.(*ecdsa.PrivateKey); ok {
			return false, nil
		}
	}
	return true, nil
}

func setAttr(keySet jwk.Set, alg string) error {
	for it := keySet.Iterate(context.Background()); it.Next(context.Background()); {
		pair := it.Pair()
		key := pair.Value.(jwk.Key)
		if err := key.Set(jwk.AlgorithmKey, alg); err != nil {
			return err
		}
		if err := key.Set(jwk.KeyUsageKey, "sig"); err != nil {
			return err
		}
	}
	return nil
}

func createRoutes(p *Pem) {
	routes := []*_interface.Route{
		{
			Method:  "GET",
			Path:    p.rawConf.PathPrefix + "/jwk",
			Handler: GetJwkKeys(p),
		},
		{
			Method:  "GET",
			Path:    p.rawConf.PathPrefix + "/pem",
			Handler: GetPemKeys(p),
		},
	}
	cryptokey.Repository.PluginApi.Router.AddProjectRoutes(routes)
}

func (p *Pem) GetPrivateSet() jwk.Set {
	return p.privateSet
}

func (p *Pem) GetPublicSet() jwk.Set {
	return p.publicSet
}
