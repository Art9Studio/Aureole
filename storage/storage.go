package storage

import (
	"errors"
	"strings"
)

// OpenConfig attempts to establish a connection with a database by ConnectionConfig
func Open(data RawConnectionData) (Session, error) {
	if connStr := data["connection_url"].(string); connStr != "" {
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

	} else if connConf := data["connection_config"].(map[string]interface{}); connConf != nil {
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
	}
	return nil, errors.New("missing connection data")

	//a, err := GetAdapter(connConf.AdapterName())
	//if err != nil {
	//	return nil, err
	//}
	//
	//return a.OpenConfig(connConf)
}
