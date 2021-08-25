package urls

import "aureole/internal/configs"

func (c *config) setDefaults() {
	configs.SetDefault(&c.Path, "urls")
}
