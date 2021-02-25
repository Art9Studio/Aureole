package pbkdf2

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"gouth/pwhash"
)

// AdapterName is the internal name of the adapter
const AdapterName = "pbkdf2"

// init initializes package by register adapter
func init() {
	pwhash.RegisterAdapter(AdapterName, pbkdf2Adapter{})
}

// pbkdf2Adapter represents adapter for pbkdf2 pwhash algorithm
type pbkdf2Adapter struct {
}

//GetHasher returns Argon2 hasher with the given settings
func (a pbkdf2Adapter) GetPwHasher(rawConf *pwhash.RawHashConfig) (pwhash.PwHasher, error) {
	config, err := newConfig(rawConf)
	if err != nil {
		return nil, err
	}

	return &Pbkdf2{conf: config}, nil
}

// newConfig creates new HashConfig struct from the raw data, parsed from the config file
func newConfig(rawConf *pwhash.RawHashConfig) (*HashConfig, error) {
	requiredKeys := []string{"iterations", "salt_length", "key_length", "func"}

	for _, key := range requiredKeys {
		if _, ok := (*rawConf)[key]; !ok {
			return &HashConfig{}, fmt.Errorf("pwhash config: missing %s statement", key)
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
