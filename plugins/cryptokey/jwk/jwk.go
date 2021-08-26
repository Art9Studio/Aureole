package jwk

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/cryptokey"
	"aureole/internal/plugins/cryptokey/types"
	_interface "aureole/internal/router/interface"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/decred/dcrd/dcrec/secp256k1/v3"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/x25519"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/crypto/ed25519"
	"hash"
	"io/ioutil"
	"math/big"
	"os"
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

func initKeySets(j *Jwk) (err error) {
	if j.conf.Kty != "" && j.conf.Alg != "" {
		var keySet jwk.Set
		if _, err := os.Stat(j.conf.Path); err == nil {
			keySet, err = jwk.ReadFile(j.conf.Path)
			if err != nil {
				return err
			}

			isMatch, err := isMatchConfig(keySet, j.conf)
			if err != nil {
				return err
			}

			if !isMatch {
				if err := os.Rename(j.conf.Path, j.conf.Path+".bkp"); err != nil {
					return err
				}
				keySet, err = generateKey(j.conf)
				if err != nil {
					return err
				}
			}
		} else {
			keySet, err = generateKey(j.conf)
			if err != nil {
				return err
			}
		}

		j.privateSet = keySet
		if j.publicSet, err = jwk.PublicSetOf(j.privateSet); err != nil {
			return err
		}
		return nil
	}

	keySet, err := getKeys(j.conf.Path)
	if err != nil {
		return err
	}

	setType, err := getKeySetType(keySet)
	if err != nil {
		return err
	}

	if setType == types.Private {
		j.privateSet = keySet
		if j.publicSet, err = jwk.PublicSetOf(j.privateSet); err != nil {
			return err
		}
	} else {
		j.publicSet = keySet
	}

	return nil
}

func isMatchConfig(keySet jwk.Set, conf *config) (bool, error) {
	key, ok := keySet.Get(0)
	if !ok {
		return false, errors.New("cannot get key from key set")
	}

	var kty jwa.KeyType
	if err := kty.Accept(conf.Kty); err != nil {
		return false, err
	}

	return keySet.Len() == 1 && key.Algorithm() == conf.Alg && key.KeyType() == kty, nil
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
	if err := key.Set(jwk.KeyUsageKey, "sig"); err != nil {
		return nil, err
	}

	keySet = jwk.NewSet()
	keySet.Add(key)
	jwkFile, err := json.MarshalIndent(keySet, "", "  ")
	if err != nil {
		return nil, err
	}
	if err = ioutil.WriteFile(conf.Path, jwkFile, 0644); err != nil {
		return nil, err
	}

	return keySet, nil
}

func generateRawKey(conf *config) (pubRawKey interface{}, privRawKey interface{}, err error) {
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
		case "secp256k1":
			curve = secp256k1.S256()
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
	var h hash.Hash
	switch kidType {
	case "SHA-256":
		h = sha256.New()
	case "SHA-1":
		h = sha1.New()
	default:
		return kidType, nil
	}

	var keyBytes []byte
	if b, ok := rawKey.([]byte); ok {
		keyBytes = b
	} else {
		keyBytes, err = x509.MarshalPKIXPublicKey(rawKey)
		if err != nil {
			return "", err
		}
	}

	_, err = h.Write(keyBytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
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
