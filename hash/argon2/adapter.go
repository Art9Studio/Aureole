package argon2

import (
	"fmt"
	"gouth/hash"
)

const AdapterName = "argon2"

func init() {
	hash.RegisterAdapter(AdapterName, argon2Adapter{})
}

type argon2Adapter struct {
}

func (a argon2Adapter) GetHasher(config hash.HashConfig) hash.Hasher {
	return Argon2{conf: config.(HashConfig)}
}

func (a argon2Adapter) NewConfig(data map[string]interface{}) (hash.HashConfig, error) {
	requiredKeys := []string{"mode", "iterations", "parallelism", "salt_length", "key_length", "memory"}

	for _, key := range requiredKeys {
		if _, ok := data[key]; !ok {
			return nil, fmt.Errorf("hash config: missing %s statement", key)
		}
	}

	// TODO: add data validation

	return HashConfig{
		Mode:     data["mode"].(string),
		Iter:     uint32(data["iterations"].(int)),
		Parallel: uint8(data["parallelism"].(int)),
		SaltLen:  uint32(data["salt_length"].(int)),
		KeyLen:   uint32(data["key_length"].(int)),
		Memory:   uint32(data["memory"].(int)),
	}, nil
}
