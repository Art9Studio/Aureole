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
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-openapi/spec"

	"github.com/lestrrat-go/jwx/x25519"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/crypto/ed25519"
)

const pluginID = "7851"

type cryptoKey struct {
	pluginAPI       core.PluginAPI
	rawConf         *configs.CryptoKey
	conf            *config
	cryptoStorage   plugins.CryptoStorage
	refreshInterval time.Duration
	muPrivSet       sync.RWMutex
	privateSet      jwk.Set
	muPubSet        sync.RWMutex
	publicSet       jwk.Set
	refreshDone     chan struct{}
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
		//go refreshKeys(j)
	}

	err = json.Unmarshal(swaggerJson, &ck.swagger)
	if err != nil {
		fmt.Printf("jwk crypto-key: cannot marshal swagger docs: %v", err)
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

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()

	return adapterConf, nil
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

func initKeySets(j *cryptoKey) (err error) {
	var (
		rawKeys []byte
		keySet  jwk.Set
	)

	found, err := j.cryptoStorage.Read(&rawKeys)
	if err != nil {
		return err
	}

	if found {
		keySet, err = jwk.Parse(rawKeys)
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
		if j.publicSet, err = jwk.PublicSetOf(j.privateSet); err != nil {
			return err
		}
	} else {
		j.publicSet = keySet
	}

	return nil
}

func createRoutes(j *cryptoKey) {
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

var s = `{"a": 10}`

func refreshKeys(j *cryptoKey) {
	ticker := time.NewTicker(j.refreshInterval)
	/*var (
		rawKeys []byte
		keySet  jwk.Set
		setType plugins.KeyType
		pubSet  jwk.Set
		ok      bool
		err     error
	)*/
	var a interface{}

	//_ = keySet

	defer ticker.Stop()

	for {
		select {
		case <-j.refreshDone:
			return
		case <-ticker.C:
			/*err := retry.Do(
				func() error {
					ok, err := j.cryptoStorage.Read(&rawKeys)
					if err != nil {
						return err
					}
					if !ok {
						return errors.New("keys don't find")
					}

					keySet, err = jwk.Parse(rawKeys)
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
			*/

			fmt.Println("refresh key")

			/*ok, err := j.cryptoStorage.Read(&rawKeys)
			if err != nil {
				fmt.Printf("jwk '%s': an error occured while refreshing keys: %v\n", j.rawConf.Name, err)
				continue
			}
			if !ok {
				fmt.Printf("jwk '%s': an error occured while refreshing keys: keys don't find\n", j.rawConf.Name)
				continue
			}*/

			err := json.Unmarshal([]byte(s), &a)
			if err != nil {
				fmt.Printf("jwk '%s': an error occured while refreshing keys: %v\n", j.rawConf.Name, err)
				continue
			}

			/*keySet, err = jwk.Parse(rawKeys)
			if err != nil {
				fmt.Printf("jwk '%s': an error occured while refreshing keys: %v\n", j.rawConf.Name, err)
				continue
			}*/

			/*
				setType, err = getKeySetType(keySet)
				if err != nil {
					fmt.Printf("jwk '%s': an error occured while refreshing keys: %v\n", j.rawConf.Name, err)
					continue
				}

				if setType == plugins.Private {
					pubSet, err = jwk.PublicSetOf(keySet)
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
				}*/
		}
	}
}

func generateKey(conf *config) (keySet jwk.Set, err error) {
	pubRawKey, privRawKey, err := generateRawKey(conf)
	if err != nil {
		return nil, err
	}

	key, err := jwk.New(privRawKey)
	if err != nil {
		return nil, err
	}
	kid, err := generateKid(pubRawKey, conf.Kid)
	if err != nil {
		return nil, err
	}

	if err := key.Set(jwk.KeyIDKey, kid); err != nil {
		return nil, err
	}
	if err := key.Set(jwk.AlgorithmKey, conf.Alg); err != nil {
		return nil, err
	}
	if err := key.Set(jwk.KeyUsageKey, conf.Use); err != nil {
		return nil, err
	}

	keySet = jwk.NewSet()
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
