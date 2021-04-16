package pbkdf2

import (
	"aureole/configs"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/crypto/pbkdf2"
	"hash"
	"strings"
)

// Pbkdf2 represents pbkdf2 hasher
type Pbkdf2 struct {
	rawConf *configs.PwHasher
	conf    *config
	// Pseudorandom function used to derive a secure encryption key based on the password
	function func() hash.Hash
}

var ErrInvalidHash = errors.New("pbkdf2: the encoded pwhasher is not in the correct format")

func (p *Pbkdf2) Init() error {
	adapterConf := &config{}
	if err := mapstructure.Decode(p.rawConf.Config, adapterConf); err != nil {
		return err
	}
	adapterConf.setDefaults()
	p.conf = adapterConf

	function, err := initFunc(p.conf.FuncName)
	if err != nil {
		return err
	}
	p.function = function

	return nil
}

// HashPw returns a Pbkdf2 pwhasher of a plain-text password using the provided
// algorithm parameters. The returned pwhasher follows the format used by the
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

	key := pbkdf2.Key([]byte(pw), salt, p.conf.Iterations, p.conf.KeyLen, p.function)

	hashed := fmt.Sprintf("pbkdf2_%s$%d$%s$%s",
		p.conf.FuncName,
		p.conf.Iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)

	return hashed, nil
}

// ComparePw performs a constant-time comparison between a plain-text password and
// Pbkdf2 pwhasher, using the parameters and salt contained in the pwhasher.
// It returns true if they match, otherwise it returns false.
func (p *Pbkdf2) ComparePw(pw, hash string) (bool, error) {
	conf, function, salt, key, err := decodePwHash(hash)
	if err != nil {
		return false, err
	}

	otherKey := pbkdf2.Key([]byte(pw), salt, conf.Iterations, conf.KeyLen, function)

	if subtle.ConstantTimeCompare(key, otherKey) == 1 {
		return true, nil
	}

	return false, nil
}

// decodePwHash expects a pwhasher created from this package, and parses it to return the config
// used to create it, as well as the salt and key
func decodePwHash(hashed string) (*config, func() hash.Hash, []byte, []byte, error) {
	vals := strings.Split(hashed, "$")
	if len(vals) != 4 {
		return nil, nil, nil, nil, ErrInvalidHash
	}

	conf := &config{}
	var funcName string

	_, err := fmt.Sscanf(vals[0], "%s", &funcName)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	funcName = strings.TrimLeft(funcName, "pbkdf2_")
	var function (func() hash.Hash)
	switch funcName {
	case "sha1":
		function = sha1.New
	case "sha224":
		function = sha256.New224
	case "sha256":
		function = sha256.New
	case "sha384":
		function = sha512.New384
	case "sha512":
		function = sha512.New
	default:
		return nil, nil, nil, nil, fmt.Errorf("pbkdf2: function '%s' don't supported", funcName)
	}

	conf.FuncName = funcName

	_, err = fmt.Sscanf(vals[1], "%d", &conf.Iterations)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(vals[2])
	if err != nil {
		return nil, nil, nil, nil, err
	}
	conf.SaltLen = len(salt)

	key, err := base64.RawStdEncoding.DecodeString(vals[3])
	if err != nil {
		return nil, nil, nil, nil, err
	}
	conf.KeyLen = len(key)

	return conf, function, salt, key, nil
}

func initFunc(funcName string) (func() hash.Hash, error) {
	switch funcName {
	case "sha1":
		return sha1.New, nil
	case "sha224":
		return sha256.New224, nil
	case "sha256":
		return sha256.New, nil
	case "sha384":
		return sha512.New384, nil
	case "sha512":
		return sha512.New, nil
	default:
		return nil, fmt.Errorf("pbkdf2: function '%s' don't supported", funcName)
	}
}
