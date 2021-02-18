package pbkdf2

import (
	"crypto/sha1"
	"hash"
)

// TODO: figure out best default settings
// DefaultConfig provides some sane default settings for hashing passwords
var DefaultConfig = &HashConfig{
	Iterations: 4096,
	SaltLen:    16,
	KeyLen:     32,
	FuncName:   "sha1",
	Func:       sha1.New,
}

// HashConfig represents parsed pwhash config from the config file
type HashConfig struct {
	// The number of iterations over the memory
	Iterations int

	// Length of the random salt. 16 bytes is recommended for password hashing
	SaltLen int

	// Length of the generated key. 16 bytes or more is recommended
	KeyLen int

	// Name of the pseudorandom function
	FuncName string

	// Pseudorandom function used to derive a secure encryption key based on the password
	Func func() hash.Hash
}
