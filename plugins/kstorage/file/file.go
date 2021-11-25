package file

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/core"
	"github.com/mitchellh/mapstructure"
	"os"
)

const PluginID = "3827"

type storage struct {
	pluginApi core.PluginAPI
	rawConf   *configs.KeyStorage
	conf      *config
}

func (s *storage) Init(api core.PluginAPI) error {
	s.pluginApi = api
	adapterConf := &config{}
	if err := mapstructure.Decode(s.rawConf.Config, adapterConf); err != nil {
		return err
	}
	s.conf = adapterConf
	return nil
}

func (*storage) GetPluginID() string {
	return PluginID
}

func (s *storage) Write(v []byte) error {
	return os.WriteFile(s.conf.Path, v, 0o644)
}

func (s *storage) Read(v *[]byte) (ok bool, err error) {
	if _, err := os.Stat(s.conf.Path); err != nil && os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	*v, err = os.ReadFile(s.conf.Path)
	if err != nil {
		return false, err
	}
	return true, nil
}
