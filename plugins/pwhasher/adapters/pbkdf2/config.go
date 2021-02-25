package pbkdf2

import (
	"aureole/configs"
	"aureole/plugins/pwhasher"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"hash"
)

// TODO: figure out best default settings
// DefaultConfig provides some sane default settings for hashing passwords
var DefaultConfig = &HashConfig{
	Iterations: 4096,
	SaltLen:    16,
	KeyLen:     32,
	FuncName:   "sha1",
	Function:   sha1.New,
}

// HashConfig represents parsed pwhasher configs from the configs file
type HashConfig struct {
	// The number of iterations over the memory
	Iterations int `mapstructure:"iterations"`

	// Length of the random salt. 16 bytes is recommended for password hashing
	SaltLen int `mapstructure:"salt_length"`

	// Length of the generated key. 16 bytes or more is recommended
	KeyLen int `mapstructure:"key_length"`

	// Name of the pseudorandom function
	FuncName string `mapstructure:"func"`

	// Pseudorandom function used to derive a secure encryption key based on the password
	Function func() hash.Hash
}

// GetPwHasher returns Pbkdf2 hasher with the given settings
func (a pbkdf2Adapter) GetPwHasher(confMap *configs.RawConfig) (pwhasher.PwHasher, error) {
	hashConfig := &HashConfig{}
	err := mapstructure.Decode(confMap, hashConfig)
	if err != nil {
		return nil, err
	}

	switch hashConfig.FuncName {
	case "sha1":
		hashConfig.Function = sha1.New
	case "sha224":
		hashConfig.Function = sha256.New224
	case "sha256":
		hashConfig.Function = sha256.New
	case "sha384":
		hashConfig.Function = sha512.New384
	case "sha512":
		hashConfig.Function = sha512.New
	default:
		return nil, fmt.Errorf("pbkdf2: function '%s' don't supported", hashConfig.FuncName)
	}

	return &Pbkdf2{conf: hashConfig}, nil
}
