package storage

import (
	"errors"
	"strings"
)

// Open attempts to establish a connection with a database
func Open(data RawConnConfig, features []string) (ConnSession, error) {
	if connConf, ok := data["connection_config"].(map[string]interface{}); ok {
		adapterName, ok := connConf["adapter"].(string)
		if !ok {
			return nil, errors.New("invalid adapter Name")
		}

		adapter, err := GetAdapter(adapterName, features)
		if err != nil {
			return nil, err
		}

		config, err := adapter.NewConfig(connConf)
		if err != nil {
			return nil, err
		}

		return adapter.OpenWithConfig(config)
	} else if connStr, ok := data["connection_url"].(string); ok && connStr != "" {
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
