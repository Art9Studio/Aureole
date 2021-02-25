package argon2

import (
	"aureole/configs"
	"aureole/plugins/pwhasher"
	"github.com/mitchellh/mapstructure"
)

// TODO: figure out best default settings
// DefaultConfig provides some sane default settings for hashing passwords
var DefaultConfig = &HashConfig{
	Kind:        "argon2i",
	Iterations:  3,
	Parallelism: 2,
	SaltLen:     16,
	KeyLen:      32,
	Memory:      32 * 1024,
}

// HashConfig represents parsed pwhasher configs from the configs file
type HashConfig struct {
	// AlgName kind (argon2i, argon2id)
	Kind string `mapstructure:"kind"`

	// The number of iterations over the memory
	Iterations uint32 `mapstructure:"iterations"`

	// The number of threads (or lanes) used by the algorithm.
	// Recommended value is between 1 and runtime.NumCPU()
	Parallelism uint8 `mapstructure:"parallelism"`

	// Length of the random salt. 16 bytes is recommended for password hashing
	SaltLen uint32 `mapstructure:"salt_length"`

	// Length of the generated key. 16 bytes or more is recommended
	KeyLen uint32 `mapstructure:"key_length"`

	// The amount of memory used by the algorithm (in kibibytes)
	Memory uint32 `mapstructure:""`
}

//GetHasher returns Argon2 hasher with the given settings
func (a argon2Adapter) GetPwHasher(confMap *configs.RawConfig) (pwhasher.PwHasher, error) {
	hashConfig := &HashConfig{}
	err := mapstructure.Decode(confMap, hashConfig)
	if err != nil {
		return nil, err
	}

	return &Argon2{conf: hashConfig}, nil
}
