package file

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	_ "embed"
	"os"

	"github.com/mitchellh/mapstructure"
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
}

func (s *storage) GetCustomAppRoutes() []*core.Route {
	return []*core.Route{}
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
	return nil
}

func (s storage) GetMetadata() core.Metadata {
	return meta
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
