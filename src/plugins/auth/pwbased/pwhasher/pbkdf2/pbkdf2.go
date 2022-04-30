package pbkdf2

import (
	"aureole/internal/configs"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"math/big"
	"strings"

	"github.com/mitchellh/mapstructure"
	pbkdf "golang.org/x/crypto/pbkdf2"
)

type (
	PWHasher struct {
		conf     *config
		function func() hash.Hash
	}

	// config represents parsed pwhasher config from the config file
	config struct {
		// The number of iterations over the memory
		Iterations int `mapstructure:"iterations"`
		// Length of the random salt. 16 bytes is recommended for password hashing
		SaltLen int `mapstructure:"salt_length"`
		// Length of the generated key. 16 bytes or more is recommended
		KeyLen int `mapstructure:"key_length"`
		// ProviderName of the pseudorandom function
		FuncName string `mapstructure:"func"`
	}
)

func (p *PWHasher) Init(rawConf configs.RawConfig) error {
	PluginConf := &config{}
	if err := mapstructure.Decode(rawConf, PluginConf); err != nil {
		return err
	}
	PluginConf.setDefaults()
	p.conf = PluginConf

	function, err := initFunc(p.conf.FuncName)
	if err != nil {
		return err
	}
	p.function = function

	return nil
}

// HashPw returns a PWHasher pwhasher of a plain-text password using the provided
// algorithm parameters. The returned pwhasher follows the format used by the
// Django and contains the base64-encoded PWHasher derived key prefixed by the
// salt and parameters. It looks like this:
//
//		pbkdf2_sha1$4096$c29tZXNhbHQ$RdescudvJCsgt3ub+b+dWRWJTmaaJObG
//
func (p *PWHasher) HashPw(pw string) (string, error) {
	salt, err := getRandStr(p.conf.SaltLen, "alphanum")
	if err != nil {
		return "", err
	}

	key := pbkdf.Key([]byte(pw), []byte(salt), p.conf.Iterations, p.conf.KeyLen, p.function)
	hashed := fmt.Sprintf("pbkdf2_%s$%d$%s$%s",
		p.conf.FuncName,
		p.conf.Iterations,
		salt,
		base64.StdEncoding.EncodeToString(key),
	)
	return hashed, nil
}

// ComparePw performs a constant-time comparison between a plain-text password and
// PWHasher pwhasher, using the parameters and salt contained in the pwhasher
// It returns true if they match, otherwise it returns false
func (*PWHasher) ComparePw(pw, hash string) (bool, error) {
	conf, function, salt, key, err := decodePwHash(hash)
	if err != nil {
		return false, err
	}

	otherKey := pbkdf.Key([]byte(pw), salt, conf.Iterations, conf.KeyLen, function)
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
		return nil, nil, nil, nil, errors.New("PWHasher: the encoded pwhasher is not in the correct format")
	}

	conf := &config{}
	var funcName string

	_, err := fmt.Sscanf(vals[0], "%s", &funcName)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	funcName = strings.TrimLeft(funcName, "pbkdf2_")

	var function func() hash.Hash
	switch funcName {
	case "sha224":
		function = sha256.New224
	case "sha256":
		function = sha256.New
	case "sha384":
		function = sha512.New384
	case "sha512":
		function = sha512.New
	default:
		return nil, nil, nil, nil, fmt.Errorf("PWHasher: function '%s' don't supported", funcName)
	}

	conf.FuncName = funcName

	_, err = fmt.Sscanf(vals[1], "%d", &conf.Iterations)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	salt := []byte(vals[2])
	conf.SaltLen = len(salt)

	key, err := base64.StdEncoding.DecodeString(vals[3])
	if err != nil {
		return nil, nil, nil, nil, err
	}
	conf.KeyLen = len(key)

	return conf, function, salt, key, nil
}

func initFunc(funcName string) (func() hash.Hash, error) {
	switch funcName {
	case "sha224":
		return sha256.New224, nil
	case "sha256":
		return sha256.New, nil
	case "sha384":
		return sha512.New384, nil
	case "sha512":
		return sha512.New, nil
	default:
		return nil, fmt.Errorf("PWHasher: function '%s' don't supported", funcName)
	}
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
