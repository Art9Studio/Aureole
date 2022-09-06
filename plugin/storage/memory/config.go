package memory

import (
	"aureole/configs"
)

type config struct {
	Size int `mapstructure:"size" json:"size"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.Size, 128)
}
