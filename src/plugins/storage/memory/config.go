package memory

import (
	"aureole/internal/configs"
)

type config struct {
	Size int `mapstructure:"size" json:"size"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.Size, 128)
}
