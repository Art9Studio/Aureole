package argon2

import (
	"fmt"
	"gouth/adapters/pwhasher"
)

// HashConfig represents parsed pwhasher configs from the configs file
type HashConfig struct {
	// AlgName type (argon2i, argon2id)
	Type string

	// The number of iterations over the memory
	Iterations uint32

	// The number of threads (or lanes) used by the algorithm.
	// Recommended value is between 1 and runtime.NumCPU()
	Parallelism uint8

	// Length of the random salt. 16 bytes is recommended for password hashing
	SaltLen uint32

	// Length of the generated key. 16 bytes or more is recommended
	KeyLen uint32

	// The amount of memory used by the algorithm (in kibibytes)
	Memory uint32
}

// TODO: figure out best default settings
// DefaultConfig provides some sane default settings for hashing passwords
var DefaultConfig = &HashConfig{
	Type:        "argon2i",
	Iterations:  3,
	Parallelism: 2,
	SaltLen:     16,
	KeyLen:      32,
	Memory:      32 * 1024,
}

//GetHasher returns Argon2 hasher with the given settings
func (a argon2Adapter) GetPwHasher(rawConf *pwhasher.RawHashConfig) (pwhasher.PwHasher, error) {
	config, err := newConfig(rawConf)
	if err != nil {
		return nil, err
	}

	return &Argon2{conf: config}, nil
}

// todo: completely rewrite this method
// newConfig creates new HashConfig struct from the raw data, parsed from the configs file
func newConfig(rawConf *pwhasher.RawHashConfig) (*HashConfig, error) {
	// todo: make validation with package for this and converting to structure
	requiredKeys := []string{"kind", "iterations", "parallelism", "salt_length", "key_length", "memory"}

	for _, key := range requiredKeys {
		if _, ok := (*rawConf)[key]; !ok {
			return &HashConfig{}, fmt.Errorf("pwhasher configs: missing %s statement", key)
		}
	}

	// TODO: add rawConf validation

	return &HashConfig{
		Type:        (*rawConf)["kind"].(string),
		Iterations:  uint32((*rawConf)["iterations"].(int)),
		Parallelism: uint8((*rawConf)["parallelism"].(int)),
		SaltLen:     uint32((*rawConf)["salt_length"].(int)),
		KeyLen:      uint32((*rawConf)["key_length"].(int)),
		Memory:      uint32((*rawConf)["memory"].(int)),
	}, nil
}
