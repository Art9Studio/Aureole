package pbkdf2

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"gouth/adapters/pwhasher"
	"hash"
)

// TODO: figure out best default settings
// DefaultConfig provides some sane default settings for hashing passwords
var DefaultConfig = &HashConfig{
	Iterations: 4096,
	SaltLen:    16,
	KeyLen:     32,
	FuncName:   "sha1",
	Func:       sha1.New,
}

// HashConfig represents parsed pwhasher config from the config file
type HashConfig struct {
	// The number of iterations over the memory
	Iterations int

	// Length of the random salt. 16 bytes is recommended for password hashing
	SaltLen int

	// Length of the generated key. 16 bytes or more is recommended
	KeyLen int

	// Name of the pseudorandom function
	FuncName string

	// Pseudorandom function used to derive a secure encryption key based on the password
	Func func() hash.Hash
}

//GetHasher returns Argon2 hasher with the given settings
func (a pbkdf2Adapter) GetPwHasher(rawConf *pwhasher.RawHashConfig) (pwhasher.PwHasher, error) {
	config, err := newConfig(rawConf)
	if err != nil {
		return nil, err
	}

	return &Pbkdf2{conf: config}, nil
}

// newConfig creates new HashConfig struct from the raw data, parsed from the config file
func newConfig(rawConf *pwhasher.RawHashConfig) (*HashConfig, error) {
	requiredKeys := []string{"iterations", "salt_length", "key_length", "func"}

	for _, key := range requiredKeys {
		if _, ok := (*rawConf)[key]; !ok {
			return &HashConfig{}, fmt.Errorf("pwhasher config: missing %s statement", key)
		}
	}

	// TODO: add rawConf validation

	conf := &HashConfig{
		Iterations: (*rawConf)["iterations"].(int),
		SaltLen:    (*rawConf)["salt_length"].(int),
		KeyLen:     (*rawConf)["key_length"].(int),
	}

	funcName := (*rawConf)["func"].(string)

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
		return nil, fmt.Errorf("pbkdf2: function '%s' don't supported", funcName)
	}

	conf.FuncName = funcName

	return conf, nil
}
