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

type cryptoStorage struct {
	pluginApi core.PluginAPI
	rawConf   *configs.CryptoStorage
	conf      *config
	client    *vaultAPI.Client
}

func (cs *cryptoStorage) Init(api core.PluginAPI) error {
	cs.pluginApi = api
	adapterConf := &config{}
	if err := mapstructure.Decode(cs.rawConf.Config, adapterConf); err != nil {
		return err
	}
	cs.conf = adapterConf

	client, err := vaultAPI.NewClient(&vaultAPI.Config{Address: cs.conf.Address})
	if err != nil {
		return err
	}
	client.SetToken(cs.conf.Token)
	cs.client = client

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
	_, err := cs.client.Logical().WriteBytes(cs.conf.Path, v)
	return err
}

func (cs *cryptoStorage) Read(v *[]byte) (ok bool, err error) {
	scr, err := cs.client.Logical().Read(cs.conf.Path)
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
