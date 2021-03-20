package pbkdf2

import "aureole/configs"

// TODO: figure out best default settings
func (c *config) setDefaults() {
	configs.SetDefault(&c.Iterations, 4096)
	configs.SetDefault(&c.SaltLen, 16)
	configs.SetDefault(&c.KeyLen, 32)
	configs.SetDefault(&c.FuncName, "sha1")
}
