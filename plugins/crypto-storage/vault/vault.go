package vault

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"encoding/json"

	vaultAPI "github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
)

const pluginID = "3521"

type storage struct {
	pluginApi core.PluginAPI
	rawConf   *configs.CryptoStorage
	conf      *config
	client    *vaultAPI.Client
}

func (s *storage) Init(api core.PluginAPI) error {
	s.pluginApi = api
	adapterConf := &config{}
	if err := mapstructure.Decode(s.rawConf.Config, adapterConf); err != nil {
		return err
	}
	s.conf = adapterConf

	client, err := vaultAPI.NewClient(&vaultAPI.Config{Address: s.conf.Address})
	if err != nil {
		return err
	}
	client.SetToken(s.conf.Token)
	s.client = client

	return nil
}

func (s *storage) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: s.rawConf.Name,
		ID:   pluginID,
	}
}

func (s *storage) Write(v []byte) error {
	_, err := s.client.Logical().WriteBytes(s.conf.Path, v)
	return err
}

func (s *storage) Read(v *[]byte) (ok bool, err error) {
	scr, err := s.client.Logical().Read(s.conf.Path)
	if err != nil {
		return false, err
	} else if scr == nil {
		return false, nil
	}

	*v, err = json.Marshal(scr.Data)
	if err != nil {
		return false, err
	} else if string(*v) == "null" {
		return false, nil
	}
	return true, nil
}
