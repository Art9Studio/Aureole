package argon2

import (
	"aureole/configs"
	"aureole/plugins/pwhasher/types"
	"github.com/mitchellh/mapstructure"
)

// Conf represents parsed pwhasher config from the config file
type Conf struct {
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

// Create returns Argon2 hasher with the given settings
func (a argon2Adapter) Create(conf *configs.PwHasher) (types.PwHasher, error) {
	adapterConfMap := conf.Config
	adapterConf := &Conf{}

	err := mapstructure.Decode(adapterConfMap, adapterConf)
	if err != nil {
		return nil, err
	}

	adapterConf.setDefaults()

	return initAdapter(conf, adapterConf)
}

func initAdapter(conf *configs.PwHasher, adapterConf *Conf) (*Argon2, error) {
	return &Argon2{Conf: adapterConf}, nil
}
