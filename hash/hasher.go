package hash

// Hasher is an interface that defined method for hash implementation
type Hasher interface {
	// Hash returns hashed data encoded by base64
	Hash(string) (string, error)

	// Compare compares plain data and hashed data encoded by base64
	Compare(string, string) (bool, error)
}
