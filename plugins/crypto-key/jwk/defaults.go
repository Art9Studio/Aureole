package jwk

import (
	"aureole/internal/configs"
)

func (c *config) setDefaults() {
	if c.Use == "enc" {
		configs.SetDefault(&c.Kty, "RSA")
		configs.SetDefault(&c.Alg, "RSA-OAEP-256")
		configs.SetDefault(&c.Size, 4096)
	} else {
		configs.SetDefault(&c.Kty, "EC")
		configs.SetDefault(&c.Alg, "ES256")
		configs.SetDefault(&c.Use, "sig")
		configs.SetDefault(&c.Curve, "P-256")
	}
	configs.SetDefault(&c.Kid, "SHA-256")
	configs.SetDefault(&c.RetriesNum, 1)
	configs.SetDefault(&c.RetryInterval, 100)
}
