package url

import (
	"aureole/internal/configs"
	"errors"
	"github.com/mitchellh/mapstructure"
	"io"
	"net/http"
)

type Storage struct {
	rawConf *configs.Storage
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

func (s *Storage) Read() ([]byte, error) {
	request, err := http.NewRequest("GET", s.conf.Path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (s *Storage) Write(value []byte) error {
	return errors.New("url storage: write method not allowed")
}
