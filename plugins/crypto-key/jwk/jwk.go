package jwk

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
	"github.com/go-openapi/spec"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/lestrrat-go/jwx/x25519"

	jwx "github.com/lestrrat-go/jwx/jwk"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/crypto/ed25519"
)

const pluginID = "7851"

type (
	jwk struct {
		pluginAPI       core.PluginAPI
		rawConf         *configs.CryptoKey
		conf            *config
		cryptoStorage   plugins.CryptoStorage
		refreshInterval time.Duration
		muPrivSet       sync.RWMutex
		privateSet      jwx.Set
		muPubSet        sync.RWMutex
		publicSet       jwx.Set
		refreshDone     chan struct{}
		swagger         struct {
			Paths       *spec.Paths
			Definitions spec.Definitions
		}
	}
)

//go:embed swagger.json
var swaggerJson []byte

func (j *jwk) Init(api core.PluginAPI) (err error) {
	j.pluginAPI = api
	if j.conf, err = initConfig(&j.rawConf.Config); err != nil {
		return err
	}
	j.conf.PathPrefix = "/" + strings.ReplaceAll(j.rawConf.Name, "_", "-")

	var ok bool
	j.cryptoStorage, ok = j.pluginAPI.GetCryptoStorage(j.conf.Storage)
	if !ok {
		return fmt.Errorf("crytpo storage named '%s' is not declared", j.conf.Storage)
	}

	err = initKeySets(j)
	if err != nil {
		return err
	}
	createRoutes(j)

	if j.conf.RefreshInterval != 0 {
		j.refreshInterval = time.Duration(j.conf.RefreshInterval) * time.Millisecond
		j.refreshDone = make(chan struct{})
		go refreshKeys(j)
	}

	err = json.Unmarshal(swaggerJson, &j.swagger)
	if err != nil {
		fmt.Printf("jwk crypto-key: cannot marshal swagger docs: %v", err)
	}

	jwkHandler := j.swagger.Paths.Paths["/jwk"]
	pemHandler := j.swagger.Paths.Paths["/pem"]
	j.swagger.Paths.Paths = map[string]spec.PathItem{
		j.conf.PathPrefix + "/jwk": jwkHandler,
		j.conf.PathPrefix + "/pem": pemHandler,
	}

	return nil
}

func (j *jwk) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: j.rawConf.Name,
		ID:   pluginID,
	}
}

func (j *jwk) GetHandlersSpec() (*spec.Paths, spec.Definitions) {
	return j.swagger.Paths, j.swagger.Definitions
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()

	return adapterConf, nil
}

func (j *jwk) GetPrivateSet() jwx.Set {
	j.muPrivSet.RLock()
	privSet := j.privateSet
	j.muPrivSet.RUnlock()
	return privSet
}

func (j *jwk) GetPublicSet() jwx.Set {
	j.muPubSet.RLock()
	pubSet := j.publicSet
	j.muPubSet.RUnlock()
	return pubSet
}

func initKeySets(j *jwk) (err error) {
	var (
		rawKeys []byte
		keySet  jwx.Set
	)

	found, err := j.cryptoStorage.Read(&rawKeys)
	if err != nil {
		return err
	}

	if found {
		keySet, err = jwx.Parse(rawKeys)
		if err != nil {
			return err
		}
	} else {
		keySet, err = generateKey(j.conf)
		if err != nil {
			return err
		}

		b, err := json.Marshal(keySet)
		if err != nil {
			return err
		}
		if err := j.cryptoStorage.Write(b); err != nil {
			return err
		}
	}

	setType, err := getKeySetType(keySet)
	if err != nil {
		return err
	}

	if setType == plugins.Private {
		j.privateSet = keySet
		if j.publicSet, err = jwx.PublicSetOf(j.privateSet); err != nil {
			return err
		}
	} else {
		j.publicSet = keySet
	}

	return nil
}

func createRoutes(j *jwk) {
	routes := []*core.Route{
		{
			Method:  http.MethodGet,
			Path:    j.conf.PathPrefix + "/jwk",
			Handler: getJwkKeys(j),
		},
		{
			Method:  http.MethodGet,
			Path:    j.conf.PathPrefix + "/pem",
			Handler: getPemKeys(j),
		},
	}
	j.pluginAPI.AddProjectRoutes(routes)
}

