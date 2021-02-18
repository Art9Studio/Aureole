package argon2

// HashConfig represents parsed pwhash config from the config file
type HashConfig struct {
	// Algorithm type (argon2i, argon2id)
	Type string

	// The number of iterations over the memory
	Iterations uint32

	// The number of threads (or lanes) used by the algorithm.
	// Recommended value is between 1 and runtime.NumCPU()
	Parallelism uint8

	// Length of the random salt. 16 bytes is recommended for password hashing
	SaltLen uint32

	// Length of the generated key. 16 bytes or more is recommended
	KeyLen uint32

	// The amount of memory used by the algorithm (in kibibytes)
	Memory uint32
}

// TODO: figure out best default settings
// DefaultConfig provides some sane default settings for hashing passwords
var DefaultConfig = &HashConfig{
	Type:        "argon2i",
	Iterations:  3,
	Parallelism: 2,
	SaltLen:     16,
	KeyLen:      32,
	Memory:      32 * 1024,
}
