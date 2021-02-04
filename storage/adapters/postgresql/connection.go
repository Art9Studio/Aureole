package postgresql

import (
	"fmt"
	"gouth/config"
	"net/url"
	"strings"
)

type ConnectionString struct {
	url string
}

// ConnectionConfig represents a parsed PostgreSQL connection URL
type ConnectionConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
	Options  map[string]string
}

// String reassembles PostgreSQL connection config into a valid connection url
func (connStr ConnectionString) String() string {
	return connStr.url
}

// String reassembles PostgreSQL connection config into a valid connection url
func (connStr ConnectionString) AdapterName() string {
	return strings.Split(connStr.url, "://")[0]
}

func (appConf config.ConnectionConfig) ToPostgresql() ConnectionConfig {
	return ConnectionConfig{
		User:     "",
		Password: "",
		Host:     "",
		Port:     "",
		Database: "",
		Options:  nil,
	}
}

// String reassembles PostgreSQL connection config into a valid connection url
func (connConf ConnectionConfig) String() (string, error) {
	vv := url.Values{}
	if connConf.Options != nil {
		for k, v := range connConf.Options {
			vv.Set(k, v)
		}
	}

	if connConf.AdapterName() == "" ||
		connConf.User == "" ||
		connConf.Password == "" ||
		connConf.Host == "" ||
		connConf.Port == "" ||
		connConf.Database == "" {
		return "", fmt.Errorf("invalid connection url")
	}

	u := url.URL{
		Scheme:     connConf.AdapterName(),
		User:       url.UserPassword(connConf.User, connConf.Password),
		Host:       fmt.Sprintf("%s:%s", connConf.Host, connConf.Port),
		Path:       connConf.Database,
		ForceQuery: false,
		RawQuery:   vv.Encode(),
	}
	return u.String(), nil
}

// DBName returns the name of the database, that we've connected by this config
func (connConf ConnectionConfig) DBName() string {
	return connConf.Database
}

// AdapterName return the adapter name, that was used to set up connection
func (connConf ConnectionConfig) AdapterName() string {
	return AdapterName
}

// ParseUrl parses the connection url into ConnectionConfig struct
func ParseUrl(connUrl string) (connConf ConnectionConfig, err error) {
	if !strings.HasPrefix(connUrl, connConf.AdapterName()+"://") {
		return connConf, fmt.Errorf("expecting postgresql:// connection schema")
	}

	var u *url.URL
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
