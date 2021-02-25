package argon2

import (
	"fmt"
	"gouth/pwhash"
)

// AdapterName is the internal name of the adapter
const AdapterName = "argon2"

// init initializes package by register adapter
func init() {
	pwhash.RegisterAdapter(AdapterName, argon2Adapter{})
}

// argon2Adapter represents adapter for argon2 pwhash algorithm
type argon2Adapter struct {
}

//GetHasher returns Argon2 hasher with the given settings
func (a argon2Adapter) GetPwHasher(rawConf *pwhash.RawHashConfig) (pwhash.PwHasher, error) {
	config, err := newConfig(rawConf)
	if err != nil {
		return nil, err
	}

	return &Argon2{conf: config}, nil
}

// newConfig creates new HashConfig struct from the raw data, parsed from the config file
func newConfig(rawConf *pwhash.RawHashConfig) (*HashConfig, error) {
	requiredKeys := []string{"type", "iterations", "parallelism", "salt_length", "key_length", "memory"}

	for _, key := range requiredKeys {
		if _, ok := (*rawConf)[key]; !ok {
			return &HashConfig{}, fmt.Errorf("pwhash config: missing %s statement", key)
		}
	}

	// TODO: add rawConf validation

	return &HashConfig{
		Type:        (*rawConf)["type"].(string),
		Iterations:  uint32((*rawConf)["iterations"].(int)),
		Parallelism: uint8((*rawConf)["parallelism"].(int)),
		SaltLen:     uint32((*rawConf)["salt_length"].(int)),
		KeyLen:      uint32((*rawConf)["key_length"].(int)),
		Memory:      uint32((*rawConf)["memory"].(int)),
	}, nil
}
