package jwt

import "aureole/internal/configs"

func (c *config) setDefaults() {
	configs.SetDefault(&c.Iss, "Aureole Server")
	configs.SetDefault(&c.Aud, []string{})
	configs.SetDefault(&c.Nbf, 0)
	configs.SetDefault(&c.Iat, true)
	configs.SetDefault(&c.Sub, true)
	configs.SetDefault(&c.AccessBearer, header)
	configs.SetDefault(&c.RefreshBearer, cookie)
	configs.SetDefault(&c.AccessExp, 900)
	configs.SetDefault(&c.RefreshExp, 7890000)
	configs.SetDefault(&c.VerifyKeys, []string{c.SignKey})
}
