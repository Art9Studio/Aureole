package url

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/mitchellh/mapstructure"
)

const pluginID = "4896"

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

func (*cryptoStorage) Write(_ []byte) error {
	return errors.New("url key key storage: Write method is redundant and not allowed")
}

func (cs *cryptoStorage) Read(v *[]byte) (ok bool, err error) {
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, cs.conf.Path, http.NoBody)
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
