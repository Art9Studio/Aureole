package google

import "aureole/internal/configs"

func (c *config) setDefaults() {
	c.RedirectUri = "/login"
	configs.SetDefault(&c.Scopes, []string{"https://www.googleapis.com/auth/userinfo.email"})
}
