package url

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/core"
	"context"
	"errors"
	"github.com/mitchellh/mapstructure"
	"io"
	"net/http"
)

const PluginID = "4896"

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

func (*storage) Write(_ []byte) error {
	return errors.New("url key key storage: Write method is redundant and not allowed")
}

func (s *storage) Read(v *[]byte) (ok bool, err error) {
	request, err := http.NewRequestWithContext(context.Background(), "GET", s.conf.Path, http.NoBody)
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
