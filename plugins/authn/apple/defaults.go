package apple

import "aureole/internal/configs"

func (c *config) setDefaults() {
	configs.SetDefault(&c.Scopes, []string{"email"})
}
