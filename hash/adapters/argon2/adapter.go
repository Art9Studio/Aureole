package argon2

import (
	"fmt"
	"gouth/hash"
)

// AdapterName is the internal name of the adapter
const AdapterName = "argon2"

// init initializes package by register adapter
func init() {
	hash.RegisterAdapter(AdapterName, argon2Adapter{})
}

// argon2Adapter represents adapter for argon2 hash algorithm
type argon2Adapter struct {
}

//GetHasher returns Argon2 hasher with the given settings
func (a argon2Adapter) GetHasher(rawConf *hash.RawHashConfig) (hash.Hasher, error) {
	config, err := newConfig(rawConf)
	if err != nil {
		return nil, err
	}

	return Argon2{conf: config}, nil
}

// newConfig creates new HashConfig struct from the raw data, parsed from the config file
func newConfig(rawConf *hash.RawHashConfig) (*HashConfig, error) {
	requiredKeys := []string{"mode", "iterations", "parallelism", "salt_length", "key_length", "memory"}

	for _, key := range requiredKeys {
		if _, ok := (*rawConf)[key]; !ok {
			return &HashConfig{}, fmt.Errorf("hash config: missing %s statement", key)
		}
	}

	// TODO: add rawConf validation

	return &HashConfig{
		Type:        (*rawConf)["mode"].(string),
		Iterations:  uint32((*rawConf)["iterations"].(int)),
		Parallelism: uint8((*rawConf)["parallelism"].(int)),
		SaltLen:     uint32((*rawConf)["salt_length"].(int)),
		KeyLen:      uint32((*rawConf)["key_length"].(int)),
		Memory:      uint32((*rawConf)["memory"].(int)),
	}, nil
}
