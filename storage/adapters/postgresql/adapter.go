package postgresql

import (
	"context"
	"fmt"
	"gouth/storage"
	"net/url"
	"strings"
)

// AdapterName is the internal name of the adapter
const AdapterName = "postgresql"

var AdapterFeatures = map[string]bool{"users": true, "sessions": true}

// init initializes package by register adapter
func init() {
	storage.RegisterAdapter(AdapterName, &pgAdapter{})
}

// pgAdapter represents adapter for postgresql database
type pgAdapter struct {
}

func (pg pgAdapter) GetFeatures() map[string]bool {
	return AdapterFeatures
}

// OpenConfig attempts to establish a connection with a db by connection config
func (pg pgAdapter) OpenWithConfig(connConf storage.ConnConfig) (storage.ConnSession, error) {
	sess := &ConnSession{
		ctx:      context.Background(),
		connConf: connConf,
	}

	if err := sess.Open(); err != nil {
		return nil, err
	}

	return sess, nil
}

// ParseUrl parses the connection url into ConnConfig struct
func (pg pgAdapter) ParseUrl(connUrl string) (storage.ConnConfig, error) {
	connConf := ConnConfig{}
	if !strings.HasPrefix(connUrl, connConf.AdapterName()+"://") {
		return nil, fmt.Errorf("expecting postgresql:// connection schema")
	}

	var (
		u   *url.URL
		err error
	)
	if u, err = url.Parse(connUrl); err != nil {
		return nil, err
	}

	var addr = strings.Split(u.Host, ":")
	if len(addr) < 2 {
		return nil, fmt.Errorf("invalid connection url")
	}

	_, isSetPasswd := u.User.Password()
	dbName := strings.Trim(u.Path, "/")
	if addr[0] == "" ||
		addr[1] == "" ||
		dbName == "" ||
		u.User.Username() == "" ||
		!isSetPasswd {
		return nil, fmt.Errorf("invalid connection url")
	}

	connConf.Host = addr[0]
	connConf.Port = addr[1]
	connConf.Database = dbName
	connConf.User = u.User.Username()
	connConf.Password, _ = u.User.Password()
	connConf.Options = map[string]string{}

	var vv url.Values
	if vv, err = url.ParseQuery(u.RawQuery); err != nil {
		return nil, err
	}

	for k := range vv {
		connConf.Options[k] = vv.Get(k)
	}

	return connConf, err
}

// NewConfig creates new ConnConfig struct from the raw data, parsed from the config file
func (pg pgAdapter) NewConfig(data map[string]interface{}) (storage.ConnConfig, error) {
	requiredKeys := []string{"username", "password", "host", "port", "db_name"}

	for _, key := range requiredKeys {
		if _, ok := data[key]; !ok {
			return nil, fmt.Errorf("connection config: missing %s statement", key)
		} else if data[key] == "" {
			return nil, fmt.Errorf("connection config: %s statement cannot be empty", key)
		}
	}

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
