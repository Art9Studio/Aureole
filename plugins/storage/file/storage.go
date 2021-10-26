package file

import (
	"aureole/internal/configs"
	"github.com/mitchellh/mapstructure"
	"os"
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
	return os.ReadFile(s.conf.Path)
}

func (s *Storage) Write(value []byte) error {
	return os.WriteFile(s.conf.Path, value, 0644)
}
