package file

import (
	"aureole/internal/configs"
	"github.com/mitchellh/mapstructure"
	"os"
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

func (s *Storage) Write(v []byte) error {
	return os.WriteFile(s.conf.Path, v, 0o644)
}

func (s *Storage) Read(v *[]byte) (ok bool, err error) {
	if _, err := os.Stat(s.conf.Path); err != nil && !os.IsNotExist(err) {
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
