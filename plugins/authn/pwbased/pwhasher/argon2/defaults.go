package argon2

import "aureole/internal/configs"

// TODO: figure out best default settings
func (c *config) setDefaults() {
	configs.SetDefault(&c.Kind, "argon2i")
	configs.SetDefault(&c.Iterations, uint32(3))
	configs.SetDefault(&c.Parallelism, uint8(2))
	configs.SetDefault(&c.SaltLen, uint32(16))
	configs.SetDefault(&c.KeyLen, uint32(32))
	configs.SetDefault(&c.Memory, uint32(32*1024))
}
