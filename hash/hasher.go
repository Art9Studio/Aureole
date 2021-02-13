package hash

// Hasher is an interface that defined method for hash implementation
type Hasher interface {
	// TODO: think about parameter: interface{} or []byte?
	// Hash hashes given data
	Hash(interface{}) ([]byte, error)

	// Compare given plain data with the hash
	Compare(interface{}, []byte) (bool, error)
}
