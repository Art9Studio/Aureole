package email

import (
	"aureole/message"
	"context"
	"fmt"
)

// AdapterName is the internal name of the adapter
const AdapterName = "email"

// init initializes package by register adapter
func init() {
	message.RegisterAdapter(AdapterName, &emailAdapter{})
}

// emailAdapter represents adapter for email database
type emailAdapter struct {
}

// OpenConfig attempts to establish a connection with a db by connection config
func (pg emailAdapter) OpenWithConfig(connConf message.ConnConfig) (message.ConnSession, error) {
	sess := &ConnSession{
		ctx:      context.Background(),
		connConf: connConf,
	}

	if err := sess.Open(); err != nil {
		return nil, err
	}

	return sess, nil
}

// NewConfig creates new ConnConfig struct from the raw data, parsed from the config file
func (pg emailAdapter) NewConfig(data map[string]interface{}) (message.ConnConfig, error) {
	opts := make(map[string]string)
	if rawOpts, ok := data["options"].(map[string]interface{}); ok {
		for key, value := range rawOpts {
			opts[key] = fmt.Sprintf("%v", value)
		}
	}

	return ConnConfig{
		User:     data["username"].(string),
		Password: data["password"].(string),
		Host:     data["host"].(string),
		Port:     data["port"].(string),
		Database: data["db_name"].(string),
		Options:  opts,
	}, nil
}
