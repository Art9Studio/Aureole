package postgresql

import (
	"context"
	"errors"
	"fmt"
	adapters "gouth/storage"
	"net/url"
	"strings"
)

// AdapterName is the internal name of the adapter
const AdapterName = "postgresql"

// init initializes package by register adapter
func init() {
	adapters.RegisterAdapter(AdapterName, &pgAdapter{})
}

// pgAdapter represents adapter for postgresql database
type pgAdapter struct {
}

// OpenConfig attempts to establish a connection with a db by connection config
func (pg pgAdapter) OpenConfig(connConf adapters.ConnectionConfig) (adapters.Session, error) {
	sess := &Session{
		ctx:      context.Background(),
		connConf: connConf,
	}

	if err := sess.Open(); err != nil {
		return nil, err
	}

	return sess, nil
}

// ParseUrl parses the connection url into ConnectionConfig struct
func (pg pgAdapter) ParseUrl(connUrl string) (adapters.ConnectionConfig, error) {
	connConf := ConnectionConfig{}
	if !strings.HasPrefix(connUrl, connConf.AdapterName()+"://") {
		return connConf, fmt.Errorf("expecting postgresql:// connection schema")
	}

	var (
		u   *url.URL
		err error
	)
	if u, err = url.Parse(connUrl); err != nil {
		return connConf, err
	}

	var addr = strings.Split(u.Host, ":")
	if len(addr) < 2 {
		return ConnectionConfig{}, fmt.Errorf("invalid connection url")
	}

	_, isSetPasswd := u.User.Password()
	dbName := strings.Trim(u.Path, "/")
	if addr[0] == "" ||
		addr[1] == "" ||
		dbName == "" ||
		u.User.Username() == "" ||
		!isSetPasswd {
		return connConf, fmt.Errorf("invalid connection url")
	}

	connConf.Host = addr[0]
	connConf.Port = addr[1]
	connConf.Database = dbName
	connConf.User = u.User.Username()
	connConf.Password, _ = u.User.Password()
	connConf.Options = map[string]string{}

	var vv url.Values

	if vv, err = url.ParseQuery(u.RawQuery); err != nil {
		return connConf, err
	}

	for k := range vv {
		connConf.Options[k] = vv.Get(k)
	}

	return connConf, err
}

func (pg pgAdapter) NewConfig(data map[string]interface{}) (adapters.ConnectionConfig, error) {
	User, ok := data["user"].(string)
	if !ok {
		return ConnectionConfig{}, errors.New("Smth")
	}

	return ConnectionConfig{
		User:     User,
		Password: "",
		Host:     "",
		Port:     "",
		Database: "",
		Options:  nil,
	}, nil
}
