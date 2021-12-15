package standard

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

type config struct {
	DBUrl string `mapstructure:"db_url"`
}

func (adapter) Create(conf *configs.IDManager) plugins.IDManager {
	return &manager{rawConf: conf}
}
