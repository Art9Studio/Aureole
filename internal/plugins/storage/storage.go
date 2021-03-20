package storage

import (
	"aureole/configs"
	"aureole/internal/plugins/storage/types"
	"errors"
	"fmt"
	"strings"
)

// New returns desired Storage depends on the given config
func New(conf *configs.Storage) (types.Storage, error) {
	name, err := getAdapterName(conf)
	if err != nil {
		return nil, err
	}

	a, err := Repository.Get(name)
	if err != nil {
		return nil, err
	}

	adapter, ok := interface{}(a).(Adapter)
	if !ok {
		return nil, fmt.Errorf("trying to cast adapter was failed: %v", err)
	}

	return adapter.Create(conf)
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
