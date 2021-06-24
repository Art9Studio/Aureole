package google

import "aureole/internal/configs"

func (c *config) SetDefaults() {
	configs.SetDefault(&c.Scopes, []string{"https://www.googleapis.com/auth/userinfo.email"})
}
