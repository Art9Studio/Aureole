package pwhasher

import (
	"aureole/internal/configs"
	"aureole/plugins/authn/pwbased/pwhasher/argon2"
	"aureole/plugins/authn/pwbased/pwhasher/pbkdf2"
	"fmt"
)

type (
	Config struct {
		Type   string            `config:"type"`
		Config configs.RawConfig `config:"config"`
	}

	PWHasher interface {
		Init(conf configs.RawConfig) error
		HashPw(pw string) (hashPw string, err error)
		ComparePw(pw string, hashPw string) (match bool, err error)
	}
)

func NewPWHasher(conf Config) (hasher PWHasher, err error) {
	switch conf.Type {
	case "pbkdf2":
		hasher = &pbkdf2.PWHasher{}
	case "argon2":
		hasher = &argon2.PWHasher{}
	default:
		return nil, fmt.Errorf("pw hasher '%s' is not supported", conf.Type)
	}
	err = hasher.Init(conf.Config)
	return hasher, err
}
