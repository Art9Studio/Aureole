package pbkdf2

import (
	"aureole/configs"
	"aureole/plugins/pwhasher/types"
	"github.com/mitchellh/mapstructure"
)

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

// TODO: figure out best default settings
func (c *Conf) setDefaults() {
	if c.Iterations == 0 {
		c.Iterations = 4096
	}

	if c.SaltLen == 0 {
		c.SaltLen = 16
	}

	if c.KeyLen == 0 {
		c.KeyLen = 32
	}

	if c.FuncName == "" {
		c.FuncName = "sha1"
	}
}

// Create returns Pbkdf2 hasher with the given settings
func (a pbkdf2Adapter) Create(conf *configs.PwHasher) (types.PwHasher, error) {
	adapterConfMap := conf.Config
	adapterConf := &Conf{}

	err := mapstructure.Decode(adapterConfMap, adapterConf)
	if err != nil {
		return nil, err
	}

	adapterConf.setDefaults()

	return initAdapter(conf, adapterConf)
}

func initAdapter(conf *configs.PwHasher, adapterConf *Conf) (*Pbkdf2, error) {
	function, err := initFunc(adapterConf.FuncName)
	if err != nil {
		return nil, err
	}

	return &Pbkdf2{Conf: adapterConf, Func: function}, nil
}
