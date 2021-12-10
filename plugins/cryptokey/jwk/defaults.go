package jwk

import (
	"aureole/internal/configs"
)

func (c *config) setDefaults() {
	configs.SetDefault(&c.Kty, "EC")
	configs.SetDefault(&c.Alg, "ES256")
	configs.SetDefault(&c.Use, "sig")
	configs.SetDefault(&c.Curve, "P-256")
	configs.SetDefault(&c.Kid, "SHA-256")
}
