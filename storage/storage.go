package storage

import (
	"errors"
	"strings"
)

// Open attempts to establish a connection with a database
func Open(data RawConnData) (Session, error) {
	if connConf, ok := data["connection_config"].(map[string]interface{}); ok {
		adapterName, ok := connConf["adapter"].(string)
		if !ok {
			return nil, errors.New("invalid adapter name")
		}

		adapter, err := GetAdapter(adapterName)
		if err != nil {
			return nil, err
		}

		config, err := adapter.NewConfig(connConf)
		if err != nil {
			return nil, err
		}

		return adapter.OpenConfig(config)
	} else if connStr, ok := data["connection_url"].(string); ok && connStr != "" {
		adapterName := strings.Split(connStr, "://")[0]

		adapter, err := GetAdapter(adapterName)
		if err != nil {
			return nil, err
		}

		config, err := adapter.ParseUrl(connStr)
		if err != nil {
			return nil, err
		}

		return adapter.OpenConfig(config)

	}
	return nil, errors.New("missing connection data")
}
