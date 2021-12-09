package jwk

import (
	"aureole/internal/plugins"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/decred/dcrd/dcrec/secp256k1/v3"
	jwx "github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/x25519"
	"golang.org/x/crypto/ed25519"
	"hash"
	"math/big"
)

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
