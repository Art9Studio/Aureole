package pbkdf2

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"gouth/pwhash"
	"hash"
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

	return Pbkdf2{conf: config}, nil
}

// newConfig creates new HashConfig struct from the raw data, parsed from the config file
func newConfig(rawConf *pwhash.RawHashConfig) (*HashConfig, error) {
	requiredKeys := []string{"mode", "iterations", "parallelism", "salt_length", "key_length", "memory"}

	for _, key := range requiredKeys {
		if _, ok := (*rawConf)[key]; !ok {
			return &HashConfig{}, fmt.Errorf("pwhash config: missing %s statement", key)
		}
	}

	// TODO: add rawConf validation

	var f func() hash.Hash

	switch (*rawConf)["func"].(string) {
	case "sha-1":
		f = sha1.New
	case "sha-224":
		f = sha256.New224
	case "sha-256":
		f = sha256.New
	case "sha-384":
		f = sha512.New384
	case "sha-512":
		f = sha512.New
	}

	return &HashConfig{
		Iterations: (*rawConf)["iterations"].(int),
		SaltLen:    (*rawConf)["salt_length"].(int),
		KeyLen:     (*rawConf)["key_length"].(int),
		Func:       f,
	}, nil
}
