package storage

import (
	"aureole/configs"
	"errors"
	"strings"
)

// New returns desired ConnSession depends on the given config
func New(conf *configs.Storage) (ConnSession, error) {
	name, err := getAdapterName(conf)

	adapter, err := GetAdapter(name)
	if err != nil {
		return nil, err
	}

	return adapter.Get(conf)
}

func getAdapterName(conf *configs.Storage) (string, error) {
	if conf.Type != "" {
		return conf.Type, nil
	} else {
		specificConf := conf.Config
		if specificConf["adapter"] != nil {
			adapterName, ok := specificConf["adapter"].(string)
			if !ok {
				return "", errors.New("invalid adapter name")
			}

			return adapterName, nil
		} else if connStr, ok := specificConf["url"].(string); ok && connStr != "" {
			adapterName := strings.Split(connStr, "://")[0]

			return adapterName, nil
		}
	}

	return "", errors.New("invalid connection config")
}
