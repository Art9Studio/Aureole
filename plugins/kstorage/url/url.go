package url

import (
	"aureole/internal/configs"
	"context"
	"errors"
	"github.com/mitchellh/mapstructure"
	"io"
	"net/http"
)

type Storage struct {
	rawConf *configs.KeyStorage
	conf    *config
}

func (s *Storage) Init() error {
	adapterConf := &config{}
	if err := mapstructure.Decode(s.rawConf.Config, adapterConf); err != nil {
		return err
	}
	s.conf = adapterConf
	return nil
}

func (*Storage) Write(_ []byte) error {
	return errors.New("url key key storage: Write method is redundant and not allowed")
}

func (s *Storage) Read(v *[]byte) (ok bool, err error) {
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
