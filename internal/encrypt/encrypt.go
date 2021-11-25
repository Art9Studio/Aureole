package encrypt

import (
	state "aureole/internal/state/interface"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	eciesgo "github.com/ecies/go"
)

var project state.ProjectState

func Init(p state.ProjectState) {
	project = p
}

func Encrypt(data interface{}) ([]byte, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	key, err := getKey()
	if err != nil {
		return nil, err
	}

	encrypted, err := eciesgo.Encrypt(key.PublicKey, bytes)
	if err != nil {
		return nil, err
	}

	return encrypted, nil
}

func Decrypt(data []byte, value interface{}) error {
	key, err := getKey()
	if err != nil {
		return err
	}

	decrypted, err := eciesgo.Decrypt(key, data)
	if err != nil {
		return err
	}

	return json.Unmarshal(decrypted, value)
}

func getKey() (*eciesgo.PrivateKey, error) {
	serviceKey, err := project.GetServiceEncKey()
	if err != nil {
		return nil, err
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
