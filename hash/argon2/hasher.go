package argon2

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"golang.org/x/crypto/argon2"
)

type Argon2 struct {
	conf HashConfig
}

func (a Argon2) Hash(in interface{}) ([]byte, error) {
	switch a.conf.Mode {
	case "argon2i":
		in, err := getBytes(in)
		if err != nil {
			return nil, err
		}

		salt, err := generateRandomBytes(a.conf.SaltLen)
		if err != nil {
			return nil, err
		}
		return argon2.Key(in, salt, a.conf.Iter, a.conf.Memory, a.conf.Parallel, a.conf.KeyLen), nil
	case "argon2id":
		in, err := getBytes(in)
		if err != nil {
			return nil, err
		}

		salt, err := generateRandomBytes(a.conf.SaltLen)
		if err != nil {
			return nil, err
		}
		return argon2.IDKey(in, salt, a.conf.Iter, a.conf.Memory, a.conf.Parallel, a.conf.KeyLen), nil
	default:
		return nil, fmt.Errorf("argon2 mode '%s' is not define", a.conf.Mode)
	}
}

func (a Argon2) Compare(in interface{}, bytes []byte) (bool, error) {
	panic("implement me")
}

func getBytes(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
