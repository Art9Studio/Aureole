package core

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"math/big"
)

func encrypt(app *app, data interface{}) ([]byte, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	key, err := getKey(app)
	if err != nil {
		return nil, err
	}

	return rsa.EncryptOAEP(sha256.New(), rand.Reader, &key.PublicKey, dataBytes, nil)
}

func decrypt(app *app, data []byte, value interface{}) error {
	key, err := getKey(app)
	if err != nil {
		return err
	}

	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, data, nil)
	if err != nil {
		return err
	}

	return json.Unmarshal(decrypted, value)
}

func getKey(app *app) (*rsa.PrivateKey, error) {
	serviceKey, ok := app.getServiceEncKey()
	if !ok {
		return nil, errors.New("cannot get internal encryption key")
	}
	set := serviceKey.GetPrivateSet()

	key, ok := set.Get(0)
	if !ok {
		return nil, errors.New("cannot get internal key")
	}

	var rsaKey rsa.PrivateKey
	if err := key.Raw(&rsaKey); err != nil {
		return nil, err
	}

	return &rsaKey, nil
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
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		randBytes[i] = alphabet[num.Int64()]
	}

	return string(randBytes), nil
}
