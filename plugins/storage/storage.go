package storage

import (
	"aureole/configs"
	"errors"
	"strings"
)

// Open attempts to establish a connection with a database
func Open(confMap configs.RawConfig) (ConnSession, error) {
	if confMap["adapter"] != nil {
		adapterName, ok := confMap["adapter"].(string)
		if !ok {
			return nil, errors.New("invalid adapter Name")
		}

		adapter, err := GetAdapter(adapterName)
		if err != nil {
			return nil, err
		}

		config, err := adapter.NewConfig(confMap)
		if err != nil {
			return nil, err
		}

		return adapter.OpenWithConfig(config)
	} else if connStr, ok := confMap["url"].(string); ok && connStr != "" {
		adapterName := strings.Split(connStr, "://")[0]

		adapter, err := GetAdapter(adapterName)
		if err != nil {
			return nil, err
		}

		config, err := adapter.ParseUrl(connStr)
		if err != nil {
			return nil, err
		}

		return adapter.OpenWithConfig(config)

	}
	return nil, errors.New("missing connection confMap")
}
