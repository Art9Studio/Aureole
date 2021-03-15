package types

// PwHasher is an interface that defined method for pwhasher implementation
type PwHasher interface {
	// Hash returns hashed data encoded by base64
	HashPw(string) (string, error)

	// Compare compares plain data and hashed data encoded by base64
	ComparePw(string, string) (bool, error)
}
