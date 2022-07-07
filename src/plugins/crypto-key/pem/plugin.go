package pem

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/crypto/ed25519"
)

//go:embed meta.yaml
var rawMeta []byte

var meta core.Meta

func init() {
	meta = core.CryptoKeyRepo.Register(rawMeta, Create)
}

type pem struct {
	pluginAPI       core.PluginAPI
	rawConf         configs.PluginConfig
	conf            *config
	cryptoStorage   core.CryptoStorage
	refreshDone     chan struct{}
	refreshInterval time.Duration
	muPrivSet       sync.RWMutex
	privateSet      jwk.Set
	muPubSet        sync.RWMutex
	publicSet       jwk.Set
}

func Create(conf configs.PluginConfig) core.CryptoKey {
	return &pem{rawConf: conf}
}

func (p *pem) Init(api core.PluginAPI) (err error) {
	p.pluginAPI = api
	if p.conf, err = initConfig(&p.rawConf.Config); err != nil {
		return err
	}
	p.conf.PathPrefix = "/" + strings.ReplaceAll(p.rawConf.Name, "_", "-")

	var ok bool
	p.cryptoStorage, ok = p.pluginAPI.GetCryptoStorage(p.conf.Storage)
	if !ok {
		return fmt.Errorf("crypto storage named '%s' is not declared", p.conf.Storage)
	}

	err = initKeySets(p)
	if err != nil {
		return err
	}

	if p.conf.RefreshInterval != 0 {
		p.refreshInterval = time.Duration(p.conf.RefreshInterval) * time.Millisecond
		p.refreshDone = make(chan struct{})
		go refreshKeys(p)
	}

	return nil
}

func (p *pem) GetMetaData() core.Meta {
	return meta
}

func (p *pem) GetPrivateSet() jwk.Set {
	p.muPrivSet.RLock()
	privSet := p.privateSet
	p.muPrivSet.RUnlock()
	return privSet
}

func (p *pem) GetPublicSet() jwk.Set {
	p.muPubSet.RLock()
	pubSet := p.publicSet
	p.muPubSet.RUnlock()
	return pubSet
}

func initConfig(conf *configs.RawConfig) (*config, error) {
	PluginConf := &config{}
	if err := mapstructure.Decode(conf, PluginConf); err != nil {
		return nil, err
	}

	return PluginConf, nil
}

func initKeySets(p *pem) (err error) {
	var (
		rawKeys []byte
		keySet  jwk.Set
	)

	ok, err := p.cryptoStorage.Read(&rawKeys)
	if err != nil {
		return err
	}

	if ok {
		keySet, err = jwk.Parse(rawKeys, jwk.WithPEM(true))
		if err != nil {
			return err
		}
		if err := setAttr(keySet, p.conf.Alg); err != nil {
			return err
		}
	} else {
		keySet, err = generateKey()
		if err != nil {
			return err
		}

		b, err := json.MarshalIndent(keySet, "", "  ")
		if err != nil {
			return err
		}
		if err := p.cryptoStorage.Write(b); err != nil {
			return err
		}
	}

	setType, err := getKeySetType(keySet)
	if err != nil {
		return err
	}

	if setType == core.Private {
		p.privateSet = keySet
		if p.publicSet, err = jwk.PublicSetOf(p.privateSet); err != nil {
			return err
		}
	} else {
		p.publicSet = keySet
	}

	return nil
}

func (p *pem) GetAppRoutes() []*core.Route {

	return []*core.Route{
		{
			Method:  http.MethodGet,
			Path:    p.conf.PathPrefix + "/jwk",
			Handler: getJwkKeys(p),
		},
		{
			Method:  http.MethodGet,
			Path:    p.conf.PathPrefix + "/pem",
			Handler: getPemKeys(p),
		},
	}
	// TODO:
	p.pluginAPI.AddProjectRoutes(routes)
}

func refreshKeys(p *pem) {
	ticker := time.NewTicker(p.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-p.refreshDone:
			return
		case <-ticker.C:
			var (
				rawKeys []byte
				keySet  jwk.Set
			)

			err := retry.Do(
				func() error {
					ok, err := p.cryptoStorage.Read(&rawKeys)
					if err != nil {
						return err
					}
					if !ok {
						fmt.Printf("pem '%s': an error occured while refreshing keys: %v", p.rawConf.Name, err)
					}

					keySet, err = jwk.Parse(rawKeys, jwk.WithPEM(true))
					return err
				},
				retry.DelayType(retry.FixedDelay),
				retry.Delay(time.Duration(p.conf.RetryInterval)*time.Millisecond),
				retry.Attempts(uint(p.conf.RetriesNum)),
			)
			if err != nil {
				fmt.Printf("pem '%s': an error occured while refreshing keys: %v", p.rawConf.Name, err)
				continue
			}

			if err := setAttr(keySet, p.conf.Alg); err != nil {
				fmt.Printf("pem '%s': cannot assign alg attribute to key while refreshing^ %v", p.rawConf.Name, err)
			}

			setType, err := getKeySetType(keySet)
			if err != nil {
				fmt.Printf("pem '%s': an error occured while refreshing keys: %v", p.rawConf.Name, err)
			}

			if setType == core.Private {
				pubSet, err := jwk.PublicSetOf(keySet)
				if err != nil {
					fmt.Printf("pem '%s': an error occured while refreshing keys: %v", p.rawConf.Name, err)
				}

				p.muPrivSet.Lock()
				p.privateSet = keySet
				p.muPrivSet.Unlock()

				p.muPubSet.Lock()
				p.publicSet = pubSet
				p.muPubSet.Unlock()
			} else {
				p.muPubSet.Lock()
				p.publicSet = keySet
				p.muPubSet.Unlock()
			}
		}
	}
}

func generateKey() (keySet jwk.Set, err error) {
	pubRawKey, privRawKey, err := generateRawKey()
	if err != nil {
		return nil, err
	}

	key, err := jwk.New(privRawKey)
	if err != nil {
		return nil, err
	}
	kid, err := generateKid(pubRawKey)
	if err != nil {
		return nil, err
	}

	if err := key.Set(jwk.KeyIDKey, kid); err != nil {
		return nil, err
	}
	if err := key.Set(jwk.AlgorithmKey, "ES256"); err != nil {
		return nil, err
	}
	if err := key.Set(jwk.KeyUsageKey, "sig"); err != nil {
		return nil, err
	}

	keySet = jwk.NewSet()
	keySet.Add(key)
	return keySet, nil
}

func generateRawKey() (interface{}, interface{}, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	return &key.PublicKey, key, nil
}

func generateKid(rawKey interface{}) (string, error) {
	keyBytes, err := x509.MarshalPKIXPublicKey(rawKey)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	_, err = h.Write(keyBytes)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

func getKeySetType(keySet jwk.Set) (core.KeyType, error) {
	isPrivate, err := isPrivateSet(keySet)
	if err != nil {
		return "", err
	}
	if isPrivate {
		return core.Private, nil
	}

	isPublic, err := isPublicSet(keySet)
	if err != nil {
		return "", err
	}
	if isPublic {
		return core.Public, nil
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
