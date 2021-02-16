package argon2

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"strings"
)

// Argon2 represents argon2 hasher
type Argon2 struct {
	conf *HashConfig
}

var (
	ErrInvalidHash         = errors.New("argon2: the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("argon2: incompatible version of argon2")
)

// Hash returns a Argon2 hash of a plain-text password using the provided algorithm
// parameters. The returned hash follows the format used by the Argon2 reference
// C implementation and contains the base64-encoded Argon2 derived key prefixed
// by the salt and parameters. It looks like this:
//
//		$argon2i$v=19$m=65536,t=3,p=2$c29tZXNhbHQ$RdescudvJCsgt3ub+b+dWRWJTmaaJObG
//
func (a Argon2) Hash(data string) (string, error) {
	salt := make([]byte, a.conf.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	var key []byte

	switch a.conf.Type {
	case "argon2i":
		key = argon2.Key([]byte(data), salt, a.conf.Iterations, a.conf.Memory, a.conf.Parallelism, a.conf.KeyLen)
	case "argon2id":
		key = argon2.IDKey([]byte(data), salt, a.conf.Iterations, a.conf.Memory, a.conf.Parallelism, a.conf.KeyLen)
	}

	hash := fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		a.conf.Type,
		argon2.Version,
		a.conf.Memory,
		a.conf.Iterations,
		a.conf.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)

	return hash, nil
}

// Compare performs a constant-time comparison between a plain-text data and
// Argon2 hash, using the parameters and salt contained in the hash.
// It returns true if they match, otherwise it returns false.
func (a Argon2) Compare(data string, hash string) (bool, error) {
	conf, salt, key, err := decodeHash(hash)
	if err != nil {
		return false, err
	}

	var otherKey []byte

	switch conf.Type {
	case "argon2i":
		otherKey = argon2.Key([]byte(data), salt, conf.Iterations, conf.Memory, conf.Parallelism, conf.KeyLen)
	case "argon2id":
		otherKey = argon2.IDKey([]byte(data), salt, conf.Iterations, conf.Memory, conf.Parallelism, conf.KeyLen)
	}

	keyLen := int32(len(key))
	otherKeyLen := int32(len(otherKey))

	if subtle.ConstantTimeEq(keyLen, otherKeyLen) == 0 {
		return false, nil
	}

	if subtle.ConstantTimeCompare(key, otherKey) == 1 {
		return true, nil
	}

	return false, nil
}

// decodeHash expects a hash created from this package, and parses it to return the config
// used to create it, as well as the salt and key
func decodeHash(hash string) (*HashConfig, []byte, []byte, error) {
	vals := strings.Split(hash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	var v int
	_, err := fmt.Sscanf(vals[2], "v=%d", &v)
	if err != nil {
		return nil, nil, nil, err
	}

	if v != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	conf := &HashConfig{}
	conf.Type = vals[1]
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &conf.Memory, &conf.Iterations, &conf.Parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	conf.SaltLen = uint32(len(salt))

	key, err := base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	conf.KeyLen = uint32(len(key))

	return conf, salt, key, nil
}
