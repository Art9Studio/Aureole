package core

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"math/big"

	eciesgo "github.com/ecies/go"
)

func encrypt(app *app, data interface{}) ([]byte, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	key, err := getKey(app)
	if err != nil {
		return nil, err
	}

	encrypted, err := eciesgo.Encrypt(key.PublicKey, bytes)
	if err != nil {
		return nil, err
	}

	return encrypted, nil
}

func decrypt(app *app, data []byte, value interface{}) error {
	key, err := getKey(app)
	if err != nil {
		return err
	}

	decrypted, err := eciesgo.Decrypt(key, data)
	if err != nil {
		return err
	}

	return json.Unmarshal(decrypted, value)
}

func getKey(app *app) (*eciesgo.PrivateKey, error) {
	serviceKey, ok := app.getServiceEncKey()
	if !ok {
		return nil, errors.New("cannot get service encryption key")
	}
	set := serviceKey.GetPrivateSet()

	key, ok := set.Get(0)
	if !ok {
		return nil, errors.New("cannot get service key")
	}

	var ecKey ecdsa.PrivateKey
	if err := key.Raw(&ecKey); err != nil {
		return nil, err
	}

	return &eciesgo.PrivateKey{
		PublicKey: &eciesgo.PublicKey{
			Curve: ecKey.Curve,
			X:     ecKey.X,
			Y:     ecKey.Y,
		},
		D: ecKey.D,
	}, nil
}

func getRandStr(length int, alphabet string) (string, error) {
	switch alphabet {
	case "alpha":
		alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	case "num":
		alphabet = "0123456789"
	case "alphanum":
		alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	}

	randBytes := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := crand.Int(crand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		randBytes[i] = alphabet[num.Int64()]
	}

	return string(randBytes), nil
}
