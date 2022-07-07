package url

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	_ "embed"
	"github.com/mitchellh/mapstructure"
	"io"
	"net/http"

	"context"
	"errors"
)

//go:embed meta.yaml
var rawMeta []byte

var meta core.Meta

// init initializes package by register pluginCreator
func init() {
	meta = core.CryptoStorageRepo.Register(rawMeta, Create)
}

type storage struct {
	pluginApi core.PluginAPI
	rawConf   configs.PluginConfig
	conf      *config
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

func (s storage) GetMetaData() core.Meta {
	return meta
}

func (s *storage) GetAppRoutes() []*core.Route {
	return []*core.Route{}
}

func (*storage) Write(_ []byte) error {
	return errors.New("url key key storage: Write method is redundant and not allowed")
}

func (s *storage) Read(v *[]byte) (ok bool, err error) {
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, s.conf.Path, http.NoBody)
	if err != nil {
		return false, err
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	*v, err = io.ReadAll(resp.Body)
	if err != nil || len(*v) == 0 {
		return false, err
	}
	return true, nil
}
