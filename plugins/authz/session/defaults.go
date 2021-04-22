package session

import "aureole/internal/configs"

func (c *config) setDefaults() {
	// todo: think about defaults parameters
	configs.SetDefault(&c.Path, "/")
	configs.SetDefault(&c.MaxAge, 3600)
}
