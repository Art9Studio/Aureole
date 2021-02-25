package pbkdf2

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
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

// HashPw returns a Pbkdf2 pwhash of a plain-text password using the provided
// algorithm parameters. The returned pwhash follows the format used by the
// Django and contains the base64-encoded Pbkdf2 derived key prefixed by the
// salt and parameters. It looks like this:
//
//		pbkdf2_sha1$4096$c29tZXNhbHQ$RdescudvJCsgt3ub+b+dWRWJTmaaJObG
//
func (p *Pbkdf2) HashPw(pw string) (string, error) {
	salt := make([]byte, p.conf.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	key := pbkdf2.Key([]byte(pw), salt, p.conf.Iterations, p.conf.KeyLen, p.conf.Func)

	hash := fmt.Sprintf("pbkdf2_%s$%d$%s$%s",
		p.conf.FuncName,
		p.conf.Iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)

	return hash, nil
}

// ComparePw performs a constant-time comparison between a plain-text password and
// Pbkdf2 pwhash, using the parameters and salt contained in the pwhash.
// It returns true if they match, otherwise it returns false.
func (p *Pbkdf2) ComparePw(pw string, hash string) (bool, error) {
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
	if len(vals) != 4 {
		return nil, nil, nil, ErrInvalidHash
	}

	conf := &HashConfig{}
	var funcName string

	_, err := fmt.Sscanf(vals[0], "%s", &funcName)
	if err != nil {
		return nil, nil, nil, err
	}
	funcName = strings.TrimLeft(funcName, "pbkdf2_")

	switch funcName {
	case "sha1":
		conf.Func = sha1.New
	case "sha224":
		conf.Func = sha256.New224
	case "sha256":
		conf.Func = sha256.New
	case "sha384":
		conf.Func = sha512.New384
	case "sha512":
		conf.Func = sha512.New
	default:
		return nil, nil, nil, fmt.Errorf("pbkdf2: function '%s' don't supported", funcName)
	}

	conf.FuncName = funcName

	_, err = fmt.Sscanf(vals[1], "%d", &conf.Iterations)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(vals[2])
	if err != nil {
		return nil, nil, nil, err
	}
	conf.SaltLen = len(salt)

	key, err := base64.RawStdEncoding.DecodeString(vals[3])
	if err != nil {
		return nil, nil, nil, err
	}
	conf.KeyLen = len(key)

	return conf, salt, key, nil
}
