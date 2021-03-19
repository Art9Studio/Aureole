package postgresql

import (
	"aureole/configs"
	"aureole/plugins/storage/types"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"net/url"
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

func (pg pgAdapter) Create(conf *configs.Storage) (types.Storage, error) {
	adapterConfMap := conf.Config
	adapterConf := &Conf{}

	err := mapstructure.Decode(adapterConfMap, adapterConf)
	if err != nil {
		return nil, err
	}

	return initAdapter(adapterConf)
}

func initAdapter(adapterConf *Conf) (*Storage, error) {
	a := &Storage{
		Conf: adapterConf,
	}

	err := a.Open()
	if err != nil {
		return nil, err
	}
	return a, nil
}

// ToURL reassembles PostgreSQL connection config into a valid connection url
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
