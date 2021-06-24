package facebook

import "aureole/internal/configs"

func (c *config) SetDefaults() {
	configs.SetDefault(&c.Scopes, []string{"email"})
}
