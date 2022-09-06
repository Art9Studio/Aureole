package pbkdf2

import (
	"aureole/configs"
)

func (c *config) setDefaults() {
	configs.SetDefault(&c.Iterations, 260000)
	configs.SetDefault(&c.SaltLen, 22)
	configs.SetDefault(&c.KeyLen, 32)
	configs.SetDefault(&c.FuncName, "sha256")
}
