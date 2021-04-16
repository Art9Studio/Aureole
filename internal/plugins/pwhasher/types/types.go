package types

// PwHasher is an interface that defined method for pwhasher implementation
type PwHasher interface {
	Init() error

	// HashPw returns hashed data encoded by base64
	HashPw(string) (string, error)

	// ComparePw compares plain data and hashed data encoded by base64
	ComparePw(string, string) (bool, error)
}
