package jwk

import (
	"aureole/internal/configs"
)

type config struct {
	Kty             string `mapstructure:"kty"`
	Alg             string `mapstructure:"alg"`
	Use             string `mapstructure:"use"`
	Curve           string `mapstructure:"curve"`
	Size            int    `mapstructure:"size"`
	Kid             string `mapstructure:"kid"`
	RefreshInterval int    `mapstructure:"refresh_interval"`
	RetriesNum      int    `mapstructure:"retries_num"`
	RetryInterval   int    `mapstructure:"retry_interval"`
	Storage         string `mapstructure:"storage"`
	PathPrefix      string
}

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