func refreshKeys(j *jwk) {
	ticker := time.NewTicker(j.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-j.refreshDone:
			return
		case <-ticker.C:
			var (
				rawKeys []byte
				keySet  jwx.Set
			)

			err := retry.Do(
				func() error {
					ok, err := j.cryptoStorage.Read(&rawKeys)
					if err != nil {
						return err
					}
					if !ok {
						return errors.New("keys don't find")
					}

					keySet, err = jwx.Parse(rawKeys)
					return err
				},
				retry.DelayType(retry.FixedDelay),
				retry.Delay(time.Duration(j.conf.RetryInterval)*time.Millisecond),
				retry.Attempts(uint(j.conf.RetriesNum)),
			)
			if err != nil {
				fmt.Printf("jwk '%s': an error occured while refreshing keys: %v\n", j.rawConf.Name, err)
				continue
			}

			setType, err := getKeySetType(keySet)
			if err != nil {
				fmt.Printf("jwk '%s': an error occured while refreshing keys: %v\n", j.rawConf.Name, err)
				continue
			}

			if setType == plugins.Private {
				pubSet, err := jwx.PublicSetOf(keySet)
				if err != nil {
					fmt.Printf("jwk '%s': an error occured while refreshing keys: %v\n", j.rawConf.Name, err)
					continue
				}

				j.muPrivSet.Lock()
				j.privateSet = keySet
				j.muPrivSet.Unlock()

				j.muPubSet.Lock()
				j.publicSet = pubSet
				j.muPubSet.Unlock()
			} else {
				j.muPubSet.Lock()
				j.publicSet = keySet
				j.muPubSet.Unlock()
			}
		}
	}
}

func generateKey(conf *config) (keySet jwx.Set, err error) {
	pubRawKey, privRawKey, err := generateRawKey(conf)
	if err != nil {
		return nil, err
	}

	key, err := jwx.New(privRawKey)
	if err != nil {
		return nil, err
	}
	kid, err := generateKid(pubRawKey, conf.Kid)
	if err != nil {
		return nil, err
	}

	if err := key.Set(jwx.KeyIDKey, kid); err != nil {
		return nil, err
	}
	if err := key.Set(jwx.AlgorithmKey, conf.Alg); err != nil {
		return nil, err
	}
	if err := key.Set(jwx.KeyUsageKey, conf.Use); err != nil {
		return nil, err
	}

	keySet = jwx.NewSet()
	keySet.Add(key)
	return keySet, nil
}

func generateRawKey(conf *config) (pubRawKey, privRawKey interface{}, err error) {
	// todo: delete defaults after adding validation
	switch conf.Kty {
	case "RSA":
		key, err := rsa.GenerateKey(rand.Reader, conf.Size)
		if err != nil {
			return nil, nil, err
		}
		pubRawKey = &key.PublicKey
		privRawKey = key
	case "oct":
		privRawKey, err = generateRandomBytes(conf.Size / 8)
		if err != nil {
			return nil, nil, err
		}
		pubRawKey = privRawKey
	case "EC":
		var curve elliptic.Curve
		switch conf.Curve {
		case "P-256":
			curve = elliptic.P256()
		case "P-384":
			curve = elliptic.P384()
		case "P-521":
			curve = elliptic.P521()
		default:
			return nil, nil, fmt.Errorf("wrong curve '%s' for kty '%s'", conf.Curve, conf.Kty)
		}

		key, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			return nil, nil, err
		}
		pubRawKey = &key.PublicKey
		privRawKey = key
	case "OKP":
		switch conf.Curve {
		case "ed25519":
			pubRawKey, privRawKey, err = ed25519.GenerateKey(rand.Reader)
			if err != nil {
				return nil, nil, err
			}
		case "x25519":
			pubRawKey, privRawKey, err = x25519.GenerateKey(rand.Reader)
			if err != nil {
				return nil, nil, err
			}
		default:
			return nil, nil, fmt.Errorf("wrong curve '%s' for kty '%s'", conf.Curve, conf.Kty)
		}
	default:
		return nil, nil, fmt.Errorf("kty '%s' is not supported", conf.Kty)
	}

	return pubRawKey, privRawKey, nil
}

func generateKid(rawKey interface{}, kidType string) (kid string, err error) {
	if kidType == "SHA-256" {
		var keyBytes []byte
		if b, ok := rawKey.([]byte); ok {
			keyBytes = b
		} else {
			keyBytes, err = x509.MarshalPKIXPublicKey(rawKey)
			if err != nil {
				return "", err
			}
		}

		h := sha256.Sum256(keyBytes)
		return base64.RawStdEncoding.EncodeToString(h[:]), nil
	}
	return kidType, nil
}

func generateRandomBytes(length int) ([]byte, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return nil, err
		}
		ret[i] = letters[num.Int64()]
	}

	return ret, nil
}

func getKeySetType(keySet jwx.Set) (plugins.KeyType, error) {
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

func isPrivateSet(keySet jwx.Set) (bool, error) {
	for it := keySet.Iterate(context.Background()); it.Next(context.Background()); {
		pair := it.Pair()
		key := pair.Value.(jwx.Key)

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

func isPublicSet(keySet jwx.Set) (bool, error) {
	for it := keySet.Iterate(context.Background()); it.Next(context.Background()); {
		pair := it.Pair()
		key := pair.Value.(jwx.Key)

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
