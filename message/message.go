package message

import (
	"errors"
)

// Open attempts to establish a connection with a database
func Open(data RawStorageConfig) (ConnSession, error) {
	if connConf, ok := data["config"].(map[string]interface{}); ok {
		adapterName, ok := connConf["adapter"].(string)
		if !ok {
			return nil, errors.New("invalid adapter Name")
		}

		adapter, err := GetAdapter(adapterName)
		if err != nil {
			return nil, err
		}

		config, err := adapter.NewConfig(connConf)
		if err != nil {
			return nil, err
		}

		return adapter.OpenWithConfig(config)
	}
	return nil, errors.New("missing connection data")
}
