package argon2

type HashConfig struct {
	Mode     string
	Iter     uint32
	Parallel uint8
	SaltLen  uint32
	KeyLen   uint32
	Memory   uint32
}
