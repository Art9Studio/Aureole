package storage

import (
	"errors"
	"strings"
)

// Open attempts to establish a connection with a database
func Open(data RawStorageConfig, features []string) (ConnSession, error) {
	if data["adapter"] != nil {
		adapterName, ok := data["adapter"].(string)
		if !ok {
			return nil, errors.New("invalid adapter Name")
		}

		adapter, err := GetAdapter(adapterName, features)
		if err != nil {
			return nil, err
		}

		config, err := adapter.NewConfig(data)
		if err != nil {
			return nil, err
		}

		return adapter.OpenWithConfig(config)
	} else if connStr, ok := data["url"].(string); ok && connStr != "" {
		adapterName := strings.Split(connStr, "://")[0]

		adapter, err := GetAdapter(adapterName, features)
		if err != nil {
			return nil, err
		}

		config, err := adapter.ParseUrl(connStr)
		if err != nil {
			return nil, err
		}

		return adapter.OpenWithConfig(config)

	}
	return nil, errors.New("missing connection data")
}
