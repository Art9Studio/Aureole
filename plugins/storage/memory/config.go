package memory

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

type config struct {
	Size int `mapstructure:"size"`
}

func (memoryAdapter) Create(conf *configs.Storage) plugins.Storage {
	return &memory{rawConf: conf}
}
