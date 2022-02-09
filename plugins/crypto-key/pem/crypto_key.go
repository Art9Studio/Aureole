package pem

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
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
	"github.com/avast/retry-go/v4"
	"github.com/go-openapi/spec"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/crypto/ed25519"
)

const pluginID = "6374"

type cryptoKey struct {
	pluginAPI       core.PluginAPI
	rawConf         *configs.CryptoKey
	conf            *config
	cryptoStorage   plugins.CryptoStorage
	refreshDone     chan struct{}
	refreshInterval time.Duration
	muPrivSet       sync.RWMutex
	privateSet      jwk.Set
	muPubSet        sync.RWMutex
	publicSet       jwk.Set
	swagger         struct {
		Paths       *spec.Paths
		Definitions spec.Definitions
	}
}

//go:embed docs/swagger.json
var swaggerJson []byte

func (ck *cryptoKey) Init(api core.PluginAPI) (err error) {
	ck.pluginAPI = api
	if ck.conf, err = initConfig(&ck.rawConf.Config); err != nil {
		return err
	}
	ck.conf.PathPrefix = "/" + strings.ReplaceAll(ck.rawConf.Name, "_", "-")

	ck.cryptoStorage, err = ck.pluginAPI.GetCryptoStorage(ck.conf.Storage)
	if err != nil {
		return err
	}

	err = initKeySets(ck)
	if err != nil {
		return err
	}
	createRoutes(ck)

	if ck.conf.RefreshInterval != 0 {
		ck.refreshInterval = time.Duration(ck.conf.RefreshInterval) * time.Millisecond
		ck.refreshDone = make(chan struct{})
		//go refreshKeys(p)
	}

	err = json.Unmarshal(swaggerJson, &ck.swagger)
	if err != nil {
		fmt.Printf("pe, crypto-key: cannot marshal swagger docs: %v", err)
	}

	jwkHandler := ck.swagger.Paths.Paths["/jwk"]
	pemHandler := ck.swagger.Paths.Paths["/pem"]
	ck.swagger.Paths.Paths = map[string]spec.PathItem{
		ck.conf.PathPrefix + "/jwk": jwkHandler,
		ck.conf.PathPrefix + "/pem": pemHandler,
	}

	return nil
}

func (ck *cryptoKey) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: ck.rawConf.Name,
		ID:   pluginID,
	}
}

func (ck *cryptoKey) GetHandlersSpec() (*spec.Paths, spec.Definitions) {
	return ck.swagger.Paths, ck.swagger.Definitions
}

func (ck *cryptoKey) GetPrivateSet() jwk.Set {
	ck.muPrivSet.RLock()
	privSet := ck.privateSet
	ck.muPrivSet.RUnlock()
	return privSet
}

func (ck *cryptoKey) GetPublicSet() jwk.Set {
	ck.muPubSet.RLock()
	pubSet := ck.publicSet
	ck.muPubSet.RUnlock()
	return pubSet
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}

	return adapterConf, nil
}

func initKeySets(p *cryptoKey) (err error) {
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

	if setType == plugins.Private {
		p.privateSet = keySet
		if p.publicSet, err = jwk.PublicSetOf(p.privateSet); err != nil {
			return err
		}
	} else {
		p.publicSet = keySet
	}

	return nil
}

func createRoutes(p *cryptoKey) {
	routes := []*core.Route{
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
	p.pluginAPI.AddProjectRoutes(routes)
}

func refreshKeys(p *cryptoKey) {
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

			if setType == plugins.Private {
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

func getKeySetType(keySet jwk.Set) (plugins.KeyType, error) {
	isPrivate, err := isPrivateSet(keySet)
	if err != nil {
		return "", err
	}
	if isPrivate {
		return plugins.Private, nil
	}

	isPublic, err := isPublicSet(keySet)
	if err != nil {
		return "", err
	}
	if isPublic {
		return plugins.Public, nil
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
