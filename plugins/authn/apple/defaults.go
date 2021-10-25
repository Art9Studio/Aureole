package apple

import "aureole/internal/configs"

func (c *config) setDefaults() {
	c.RedirectUri = "/login"
	configs.SetDefault(&c.Scopes, []string{"email"})
}
