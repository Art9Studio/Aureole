package pbkdf2

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/pbkdf2"
	"strings"
)

// Pbkdf2 represents pbkdf2 hasher
type Pbkdf2 struct {
	conf *HashConfig
}

var ErrInvalidHash = errors.New("pbkdf2: the encoded pwhash is not in the correct format")

func (p Pbkdf2) HashPw(pw string) (string, error) {
	salt := make([]byte, p.conf.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	key := pbkdf2.Key([]byte(pw), salt, p.conf.Iterations, p.conf.KeyLen, p.conf.Func)

	hash := fmt.Sprintf("$%s$p=%d$%s$%s",
		p.conf.FuncName,
		p.conf.Iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)

	return hash, nil
}

func (p Pbkdf2) ComparePw(pw string, hash string) (bool, error) {
	conf, salt, key, err := decodePwHash(hash)
	if err != nil {
		return false, err
	}

	otherKey := pbkdf2.Key([]byte(pw), salt, conf.Iterations, conf.KeyLen, conf.Func)

	if subtle.ConstantTimeCompare(key, otherKey) == 1 {
		return true, nil
	}

	return false, nil
}

// decodePwHash expects a pwhash created from this package, and parses it to return the config
// used to create it, as well as the salt and key
func decodePwHash(hash string) (*HashConfig, []byte, []byte, error) {
	vals := strings.Split(hash, "$")
	if len(vals) != 5 {
		return nil, nil, nil, ErrInvalidHash
	}

	conf := &HashConfig{}
	var funcName string

	_, err := fmt.Sscanf(vals[1], "%s", &funcName)
	if err != nil {
		return nil, nil, nil, err
	}

	_, err = fmt.Sscanf(vals[2], "p=%d", &conf.Iterations)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(vals[3])
	if err != nil {
		return nil, nil, nil, err
	}
	conf.SaltLen = len(salt)

	key, err := base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	conf.KeyLen = len(key)

	return conf, salt, key, nil
}
