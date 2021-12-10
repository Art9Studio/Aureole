package google

import "aureole/internal/configs"

func (c *config) setDefaults() {
	c.PathPrefix = "/oauth2/" + AdapterName
	c.RedirectUri = "/login"
	configs.SetDefault(&c.Scopes, []string{"https://www.googleapis.com/auth/userinfo.email"})
}
