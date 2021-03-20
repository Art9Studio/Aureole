package argon2

import "aureole/configs"

// TODO: figure out best default settings
func (c *Conf) setDefaults() {
	configs.SetDefault(&c.Kind, "argon2i")
	configs.SetDefault(&c.Iterations, 3)
	configs.SetDefault(&c.Parallelism, 2)
	configs.SetDefault(&c.SaltLen, 16)
	configs.SetDefault(&c.KeyLen, 32)
	configs.SetDefault(&c.Memory, 32*1024)
}
