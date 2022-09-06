package jwk

import (
	"aureole/configs"
)

type config struct {
	Kty             string `mapstructure:"kty" json:"kty"`
	Alg             string `mapstructure:"alg" json:"alg"`
	Use             string `mapstructure:"use" json:"use"`
	Curve           string `mapstructure:"curve" json:"curve"`
	Size            int    `mapstructure:"size" json:"size"`
	Kid             string `mapstructure:"kid" json:"kid"`
	RefreshInterval int    `mapstructure:"refresh_interval" json:"refresh_interval"`
	RetriesNum      int    `mapstructure:"retries_num" json:"retries_num"`
	RetryInterval   int    `mapstructure:"retry_interval" json:"retry_interval"`
	Storage         string `mapstructure:"storage" json:"storage"`
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
