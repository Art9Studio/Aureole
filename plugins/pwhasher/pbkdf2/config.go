package pbkdf2

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/pwhasher/types"
)

// config represents parsed pwhasher config from the config file
type config struct {
	// The number of iterations over the memory
	Iterations int `mapstructure:"iterations"`

	// Length of the random salt. 16 bytes is recommended for password hashing
	SaltLen int `mapstructure:"salt_length"`

	// Length of the generated key. 16 bytes or more is recommended
	KeyLen int `mapstructure:"key_length"`

	// Name of the pseudorandom function
	FuncName string `mapstructure:"func"`
}

// Create returns Pbkdf2 hasher with the given settings
func (pbkdf2Adapter) Create(conf *configs.PwHasher) types.PwHasher {
	return &Pbkdf2{rawConf: conf}
}
