package memory

import "aureole/internal/configs"

func (c *config) setDefaults() {
	configs.SetDefault(&c.Size, 128*1024*1024)
}
