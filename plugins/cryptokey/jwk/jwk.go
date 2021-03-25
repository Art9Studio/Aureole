package jwk

import (
	"aureole/internal/configs"
	"context"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/mitchellh/mapstructure"
	"net/url"
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

	if j.privateSet, err = createPrivateSet(j.conf.Path); err != nil {
		return err
	}
	if j.publicSet, err = createPublicSet(j.privateSet); err != nil {
		return err
	}

	return nil
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}

	return adapterConf, nil
}

func createPrivateSet(path string) (privateSet jwk.Set, err error) {
	if _, err = url.ParseRequestURI(path); err != nil {
		privateSet, err = jwk.ReadFile(path)
		if err != nil {
			return nil, err
		}
	} else {
		privateSet, err = jwk.Fetch(context.Background(), path)
		if err != nil {
			return nil, err
		}
	}

	return privateSet, nil
}

func createPublicSet(privateSet jwk.Set) (publicSet jwk.Set, err error) {
	publicSet = jwk.NewSet()

	for it := privateSet.Iterate(context.Background()); it.Next(context.Background()); {
		pair := it.Pair()
		key := pair.Value.(jwk.Key)

		publicKey, err := jwk.PublicKeyOf(key)
		if err != nil {
			return nil, err
		}
		publicSet.Add(publicKey)
	}

	return publicSet, nil
}

func (j *Jwk) GetPrivateSet() jwk.Set {
	return j.privateSet
}

func (j *Jwk) GetPublicSet() jwk.Set {
	return j.publicSet
}
