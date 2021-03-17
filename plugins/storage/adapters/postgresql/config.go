package postgresql

import (
	"aureole/configs"
	"aureole/plugins/storage/types"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"net/url"
	"strings"
)

// Conf represents a parsed PostgreSQL connection URL
type Conf struct {
	Url      string            `mapstructure:"url"`
	User     string            `mapstructure:"username"`
	Password string            `mapstructure:"password"`
	Host     string            `mapstructure:"host"`
	Port     string            `mapstructure:"port"`
	Database string            `mapstructure:"db_name"`
	Options  map[string]string `mapstructure:"options"`
}

func (pg pgAdapter) Get(conf *configs.Storage) (types.Storage, error) {
	adapterConfMap := conf.Config
	adapterConf := &Conf{}

	specificConf := conf.Config
	if specificConf["adapter"] != nil {
		err := mapstructure.Decode(adapterConfMap, adapterConf)
		if err != nil {
			return nil, err
		}
	} else if connStr, ok := specificConf["url"].(string); ok && connStr != "" {
		err := ParseUrl(specificConf, adapterConf)
		if err != nil {
			return nil, err
		}
	}

	return initAdapter(conf, adapterConf)
}

func initAdapter(conf *configs.Storage, adapterConf *Conf) (*Storage, error) {
	return &Storage{
		Conf: adapterConf,
	}, nil
}

//// String reassembles PostgreSQL connection config into a valid connection url
func (conf Conf) ToURL() (string, error) {
	vv := url.Values{}
	if conf.Options != nil {
		for k, v := range conf.Options {
			vv.Set(k, v)
		}
	}

	if conf.User == "" ||
		conf.Password == "" ||
		conf.Host == "" ||
		conf.Port == "" ||
		conf.Database == "" {
		return "", fmt.Errorf("invalid connection url")
	}

	u := url.URL{
		Scheme:     AdapterName,
		User:       url.UserPassword(conf.User, conf.Password),
		Host:       fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		Path:       conf.Database,
		ForceQuery: false,
		RawQuery:   vv.Encode(),
	}
	return u.String(), nil
}

// ParseUrl parses the connection url into ConnConfig struct
func ParseUrl(rawConf configs.RawConfig, conf *Conf) error {
	connUrl := rawConf["url"].(string)
	if !strings.HasPrefix(connUrl, AdapterName+"://") {
		return fmt.Errorf("expecting postgresql:// connection schema")
	}

	var (
		u   *url.URL
		err error
	)
	if u, err = url.Parse(connUrl); err != nil {
		return err
	}

	var addr = strings.Split(u.Host, ":")
	if len(addr) < 2 {
		return fmt.Errorf("invalid connection url")
	}

	_, isSetPasswd := u.User.Password()
	dbName := strings.Trim(u.Path, "/")
	if addr[0] == "" ||
		addr[1] == "" ||
		dbName == "" ||
		u.User.Username() == "" ||
		!isSetPasswd {
		return fmt.Errorf("invalid connection url")
	}

	conf.Host = addr[0]
	conf.Port = addr[1]
	conf.Database = dbName
	conf.User = u.User.Username()
	conf.Password, _ = u.User.Password()
	conf.Options = map[string]string{}

	var vv url.Values
	if vv, err = url.ParseQuery(u.RawQuery); err != nil {
		return err
	}

	for k := range vv {
		conf.Options[k] = vv.Get(k)
	}

	return nil
}
