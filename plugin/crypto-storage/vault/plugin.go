package vault

import (
	"aureole/configs"
	"aureole/internal/core"
	_ "embed"

	"github.com/mitchellh/mapstructure"

	"encoding/json"

	vaultAPI "github.com/hashicorp/vault/api"
)

//go:embed meta.yaml
var rawMeta []byte

var meta core.Metadata

// init initializes package by register pluginCreator
func init() {
	meta = core.CryptoStorageRepo.Register(rawMeta, Create)
}

type storage struct {
	pluginApi core.PluginAPI
	rawConf   configs.PluginConfig
	conf      *config
	client    *vaultAPI.Client
}

func Create(conf configs.PluginConfig) core.CryptoStorage {
	return &storage{rawConf: conf}
}

func (s *storage) Init(api core.PluginAPI) error {
	s.pluginApi = api
	PluginConf := &config{}
	if err := mapstructure.Decode(s.rawConf.Config, PluginConf); err != nil {
		return err
	}
	s.conf = PluginConf

	client, err := vaultAPI.NewClient(&vaultAPI.Config{Address: s.conf.Address})
	if err != nil {
		return err
	}
	client.SetToken(s.conf.Token)
	s.client = client

	return nil
}

func (s storage) GetMetadata() core.Metadata {
	return meta
}

func (s *storage) GetCustomAppRoutes() []*core.Route {
	return []*core.Route{}
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
