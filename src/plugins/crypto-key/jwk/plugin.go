package jwk

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/avast/retry-go/v4"
	jwx "github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/x25519"
	"github.com/mitchellh/mapstructure"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"golang.org/x/crypto/ed25519"
)

//go:embed meta.yaml
var rawMeta []byte

var meta core.Metadata

// init initializes package by register pluginCreator
func init() {
	meta = core.CryptoKeyRepo.Register(rawMeta, Create)
}

type (
	jwk struct {
		pluginAPI       core.PluginAPI
		rawConf         configs.PluginConfig
		conf            *config
		cryptoStorage   core.CryptoStorage
		refreshInterval time.Duration
		muPrivSet       sync.RWMutex
		privateSet      jwx.Set
		muPubSet        sync.RWMutex
		publicSet       jwx.Set
		refreshDone     chan struct{}
	}
)

func Create(conf configs.PluginConfig) core.CryptoKey {
	return &jwk{rawConf: conf}
}

func (j *jwk) Init(api core.PluginAPI) (err error) {
	j.pluginAPI = api
	if j.conf, err = initConfig(&j.rawConf.Config); err != nil {
		return err
	}
	j.conf.PathPrefix = "/" + strings.ReplaceAll(j.rawConf.Name, "_", "-")

	var ok bool
	j.cryptoStorage, ok = j.pluginAPI.GetCryptoStorage(j.conf.Storage)
	if !ok {
		return fmt.Errorf("crypto storage named '%s' is not declared", j.conf.Storage)
	}

	err = initKeySets(j)
	if err != nil {
		return err
	}

	if j.conf.RefreshInterval != 0 {
		j.refreshInterval = time.Duration(j.conf.RefreshInterval) * time.Millisecond
		j.refreshDone = make(chan struct{})
		go refreshKeys(j)
	}

	// TODO: IMPORTANT: uncomment it and fix
	//jwkHandler := j.swagger.Paths.Paths["/jwk"]
	//pemHandler := j.swagger.Paths.Paths["/pem"]
	//j.swagger.Paths.Paths = map[string]openapi3.PathItem{
	//	j.conf.PathPrefix + "/jwk": jwkHandler,
	//	j.conf.PathPrefix + "/pem": pemHandler,
	//}

	return nil
}

func (j *jwk) GetMetadata() core.Metadata {
	return meta
}

func initConfig(conf *configs.RawConfig) (*config, error) {
	PluginConf := &config{}
	if err := mapstructure.Decode(conf, PluginConf); err != nil {
		return nil, err
	}
	PluginConf.setDefaults()

	return PluginConf, nil
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

	if setType == core.Private {
		j.privateSet = keySet
		if j.publicSet, err = jwx.PublicSetOf(j.privateSet); err != nil {
			return err
		}
	} else {
		j.publicSet = keySet
	}

	return nil
}

func (j *jwk) GetCustomAppRoutes() []*core.Route {
	return []*core.Route{
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

			if setType == core.Private {
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

func getKeySetType(keySet jwx.Set) (core.KeyType, error) {
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
