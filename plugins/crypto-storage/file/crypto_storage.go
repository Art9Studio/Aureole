package file

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"os"

	"github.com/mitchellh/mapstructure"
)

const pluginID = "3827"

type cryptoStorage struct {
	pluginApi core.PluginAPI
	rawConf   *configs.CryptoStorage
	conf      *config
}

func (cs *cryptoStorage) Init(api core.PluginAPI) error {
	cs.pluginApi = api
	adapterConf := &config{}
	if err := mapstructure.Decode(cs.rawConf.Config, adapterConf); err != nil {
		return err
	}
	cs.conf = adapterConf
	return nil
}

func (cs *cryptoStorage) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: cs.rawConf.Name,
		ID:   pluginID,
	}
}

func (cs *cryptoStorage) Write(v []byte) error {
	return os.WriteFile(cs.conf.Path, v, 0o644)
}

func (cs *cryptoStorage) Read(v *[]byte) (ok bool, err error) {
	if _, err := os.Stat(cs.conf.Path); err != nil && os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	*v, err = os.ReadFile(cs.conf.Path)
	if err != nil {
		return false, err
	}
	return true, nil
}
