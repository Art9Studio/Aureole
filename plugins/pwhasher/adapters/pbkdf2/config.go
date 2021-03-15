package pbkdf2

import (
	"aureole/configs"
	"aureole/plugins/pwhasher/types"
	"github.com/mitchellh/mapstructure"
)

// TODO: figure out best default settings
// todo: use default
// DefaultConfig provides some sane default settings for hashing passwords
//var DefaultConfig = &Conf{
//	Iterations: 4096,
//	SaltLen:    16,
//	KeyLen:     32,
//	FuncName:   "sha1",
//	Func:       sha1.New,
//}

// Conf represents parsed pwhasher config from the config file
type Conf struct {
	// The number of iterations over the memory
	Iterations int `mapstructure:"iterations"`

	// Length of the random salt. 16 bytes is recommended for password hashing
	SaltLen int `mapstructure:"salt_length"`

	// Length of the generated key. 16 bytes or more is recommended
	KeyLen int `mapstructure:"key_length"`

	// Name of the pseudorandom function
	FuncName string `mapstructure:"func"`
}

// GetPwHasher returns Pbkdf2 hasher with the given settings
func (a pbkdf2Adapter) Get(conf *configs.Hasher) (types.PwHasher, error) {
	adapterConfMap := conf.Config
	adapterConf := &Conf{}
	err := mapstructure.Decode(adapterConfMap, adapterConf)
	if err != nil {
		return nil, err
	}

	return initAdapter(conf, adapterConf)
}

func initAdapter(conf *configs.Hasher, adapterConf *Conf) (*Pbkdf2, error) {
	function, err := initFunc(adapterConf.FuncName)
	if err != nil {
		return nil, err
	}

	return &Pbkdf2{Conf: adapterConf, Func: function}, nil
}
