package jwk

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/cryptokey"
	_interface "aureole/internal/router/interface"
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/crypto/ed25519"
)

type Jwk struct {
	rawConf    *configs.CryptoKey
	conf       *config
	privateSet jwk.Set
	publicSet  jwk.Set
}

func (j *Jwk) Init() (err error) {
	if j.conf, err = initConfig(&j.rawConf.Config); err != nil {
		return err
	}
	err = initKeySets(j)
	if err != nil {
		return err
	}
	createRoutes(j)

	return nil
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}

	return adapterConf, nil
}

func initKeySets(j *Jwk) error {
	keySet, err := getKeys(j.conf.Path)
	if err != nil {
		return err
	}

	setType, err := getKeySetType(keySet)
	if err != nil {
		return err
	}

	if setType == "private" {
		j.privateSet = keySet
		if j.publicSet, err = jwk.PublicSetOf(j.privateSet); err != nil {
			return err
		}
	} else {
		j.publicSet = keySet
	}

	return nil
}

func getKeys(path string) (jwk.Set, error) {
	keySet, err := jwk.Fetch(context.Background(), path)
	if err != nil {
		keySet, err = jwk.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}

	return keySet, nil
}

func getKeySetType(keySet jwk.Set) (string, error) {
	isPrivate, err := isPrivateSet(keySet)
	if err != nil {
		return "", err
	}
	if isPrivate {
		return "private", nil
	}

	isPublic, err := isPublicSet(keySet)
	if err != nil {
		return "", err
	}
	if isPublic {
		return "public", nil
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

func createRoutes(j *Jwk) {
	routes := []*_interface.Route{
		{
			Method:  "GET",
			Path:    j.rawConf.PathPrefix + "/jwk",
			Handler: GetJwkKeys(j),
		},
		{
			Method:  "GET",
			Path:    j.rawConf.PathPrefix + "/pem",
			Handler: GetPemKeys(j),
		},
	}
	cryptokey.Repository.PluginApi.Router.AddProjectRoutes(routes)
}

func (j *Jwk) GetPrivateSet() jwk.Set {
	return j.privateSet
}

func (j *Jwk) GetPublicSet() jwk.Set {
	return j.publicSet
}
